# CLI (`warlotctl`)

A single-binary command-line interface for interacting with the Warlot SQL Database API during development, testing, and operational tasks. The CLI is built on top of the official Go SDK and mirrors the same request/response semantics.

---

## Installation

### Build from source

```bash
# from repository root
go build -o warlotctl ./warlot-go/cmd/warlotctl
# verify
./warlotctl -h
```

*(If the CLI lives in a single file, adjust the path accordingly, for example `./warlot-go/warlotctl.go`.)*

---

## Global behavior

* Default output format is JSON emitted to `stdout`; errors are printed to `stderr`.
* Exit codes:

  * `0` on success
  * `>0` on failure (transport errors, non-2xx HTTP, or SQL errors)
* Authentication headers are auto-applied from environment variables and/or flags (see below).
* Requests honor retry/backoff in the underlying SDK.

---

## Environment variables

| Variable                 | Purpose                   | Example                           |
| ------------------------ | ------------------------- | --------------------------------- |
| `WARLOT_BASE_URL`        | API base URL              | `https://warlot-api.onrender.com` |
| `WARLOT_API_KEY`         | Default API key           | `a2f5…37e0`                       |
| `WARLOT_HOLDER`          | Default holder identifier | `0x2e4a…7ba3`                     |
| `WARLOT_PNAME`           | Default project name      | `project_alpha`                   |
| `WARLOT_TIMEOUT`         | Request timeout (seconds) | `90`                              |
| `WARLOT_RETRIES`         | Max retries for 429/5xx   | `6`                               |
| `WARLOT_BACKOFF_INIT_MS` | Initial backoff (ms)      | `500`                             |
| `WARLOT_BACKOFF_MAX_MS`  | Max backoff (ms)          | `8000`                            |

---

## Global flags

These flags can appear before any subcommand:

| Flag            | Description                                                | Default                              |
| --------------- | ---------------------------------------------------------- | ------------------------------------ |
| `-base`         | Base URL (overrides `WARLOT_BASE_URL`)                     | autodetected from env or SDK default |
| `-apikey`       | API key (overrides `WARLOT_API_KEY`)                       | empty                                |
| `-holder`       | Holder ID (overrides `WARLOT_HOLDER`)                      | empty                                |
| `-pname`        | Project name (overrides `WARLOT_PNAME`)                    | empty                                |
| `-timeout`      | Timeout in seconds (overrides `WARLOT_TIMEOUT`)            | 30                                   |
| `-retries`      | Max retries (overrides `WARLOT_RETRIES`)                   | 3                                    |
| `-backoff-init` | Initial backoff in ms (overrides `WARLOT_BACKOFF_INIT_MS`) | 300                                  |
| `-backoff-max`  | Max backoff in ms (overrides `WARLOT_BACKOFF_MAX_MS`)      | 3000                                 |
| `-pretty`       | Pretty-print JSON output                                   | false                                |
| `-ua`           | Custom User-Agent suffix                                   | empty                                |

Example:

```bash
./warlotctl -base "$WARLOT_BASE_URL" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" -pretty resolve
```

---

## Commands

### 1) `resolve`

Resolve a project by `(holder_id, project_name)`.

```bash
./warlotctl resolve [-holder H] [-pname N]
```

**Output (example):**

```json
{
  "exists_meta": true,
  "exists_chain": true,
  "project_id": "3c58bfc5-64a9-45ff-9510-a6584ee96248",
  "db_id": "24840e28-14c6-46ea-a313-2d6d3e32e3ef",
  "action": "ready"
}
```

**Notes:** Legacy fields `ProjectID`/`DBID` may be normalized in the CLI output.

---

### 2) `init`

Initialize a new project.

```bash
./warlotctl init -holder H -pname N -owner O [--include-pass] [--deletable]
```

**Flags:**

| Flag                                                                          | Description                       | Default  |
| ----------------------------------------------------------------------------- | --------------------------------- | -------- |
| `-owner`                                                                      | Owner address                     | required |
| `--include-pass`                                                              | Include writer pass               | true     |
| `--deletable`                                                                 | Mark project as deletable         | true     |
| `-epoch-set` `-cycle-end` `-writers-len` `-track-back-len` `-draft-epoch-dur` | Advanced fields mirrored from API | 0        |

