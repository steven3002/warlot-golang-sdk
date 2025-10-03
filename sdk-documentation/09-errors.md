# Errors

A consistent error model simplifies operational handling across transport failures, non-success HTTP statuses, and SQL runtime issues. This page documents error surfaces, the SDK’s `APIError` type, retry boundaries, recommended handling patterns, and compact unit-test definitions.

---

## Error surfaces

| Layer                               | Example symptom                                    | Returned as                                      | Notes                                                                     |
| ----------------------------------- | -------------------------------------------------- | ------------------------------------------------ | ------------------------------------------------------------------------- |
| **Transport / context**             | DNS/TLS/connect timeout, context deadline exceeded | `error` (e.g., `context.DeadlineExceeded`)       | No HTTP response observed. Retries may occur depending on failure timing. |
| **HTTP non-2xx**                    | 401/403/404/429/5xx                                | `*warlot.APIError`                               | Body is parsed for `message`, `error`, `code`, `details`.                 |
| **Successful HTTP, JSON decode**    | Type mismatch, malformed JSON                      | `error` (wrap includes `decode response:`)       | Includes original body (truncated only by logs), no retry.                |
| **SQL runtime (200 with ok=false)** | `{ "ok": false, "error": "…" }`                    | `error` (SDK returns `error` containing message) | Occurs when service reports statement-level failure.                      |
| **Streaming read**                  | Premature close, malformed array                   | `RowScanner.Err()`                               | After `Next` returns `false`, check `Err()` to distinguish EOF vs. error. |

---

## Unified error type for non-2xx

```go
type APIError struct {
    StatusCode int
    Body       string
    Message    string
    Code       string      // optional code provided by gateway
    Details    interface{} // optional details payload
}

func (e *APIError) Error() string
```

**Decoding behavior**

* The SDK attempts to unmarshal the response body into `{ message, error, code, details }`.
* `Message` is populated from `message` (preferred) or `error` fields when present.
* `Body` always retains the raw text, aiding diagnostics.

---

## Retry boundaries

* **Automatic retries** occur for **429** and **5xx** responses.
* `Retry-After` (seconds or HTTP-date) is honored; otherwise, a jittered exponential backoff is applied within configured bounds.
* Non-retriable statuses (for example, 400/401/403/404) are returned immediately.
* Transport-level retries may occur depending on error timing and idempotent request safety.

> For write statements (INSERT/UPDATE/DELETE/DDL), idempotency keys should be set to prevent duplicate effects during retries.

---

## Status mapping and recommended action

| HTTP status               | Meaning                       | SDK behavior                              | Recommended action                                                     |
| ------------------------- | ----------------------------- | ----------------------------------------- | ---------------------------------------------------------------------- |
| **400 Bad Request**       | Invalid SQL or parameters     | `APIError` (no retry)                     | Validate SQL/params; correct request.                                  |
| **401 Unauthorized**      | Missing/invalid `x-api-key`   | `APIError` (no retry)                     | Issue or refresh API key; attach to client.                            |
| **403 Forbidden**         | Project/holder scope mismatch | `APIError` (no retry)                     | Confirm `holder_id`/`project_name` and key scope.                      |
| **404 Not Found**         | Resource/route absent         | `APIError` (no retry)                     | Confirm endpoint and project ID; consider resolve/init flow.           |
| **409 Conflict**          | State conflict                | `APIError` (no retry)                     | Reconcile state; for writes, prefer idempotency keys.                  |
| **429 Too Many Requests** | Rate limit                    | Retry with backoff (honors `Retry-After`) | Allow SDK to retry; consider idempotency on writes and tuning backoff. |
| **5xx Server Error**      | Transient backend error       | Retry with backoff                        | Allow SDK to retry; escalate if persistent.                            |

---

## Handling patterns

### Distinguish API errors from other errors

```go
res, err := proj.SQL(ctx, `SELECT id FROM t`, nil)
if err != nil {
    if e, ok := err.(*warlot.APIError); ok {
        switch e.StatusCode {
        case 401, 403:
            // Re-issue key or correct scope headers.
        case 429:
            // Request already retried; consider backoff tuning.
        default:
            // Log e.Code/e.Message/e.Body for diagnostics.
        }
        return
    }
    // Transport, decode, or SQL runtime error.
    return
}
// use res.Rows / res.RowCount
```

### Context timeouts

```go
cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
defer cancel()
_, err := proj.SQL(cctx, `SELECT 1`, nil)
if errors.Is(err, context.DeadlineExceeded) {
    // Increase timeout or reduce workload.
}
```

