# Troubleshooting

Operational issues typically fall into one of four categories: build/setup, transport/network, authentication/authorization, or SQL/response-shape. This guide presents a compact diagnosis flow, common symptoms, root causes, and corrective actions. Diagnostic snippets and minimal test definitions are included where helpful.

---

## Diagnosis workflow

```mermaid
%%{init: {'theme': 'base'}}%%
flowchart TD
  A[Observed failure] --> B{Compile-time or runtime?}
  B -->|Compile-time| C[Build/SDK setup checks]
  B -->|Runtime| D{HTTP response seen?}
  D -->|No| E[Transport error: DNS/TLS/timeout]
  D -->|Yes| F{Status code class}
  F -->|2xx| G{JSON decode or SQL error?}
  G -->|Decode| H[Shape/type mismatch]
  G -->|SQL err| I[Service reported statement failure]
  F -->|4xx (≠429)| J[Auth/scope/route]
  F -->|429| K[Rate limit: retry/backoff/idempotency]
  F -->|5xx| L[Transient: retry/backoff; raise if persistent]
  E --> M[Proxy/firewall/timeout tuning]
  J --> N[Headers/keys/scope alignment]
  H --> O[Typed mapping vs. untyped rows]
```

---

## Quick reference (symptom → action)

| Symptom (verbatim or similar)                             | Likely cause                                                | Corrective action                                                                                                                     |
| --------------------------------------------------------- | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `method must have no type parameters` (compile)           | Building against an outdated snapshot or incorrect file set | Ensure the SDK tree matches the organized version (no method-level generics). Use Go ≥ 1.21. Run `go clean -modcache && go mod tidy`. |
| `stat ./example/main.go: no such file or directory`       | Running an example path that does not exist                 | Invoke examples from the correct directory, or run tests via `go test ./warlot -v`.                                                   |
| `context deadline exceeded`                               | Network latency or too-short timeouts                       | Increase `http.Client.Timeout` and/or per-call context deadlines. Warm the API using `resolve`.                                       |
| `TLS handshake timeout` / `dial tcp … i/o timeout`        | Proxy/firewall or unstable network                          | Configure `HTTP_PROXY/HTTPS_PROXY/NO_PROXY`. Validate egress to the API host and port 443.                                            |
| `401 Unauthorized`                                        | Missing/invalid `x-api-key`                                 | Issue a key via `IssueAPIKey`, assign to `Client.APIKey`, confirm header forwarding.                                                  |
| `403 Forbidden`                                           | Holder/project mismatch or insufficient scope               | Ensure `x-holder-id` and `x-project-name` match the project used to issue the key.                                                    |
| `404 Not Found` on SQL path                               | Wrong project identifier or route                           | Resolve the project, then pass the returned `project_id` to `ExecSQL`.                                                                |
| `429 Too Many Requests`                                   | Rate limit exceeded                                         | Allow SDK retries; lower concurrency; add idempotency keys on writes; consider larger backoff.                                        |
| `500 Internal Server Error`                               | Transient backend condition                                 | Retried automatically; if persistent, capture request/response metadata and escalate.                                                 |
| `json: cannot unmarshal string into Go value of type int` | Response field type differs from expectation                | Use untyped row access or adjust struct field types; prefer `Query[T]` for typed mapping where columns match struct tags.             |
| `{"ok":false,"error":"…"}` with 200                       | Statement-level error                                       | Inspect SQL and parameters; test the statement in isolation via CLI `sql` command.                                                    |
| CLI: `missing required -project`                          | Flag omission                                               | Provide `-project` for project-scoped commands.                                                                                       |
| CLI: JSON params not parsed (Windows/PowerShell)          | Quoting/escaping differences                                | Wrap JSON arrays in double quotes and escape inner quotes, or read from a file (`@params.json`).                                      |

---

## Build & setup

