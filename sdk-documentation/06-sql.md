# SQL

This page documents SQL execution via the Go SDK: request/response shapes, parameter binding, typed mapping, idempotency for writes, and guidance for large result sets.

---

## Capabilities

* Execute parameterized SQL statements (DDL/DML/SELECT).
* Receive dual-shape responses:

  * DDL/DML → `{ ok, row_count }`
  * SELECT → `{ ok, rows }`
* Map rows into typed structs via `Query[T]`.
* Supply idempotency keys for write operations.
* Switch to streaming or pagination when handling large result sets (see `07-streaming-pagination.md`).

---

## Method ↔ Endpoint mapping

| Concern     | SDK method                       | HTTP method & path                          | Headers (auto-applied when set)                                                                          |
| ----------- | -------------------------------- | ------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| Execute SQL | `Client.ExecSQL` / `Project.SQL` | `POST /warlotSql/projects/{project_id}/sql` | `Content-Type`, `User-Agent`, `x-api-key`, `x-holder-id`, `x-project-name`, optional `x-idempotency-key` |

---

## Request shape

```go
type SQLRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}
```

* **Placeholders:** positional `?` markers in the SQL string.
* **Binding:** `Params` elements are matched by position.
* **Types:** standard JSON types are accepted; the service performs SQLite-compatible conversions.

Example:

```go
proj := client.Project(projectID)

// DDL
_, _ = proj.SQL(ctx, `
  CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    price REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
  )`, nil)

// DML with parameters (idempotent write)
_, _ = proj.SQL(ctx,
  `INSERT INTO products (name, price) VALUES (?, ?)`,
  []any{"Laptop", 999.99},
  warlot.WithIdempotencyKey("insert-products-2025-10-03-01"),
)
```

---

## Response shapes

```go
type SQLResponse struct {
	OK       bool                     `json:"ok"`
	RowCount *int                     `json:"row_count,omitempty"` // DDL/DML
	Rows     []map[string]interface{} `json:"rows,omitempty"`      // SELECT
	Error    string                   `json:"error,omitempty"`
}
```

| Statement class | Fields populated          | Notes                                            |
| --------------- | ------------------------- | ------------------------------------------------ |
| DDL / DML       | `OK=true`, `RowCount` set | `Rows` omitted                                   |
| SELECT          | `OK=true`, `Rows` set     | `RowCount` omitted                               |
| Error           | `OK=false`, `Error` set   | SDK returns `error` and preserves parsed message |

**Flow (execution result):**

```mermaid
%%{init: {'theme': 'base'}}%%
flowchart TD
  A[ExecSQL] --> B{Statement type}
  B -->|DDL/DML| C[Return ok + row_count]
  B -->|SELECT| D[Return ok + rows[]]
  B -->|Error| E[Return error (APIError or SQL error)]
```

---

## Reading SELECT results

### Untyped (maps)

```go
res, err := proj.SQL(ctx, `SELECT id, name, price FROM products ORDER BY id`, nil)
if err != nil { /* handle */ }
for _, row := range res.Rows {
	// Access by column name; values are JSON-decoded (float64 for REAL, etc.).
	id, _ := row["id"].(float64)    // note: JSON numbers decode to float64
	name, _ := row["name"].(string)
	price, _ := row["price"].(float64)
	_ = id; _ = name; _ = price
}
```

### Typed (struct mapping)

`Query[T]` performs a JSON round-trip for each row.

```go
type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

items, err := warlot.Query[Product](ctx, proj,
	`SELECT id, name, price FROM products ORDER BY id`,
	nil,
)
if err != nil { /* handle */ }
for _, p := range items {
	_ = p // p is a Product with typed fields
}
```

> Mapping relies on JSON field names (snake_case column names match lowercased struct tags). Custom tags can be applied for explicit control.

---

## Updates and deletes (examples)

```go
// Update
_, _ = proj.SQL(ctx,
  `UPDATE products SET price = ? WHERE name = ?`,
  []any{899.99, "Laptop"},
  warlot.WithIdempotencyKey("update-products-2025-10-03-01"),
)

// Delete
_, _ = proj.SQL(ctx,
  `DELETE FROM products WHERE id = ?`,
  []any{1},
  warlot.WithIdempotencyKey("delete-products-1"),
)
```

---

## Idempotency for writes

* Header: `x-idempotency-key` (set via `WithIdempotencyKey`).
* Purpose: prevent duplicate effects when retries occur under 429/5xx conditions.
* Scope: recommended for INSERT/UPDATE/DELETE and DDL operations.