**Output (example):**

```json
{
  "ProjectID":"d4652d80-083b-4807-9c23-a85e5deeb3f0",
  "DBID":"0x3adaecee0ae0653f0ee844ed9e58744facdcd643ca754951c23f1ceaf9bb8044",
  "WriterPassID":"",
  "BlobID":"",
  "TxDigest":"",
  "CSVHashHex":"",
  "DigestHex":"",
  "SignatureHex":""
}
```

---

### 3) `issue-key`

Issue an API key for a specific project.

```bash
./warlotctl issue-key -project P -holder H -pname N -user U
```

**Output:**

```json
{ "apiKey": "6f80...84aa", "url": "https://api.warlot.com/3c58bfc5-..." }
```

**Tip:** Export the key for subsequent operations:

```bash
export WARLOT_API_KEY=$(./warlotctl issue-key -project "$P" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" -user "$WARLOT_OWNER" | jq -r .apiKey)
```

---

### 4) `sql`

Execute an SQL statement against a project.

```bash
./warlotctl sql -project P -q 'SQL...' [-params 'JSON-ARRAY'] [-idempotency K]
```

**Flags:**

| Flag           | Description              | Example                    |
| -------------- | ------------------------ | -------------------------- |
| `-q`           | SQL statement (required) | `'SELECT * FROM products'` |
| `-params`      | JSON array of parameters | `'["Laptop", 999.99]'`     |
| `-idempotency` | Idempotency key header   | `insert-2025-10-03-01`     |

**Responses:**

* DDL/DML:

  ```json
  {"ok": true, "row_count": 1}
  ```
* SELECT:

  ```json
  {"ok": true, "rows": [{"id":1,"name":"X","price":1.23}]}
  ```

---

### 5) `tables list`

List all tables in a project.

```bash
./warlotctl tables list -project P
```

**Output:**

```json
{"tables": ["products","categories"]}
```

---

### 6) `tables rows`

Browse rows with pagination.

```bash
./warlotctl tables rows -project P -table T [-limit N] [-offset M]
```

**Output:**

```json
{
  "limit": 2,
  "offset": 0,
  "table": "products",
  "rows": [{"id":1,"name":"A"},{"id":2,"name":"B"}]
}
```

---

### 7) `schema`

Fetch a table schema.

```bash
./warlotctl schema -project P -table T
```

**Output:** Arbitrary JSON describing table columns and types (gateway-defined).

---

### 8) `count`

Return total number of tables.

```bash
./warlotctl count -project P
```

**Output:**

```json
{"project_id":"P","table_count": 3}
```

---

### 9) `status`

Fetch project status.

```bash
./warlotctl status -project P
```

**Output:** Arbitrary JSON (gateway-defined).

---

### 10) `commit`

Commit project changes to chain.

```bash
./warlotctl commit -project P
```

**Output:** Gateway-defined JSON commit receipt.

---

## Examples

**Resolve or init → issue key → create table**

```bash
# Resolve
./warlotctl resolve -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" | tee resolve.json

# If project_id is absent, initialize
P=$(jq -r '.project_id // empty' resolve.json)
if [ -z "$P" ]; then
  P=$(./warlotctl init -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" -owner "$WARLOT_OWNER" | jq -r .ProjectID)
fi
echo "PROJECT_ID=$P"

# Issue key
export WARLOT_API_KEY=$(./warlotctl issue-key -project "$P" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" -user "$WARLOT_OWNER" | jq -r .apiKey)

# Create table
./warlotctl sql -project "$P" \
  -q 'CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL)'
```

**Insert with idempotency and query**

```bash
./warlotctl sql -project "$P" \
  -q 'INSERT INTO products (name, price) VALUES (?, ?)' \
  -params '["Laptop", 999.99]' \
  -idempotency 'cli-insert-1'

./warlotctl sql -project "$P" \
  -q 'SELECT id, name, price FROM products ORDER BY id'
```

**Paginate rows**

```bash
./warlotctl tables rows -project "$P" -table products -limit 10 -offset 0 | jq .
```