* **Go toolchain**: Recommend Go **1.21+**. Confirm with `go version`.
* **Modules**:

  * Ensure the module path in `go.mod` matches the import path used by application code.
  * After adding the SDK, run `go mod tidy`.
* **Directory layout**:

  * SDK package lives in `warlot/` with tests alongside the code.
  * CLI (if included) builds from its own main file or `cmd/warlotctl`.

---

## Transport & network

### Common causes

* Corporate proxy or egress firewall blocking `https://warlot-api.onrender.com`.
* Insufficient timeouts for cold starts or high latency paths.
* DNS issues inside containers or WSL.

### Corrective actions

* Configure environment:

  ```bash
  export HTTPS_PROXY=http://proxy.local:3128
  export NO_PROXY=localhost,127.0.0.1,warlot-api.onrender.com
  ```
* Increase timeouts:

  ```go
  cl := warlot.New(warlot.WithHTTPClient(&http.Client{ Timeout: 90 * time.Second }))
  ```
* Pre-warm with resolve:

  ```go
  _, _ = cl.ResolveProject(ctx, warlot.ResolveProjectRequest{HolderID: "...", ProjectName: "..."})
  ```

---

## Authentication & headers

### Checklist

* `Client.APIKey` must be set after issuing a key for the *same* `(holder_id, project_name, project_id)`.
* `Client.HolderID` and `Client.ProjectName` must reflect the project that is being accessed.
* For CLI, ensure environment variables are exported **in the same shell** where the binary runs.

### Diagnostic snippet

```go
cl := warlot.New(
  warlot.WithAPIKey("…"),
  warlot.WithHolderID("0x…"),
  warlot.WithProjectName("proj"),
  warlot.WithLogger(func(e string, m map[string]any) { fmt.Println(e, m) }),
)
```

The logger redacts `x-api-key` and prints attempt metadata.

---

## SQL execution & response shapes

### Statement failure with `ok=false`

The gateway may return HTTP 200 with `{ "ok": false, "error": "…" }`. The SDK surfaces this as `error` alongside the parsed response. Inspect SQL syntax and parameter counts/types.

### Typed vs. untyped reads

* Untyped path:

  ```go
  res, _ := proj.SQL(ctx, `SELECT id, price FROM t`, nil)
  for _, r := range res.Rows {
    // JSON numbers are float64 by default
    _ = r["id"].(float64)
  }
  ```
* Typed mapping:

  ```go
  type Row struct{ ID int `json:"id"`; Price float64 `json:"price"` }
  rows, _ := warlot.Query[Row](ctx, proj, `SELECT id, price FROM t`, nil)
  ```

### Known edge case: field types

If the service emits numeric values as strings in a specific route, prefer untyped access or compatible struct fields (`string` for that property) and convert explicitly.

---

## Streaming & pagination

### Streaming terminates early

* Cause: upstream closed connection or malformed JSON.
* Action: check `scanner.Err()` after loop; retry operation or switch to paginated reads.

### Resource hygiene

Always `defer scanner.Close()` to release the underlying response body.

---

## Retries & rate limits

* Automatic retries for `429` and `5xx` with jittered exponential backoff; `Retry-After` is honored.
* For **writes**, set idempotency keys to avoid duplicate effects:

  ```go
  _, _ = proj.SQL(ctx, `INSERT INTO t(x) VALUES (?)`, []any{1}, warlot.WithIdempotencyKey("t-x-1"))
  ```
* If rate limiting persists:

  * Reduce client-side concurrency.
  * Increase backoff ceilings (`WithBackoff`).
  * Verify that long chains of retries do not exceed the per-request context deadline.

---

## Migrations

### Nothing applied

* Directory path incorrect or missing `.sql` suffix.
* Ledger `_migrations` table creation failed.

### Partial progress after error

* Expected behavior. Fix the failing file and re-run; the runner skips already recorded entries.

Diagnostic call:

```go
var m warlot.Migrator
applied, err := m.Up(ctx, proj, embeddedFS, "migrations")
```