```go
_, err := proj.SQL(ctx,
  `INSERT INTO audit_log (event, payload) VALUES (?, ?)`,
  []any{"create", `{"id":123}`},
  warlot.WithIdempotencyKey("audit-evt-123"),
)
```

---

## Large result sets

For large SELECT outputs, consider:

* **Streaming** via `ExecSQLStream` + `RowScanner` to decode row-by-row.
* **Pagination** via `Pager` when rows are retrievable through the browse endpoint.

Detailed guidance: `07-streaming-pagination.md`.

---

## Errors

`APIError` is returned for non-2xx responses, with parsed fields when available.

```go
type APIError struct {
	StatusCode int
	Body       string
	Message    string
	Code       string
	Details    any
}
```

Detection example:

```go
if err != nil {
	if e, ok := err.(*warlot.APIError); ok {
		switch {
		case e.StatusCode == 429:
			// rate limited; retry/backoff
		case e.StatusCode >= 500:
			// transient server error
		default:
			// client-side or authorization issue
		}
	}
}
```

---

## Types (definition)

```go
// Entry points
func (c *Client) ExecSQL(ctx context.Context, projectID string, req SQLRequest, opts ...CallOption) (*SQLResponse, error)
func (p Project) SQL(ctx context.Context, sql string, params []any, opts ...CallOption) (*SQLResponse, error)

// Request/Response
type SQLRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}
type SQLResponse struct {
	OK       bool                     `json:"ok"`
	RowCount *int                     `json:"row_count,omitempty"`
	Rows     []map[string]interface{} `json:"rows,omitempty"`
	Error    string                   `json:"error,omitempty"`
}

// Typed mapping helper
func Query[T any](ctx context.Context, p Project, sql string, params []any, opts ...CallOption) ([]T, error)

// Per-call options (subset)
type CallOption func(*callOptions)
func WithIdempotencyKey(k string) CallOption
```

---

## Unit test definitions

### 1) DDL and DML return `row_count`

```go
// sql_rowcount_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_SQL_DDL_DML_RowCount(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"row_count":1}`))
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL))
	res, err := cl.ExecSQL(context.Background(), "P", SQLRequest{SQL: "CREATE TABLE t(x)", Params: nil})
	if err != nil || res.RowCount == nil || *res.RowCount != 1 {
		t.Fatalf("unexpected: res=%+v err=%v", res, err)
	}
}
```

### 2) SELECT returns rows and maps to typed struct

```go
// sql_select_typed_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_SQL_Select_Typed(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"rows":[{"id":1,"name":"X","price":1.23}]}`))
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL))
	p := cl.Project("P")

	type Row struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	rows, err := Query[Row](context.Background(), p, `SELECT * FROM t`, nil)
	if err != nil || len(rows) != 1 || rows[0].Name != "X" {
		t.Fatalf("unexpected: rows=%+v err=%v", rows, err)
	}
}
```

### 3) Idempotency key is forwarded

```go
// sql_idempotency_header_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_SQL_IdempotencyHeader(t *testing.T) {
	seen := ""
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("x-idempotency-key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"row_count":1}`))
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL))
	_, err := cl.ExecSQL(context.Background(), "P",
		SQLRequest{SQL: "INSERT INTO t(x) VALUES (?)", Params: []any{1}},
		WithIdempotencyKey("idemp-1"),
	)
	if err != nil || seen != "idemp-1" {
		t.Fatalf("unexpected: idempotency=%q err=%v", seen, err)
	}
}
```

---

## Troubleshooting (SQL)

| Symptom                                                   | Likely cause                                 | Action                                                                                      |
| --------------------------------------------------------- | -------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `json: cannot unmarshal string into Go value of type int` | Upstream field type differs from expectation | Use untyped map path or struct fields of compatible types; add custom decoding if necessary |
| `401/403` on ExecSQL                                      | Missing or mismatched auth headers           | Attach issued API key and confirm holder/project name                                       |
| `429` with intermittent success                           | Rate limiting                                | Provide idempotency keys for writes; allow retry/backoff                                    |
| Large memory footprint on big SELECT                      | Entire row set loaded in memory              | Use `ExecSQLStream` or pagination (`07-streaming-pagination.md`)                            |

---

## Related topics

* Authentication and header details: `03-authentication.md`
* Streaming and pagination for large results: `07-streaming-pagination.md`
* Errors and retry semantics: `09-errors.md`, `10-retries-rate-limits.md`