**Commit**

```bash
./warlotctl commit -project "$P" | jq .
```

---

## JSON shapes (reference)

* **Resolve**: may include either modern fields (`exists_meta`, `exists_chain`, `project_id`, `db_id`, `action`) or legacy (`ProjectID`, `DBID`). The CLI normalizes both.
* **Init**: matches `InitProjectResponse`.
* **Issue key**: `{"apiKey": "...","url": "..."}`.
* **SQL**: DDL/DML `{ok,row_count}`; SELECT `{ok,rows}`.
* **Tables list**: `{"tables":["…"]}`.
* **Rows**: `{"limit":N,"offset":M,"table":"T","rows":[...]}`.
* **Count**: `{"project_id":"…","table_count":N}`.
* **Status/Commit**: gateway-defined objects.

---

## Type definitions (SDK excerpts used by the CLI)

```go
// Requests
type ResolveProjectRequest struct {
  HolderID    string `json:"holder_id"`
  ProjectName string `json:"project_name"`
}
type InitProjectRequest struct {
  HolderID      string `json:"holder_id"`
  ProjectName   string `json:"project_name"`
  OwnerAddress  string `json:"owner_address"`
  EpochSet      int    `json:"epoch_set"`
  CycleEnd      int    `json:"cycle_end"`
  WritersLen    int    `json:"writers_len"`
  TrackBackLen  int    `json:"track_back_len"`
  DraftEpochDur int    `json:"draft_epoch_dur"`
  IncludePass   bool   `json:"include_pass"`
  Deletable     bool   `json:"deletable"`
}
type IssueKeyRequest struct {
  ProjectID     string `json:"projectId"`
  ProjectHolder string `json:"projectHolder"`
  ProjectName   string `json:"projectName"`
  User          string `json:"user"`
}
type SQLRequest struct {
  SQL    string        `json:"sql"`
  Params []interface{} `json:"params"`
}

// Responses
type ResolveProjectResponse struct {
  ExistsMeta      bool   `json:"exists_meta"`
  ExistsChain     bool   `json:"exists_chain"`
  ProjectID       string `json:"project_id"`
  DBID            string `json:"db_id"`
  Action          string `json:"action"`
  LegacyProjectID string `json:"ProjectID,omitempty"`
  LegacyDBID      string `json:"DBID,omitempty"`
}
type InitProjectResponse struct {
  ProjectID    string `json:"ProjectID"`
  DBID         string `json:"DBID"`
  WriterPassID string `json:"WriterPassID"`
  BlobID       string `json:"BlobID"`
  TxDigest     string `json:"TxDigest"`
  CSVHashHex   string `json:"CSVHashHex"`
  DigestHex    string `json:"DigestHex"`
  SignatureHex string `json:"SignatureHex"`
}
type IssueKeyResponse struct {
  APIKey string `json:"apiKey"`
  URL    string `json:"url"`
}
type SQLResponse struct {
  OK       bool                     `json:"ok"`
  RowCount *int                     `json:"row_count,omitempty"`
  Rows     []map[string]interface{} `json:"rows,omitempty"`
  Error    string                   `json:"error,omitempty"`
}
type ListTablesResponse struct {
  Tables []string `json:"tables"`
}
type BrowseRowsResponse struct {
  Limit  int                      `json:"limit"`
  Offset int                      `json:"offset"`
  Table  string                   `json:"table"`
  Rows   []map[string]interface{} `json:"rows"`
}
type TableCountResponse struct {
  ProjectID  string `json:"project_id"`
  TableCount int    `json:"table_count"`
}
```

---

## Tests (definitions)

The CLI can be tested with an `httptest.Server` that emulates the API, running the compiled binary with `os/exec`.

### 1) Resolve prints normalized JSON