---

## CLI

### Param quoting on Windows/PowerShell

* Use:

  ```powershell
  .\warlotctl sql -project $env:P -q "INSERT INTO t(x) VALUES (?)" -params "[1]"
  ```
* Alternative: read params from a file

  ```bash
  echo '["A", 1.23]' > /tmp/params.json
  ./warlotctl sql -project "$P" -q 'INSERT INTO t(a,b) VALUES (?,?)' -params @/tmp/params.json
  ```

### Verifying header forwarding

Run `tables list` behind a stub and echo headers to confirm `x-api-key`, `x-holder-id`, and `x-project-name`.

---

## Logging & diagnostics

Enable the SDK logger and request/response hooks:

```go
cl := warlot.New(
  warlot.WithLogger(func(evt string, meta map[string]any) {
    // persist meta to a structured sink; api key is already redacted
  }),
)
cl.BeforeHooks = append(cl.BeforeHooks, func(r *http.Request) { /* trace id */ })
cl.AfterHooks  = append(cl.AfterHooks, func(res *http.Response, body []byte, err error) { /* metrics */ })
```

When reporting an issue, capture:

* SDK version and commit (if applicable)
* Go version, OS/arch
* Base URL and timestamp
* Status codes and `Retry-After` values
* Minimal reproducer (redacting secrets)
* Whether a proxy is in use

---

## Minimal health checks

A lightweight liveness probe can reduce cold-start surprises:

```go
func Warmup(ctx context.Context, cl *warlot.Client, holder, pname string) error {
  _, err := cl.ResolveProject(ctx, warlot.ResolveProjectRequest{HolderID: holder, ProjectName: pname})
  return err
}
```

---

## Type excerpts (for reference)

```go
type APIError struct {
  StatusCode int
  Body       string
  Message    string
  Code       string
  Details    any
}

type SQLResponse struct {
  OK       bool
  RowCount *int
  Rows     []map[string]any
  Error    string
}
```

---

## Unit-test definitions (diagnostic patterns)

### Decode error surfaces cleanly

```go
// troubleshooting_decode_error_test.go
package warlot

import (
  "context"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
)

func Test_Troubleshooting_DecodeError(t *testing.T) {
  s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    w.Write([]byte(`{"ok":true,"row_count":"oops"}`)) // wrong type
  }))
  defer s.Close()

  cl := New(WithBaseURL(s.URL))
  _, err := cl.ExecSQL(context.Background(), "P", SQLRequest{SQL: "CREATE TABLE t(x)"})
  if err == nil || !strings.Contains(err.Error(), "decode response") {
    t.Fatalf("expected decode error, got %v", err)
  }
}
```

### Retry path honored for 429

```go
// troubleshooting_retry_429_test.go
package warlot

import (
  "context"
  "net/http"
  "net/http/httptest"
  "sync/atomic"
  "testing"
)

func Test_Troubleshooting_Retry429(t *testing.T) {
  var n int32
  s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if atomic.AddInt32(&n, 1) == 1 {
      http.Error(w, `{"message":"rate limited"}`, http.StatusTooManyRequests)
      return
    }
    w.WriteHeader(200)
    w.Write([]byte(`{"ok":true,"row_count":0}`))
  }))
  defer s.Close()

  cl := New(WithBaseURL(s.URL), WithRetries(1))
  if _, err := cl.ExecSQL(context.Background(), "P", SQLRequest{SQL: "CREATE TABLE t(x)"}); err != nil {
    t.Fatalf("unexpected error after retry: %v", err)
  }
}
```

---

## Final checks

Before escalating any persistent issue, confirm:

1. Toolchain and module setup are correct (`go version`, `go mod tidy`).
2. Network egress to the API host is possible; proxy variables are configured when required.
3. Correct holder/project headers and a valid API key are present.
4. For writes, idempotency keys are supplied.
5. For large reads, streaming or pagination is employed, and scanners are closed.