### Streaming errors

```go
sc, err := client.ExecSQLStream(ctx, projectID, warlot.SQLRequest{SQL: `SELECT * FROM big_table`})
if err != nil { return err }
defer sc.Close()

var row map[string]any
for sc.Next(&row) {
    // process
}
if err := sc.Err(); err != nil {
    // Handle malformed JSON or early disconnect.
}
```

### Idempotency on writes

```go
_, err := proj.SQL(ctx,
    `INSERT INTO audit(event) VALUES (?)`,
    []any{"create-123"},
    warlot.WithIdempotencyKey("audit-create-123"),
)
```

---

## Logging considerations

* The SDK’s optional logger redacts `x-api-key`.
* For sensitive environments, consider suppressing full response bodies; rely on `Message`, `Code`, and structured `Details`.

---

## Types (definitions)

Relevant selections for error handling:

```go
// Public error type for non-2xx responses.
type APIError struct {
    StatusCode int
    Body       string
    Message    string
    Code       string
    Details    any
}

// ExecSQL returns:
//   - *SQLResponse with OK=true for success,
//   - error when ok=false (SQL-reported error) or on any failure.
func (c *Client) ExecSQL(ctx context.Context, projectID string, req SQLRequest, opts ...CallOption) (*SQLResponse, error)

// Streaming scanner; terminal error accessed via Err().
type RowScanner struct {/* … */}
func (s *RowScanner) Next(dst any) bool
func (s *RowScanner) Err() error
func (s *RowScanner) Close() error
```

---

## Unit test definitions

### 1) Non-2xx produces `APIError` with parsed message

```go
// errors_apierror_parse_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_APIError_ParsesMessage(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"forbidden","code":"FORBIDDEN"}`, http.StatusForbidden)
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL))
	_, err := cl.ExecSQL(context.Background(), "P", SQLRequest{SQL: "SELECT 1"})
	if err == nil {
		t.Fatalf("expected error")
	}
	e, ok := err.(*APIError)
	if !ok || e.StatusCode != 403 || e.Message != "forbidden" || e.Code != "FORBIDDEN" {
		t.Fatalf("unexpected APIError: %#v", err)
	}
}
```

### 2) `429` is retried then succeeds

```go
// errors_retry_429_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func Test_Retry_429_ThenSuccess(t *testing.T) {
	var calls int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
            w.Header().Set("Retry-After", "1")
			http.Error(w, `{"message":"rate limited"}`, http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"row_count":0}`))
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL), WithRetries(2))
	if _, err := cl.ExecSQL(context.Background(), "P", SQLRequest{SQL: "CREATE TABLE t(x)"}); err != nil {
		t.Fatalf("unexpected failure after retry: %v", err)
	}
}
```

### 3) Decode error is propagated with context

```go
// errors_decode_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_DecodeError_IsReturned(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// invalid JSON
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"row_count":"not-an-int"}`))
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL))
	_, err := cl.ExecSQL(context.Background(), "P", SQLRequest{SQL: "CREATE TABLE t(x)"} )
	if err == nil || !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}
```

### 4) Context deadline surfaces as `context.DeadlineExceeded`

```go
// errors_context_timeout_test.go (package warlot)
package warlot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_ContextDeadlineExceeded(t *testing.T) {
	// Simulate a slow handler that outlives the context timeout.
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true,"row_count":0}`))
	}))
	defer s.Close()

	cl := New(WithBaseURL(s.URL))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := cl.ExecSQL(ctx, "P", SQLRequest{SQL: "CREATE TABLE t(x)"} )
	if err == nil || ctx.Err() == nil {
		t.Fatalf("expected context deadline error, got %v", err)
	}
}
```

---

## Troubleshooting

| Symptom                        | Likely cause                            | Action                                                                   |
| ------------------------------ | --------------------------------------- | ------------------------------------------------------------------------ |
| `APIError` with 401/403        | Missing key or scope mismatch           | Issue key; confirm holder/project headers.                               |
| `json: cannot unmarshal …`     | Upstream shape differs from expectation | Use untyped row access or adjust struct field types/tags.                |
| Persistent 429 despite retries | High request rate or tight limits       | Increase backoff, reduce concurrency, ensure idempotency on writes.      |
| Streaming terminates early     | Network interruption or malformed array | Inspect `RowScanner.Err()`; retry operation or adjust workload chunking. |

---

## Related topics

* Retries & rate limits: `10-retries-rate-limits.md`
* SQL execution and response shapes: `06-sql.md`
* Streaming and pagination: `07-streaming-pagination.md`