```go
// cli_resolve_test.go
package cli_test

import (
  "bytes"
  "net/http"
  "net/http/httptest"
  "os/exec"
  "strings"
  "testing"
)

func Test_CLI_Resolve_Normalized(t *testing.T) {
  srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/warlotSql/projects/resolve" {
      w.Header().Set("Content-Type", "application/json")
      w.Write([]byte(`{"ProjectID":"P-1","DBID":"DB-1"}`)) // legacy
      return
    }
    w.WriteHeader(404)
  }))
  defer srv.Close()

  var out bytes.Buffer
  cmd := exec.Command("./warlotctl", "-base", srv.URL, "resolve", "-holder", "H", "-pname", "N")
  cmd.Stdout = &out
  cmd.Stderr = &out
  if err := cmd.Run(); err != nil {
    t.Fatalf("run: %v\n%s", err, out.String())
  }
  s := out.String()
  if !strings.Contains(s, `"project_id":"P-1"`) {
    t.Fatalf("normalized project_id missing: %s", s)
  }
}
```

### 2) SQL DDL returns `row_count`

```go
// cli_sql_rowcount_test.go
package cli_test

import (
  "bytes"
  "net/http"
  "net/http/httptest"
  "os/exec"
  "strings"
  "testing"
)

func Test_CLI_SQL_RowCount(t *testing.T) {
  srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if strings.HasSuffix(r.URL.Path, "/sql") {
      w.Header().Set("Content-Type", "application/json")
      w.Write([]byte(`{"ok":true,"row_count":1}`))
      return
    }
    w.WriteHeader(404)
  }))
  defer srv.Close()

  var out bytes.Buffer
  cmd := exec.Command("./warlotctl", "-base", srv.URL, "sql", "-project", "P", "-q", "CREATE TABLE t(x)")
  cmd.Stdout = &out
  cmd.Stderr = &out
  if err := cmd.Run(); err != nil {
    t.Fatalf("run: %v\n%s", err, out.String())
  }
  if !strings.Contains(out.String(), `"row_count":1`) {
    t.Fatalf("row_count not present: %s", out.String())
  }
}
```

### 3) Tables list forwards auth headers from env

```go
// cli_tables_headers_test.go
package cli_test

import (
  "bytes"
  "net/http"
  "net/http/httptest"
  "os"
  "os/exec"
  "testing"
)

func Test_CLI_AuthHeaders_FromEnv(t *testing.T) {
  seen := make(map[string]string)
  srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    seen["x-api-key"] = r.Header.Get("x-api-key")
    seen["x-holder-id"] = r.Header.Get("x-holder-id")
    seen["x-project-name"] = r.Header.Get("x-project-name")
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"tables":["t"]}`))
  }))
  defer srv.Close()

  var out bytes.Buffer
  cmd := exec.Command("./warlotctl", "-base", srv.URL, "tables", "list", "-project", "P")
  cmd.Stdout = &out
  cmd.Stderr = &out
  cmd.Env = append(os.Environ(),
    "WARLOT_API_KEY=K",
    "WARLOT_HOLDER=H",
    "WARLOT_PNAME=N",
  )
  if err := cmd.Run(); err != nil {
    t.Fatalf("run: %v\n%s", err, out.String())
  }
  if seen["x-api-key"] != "K" || seen["x-holder-id"] != "H" || seen["x-project-name"] != "N" {
    t.Fatalf("headers not forwarded: %#v", seen)
  }
}
```

---

## Troubleshooting

| Symptom                          | Likely cause                      | Action                                                           |
| -------------------------------- | --------------------------------- | ---------------------------------------------------------------- |
| `missing required -project`      | Flag omitted                      | Provide `-project` for project-bound commands                    |
| `401/403 Unauthorized/Forbidden` | API key missing or scope mismatch | Issue key via `issue-key`; confirm `-holder` and `-pname`        |
| `429 Too Many Requests`          | Rate limits                       | Reduce call rate; rely on retries; add `-idempotency` for writes |
| `500 Internal Server Error`      | Transient                         | Retry, or increase `-retries` / backoff settings                 |
| `context deadline exceeded`      | Timeout too small                 | Increase `-timeout` or use environment overrides                 |

---

## Related topics

* Authentication and headers: `03-authentication.md`
* Configuration knobs and retry/backoff: `04-configuration.md`, `10-retries-rate-limits.md`
* SQL execution model: `06-sql.md`
* Streaming and pagination: `07-streaming-pagination.md`
* Migrations: `08-migrations.md`
