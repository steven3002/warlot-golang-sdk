# Glossary

Definitions of terms used throughout the Warlot Go SDK and official documentation. The entries group related concepts and include short verification snippets where appropriate.

---

## Core entities

* **Project**
  Logical container for a dedicated SQL database. Identified by a **Project ID** and referenced by a human-readable **Project Name** and **Holder ID**.

* **Project ID (`project_id`, `ProjectID`)**
  UUID assigned at project initialization. Used in path segments for project-scoped endpoints.

* **Project Name (`x-project-name`)**
  Human-readable label for a project, forwarded in headers for scoping.

* **Holder ID (`x-holder-id`)**
  Address or account identifier that owns or manages the project.

* **DB ID (`db_id`, `DBID`)**
  Internal database identifier associated with a project.

---

## Authentication & headers

* **API Key (`x-api-key`)**
  Project-scoped credential issued via `IssueAPIKey`. Must match the `(holder_id, project_name, project_id)` tuple.

* **Idempotency Key (`x-idempotency-key`)**
  Optional header that ensures a write executes at most once across retries.

* **Base URL**
  Root API endpoint, for example `https://warlot-api.onrender.com`.

* **User-Agent**
  Client identifier sent with requests. Default value includes SDK name and version.

**Common headers (reference)**

| Header              | Purpose                                        |
| ------------------- | ---------------------------------------------- |
| `x-api-key`         | Authorization for project-scoped access        |
| `x-holder-id`       | Holder identity associated with the project    |
| `x-project-name`    | Human-readable project scope                   |
| `x-idempotency-key` | Optional, deduplicates write requests on retry |

---

## Project lifecycle operations

* **Resolve (ResolveProject)**
  Looks up an existing project by holder and project name. Response may include legacy fields (`ProjectID`, `DBID`) or modern fields (`project_id`, `db_id`). The SDK normalizes both.

* **Init (InitProject)**
  Creates a new project and returns identifiers including `ProjectID` and `DBID`.

* **Issue Key (IssueAPIKey)**
  Generates a project-scoped API key for authenticated access.

* **Status (GetProjectStatus)**
  Returns gateway-defined metadata describing the current state of a project.

* **Commit (CommitProject)**
  Persists project changes to chain-backed storage and returns a receipt object.

---

## SQL & data model

* **DDL (Data Definition Language)**
  Statements that change schema, for example `CREATE TABLE`. Responses typically include `{ok:true, row_count:0}`.

* **DML (Data Manipulation Language)**
  Statements that mutate data, for example `INSERT`, `UPDATE`, `DELETE`. Responses include `{ok:true, row_count:N}`.

* **SELECT (Query)**
  Read operations that return a row set, typically `{ok:true, rows:[...]}`.

* **Parameterized Query**
  SQL with `?` placeholders and a separate JSON `params` array to avoid injection.

* **Schema (GetTableSchema)**
  Gateway-defined structure describing table columns and types.

* **Table Count (GetTableCount)**
  Returns the number of tables in a project as an integer.

---

## SDK surfaces

* **Client**
  Primary entry point that holds configuration, HTTP transport, retry policy, and default headers. Methods cover projects, SQL, tables, status, and commit.

* **Project (wrapper)**
  Lightweight handle binding a project ID to the client, providing ergonomic methods that omit the explicit ID parameter.

* **Query[T]**
  Helper that maps `SELECT` rows into typed structs using `encoding/json`.

* **Migrator**
  File-based migration runner. Applies ordered `.sql` files and records progress in `_migrations`.

* **RowScanner**
  Streaming JSON row reader for large `SELECT` results. Supports `Next`, `Err`, and `Close`.

* **Pager**
  Helper that advances through paginated table browsing (`BrowseRows`) by maintaining `limit` and `offset`.

---

## Reliability & limits

* **Rate Limit (`429`)**
  Throttling signal from the service. The SDK retries with backoff and honors `Retry-After`.

* **`Retry-After`**
  Header that dictates the minimum delay before the next attempt; expressed in seconds or HTTP-date.

* **Backoff (Exponential with Jitter)**
  Retry strategy that increases delay between attempts and randomizes intervals to reduce thundering herd behavior.

* **Timeout / Context**
  Per-request deadline control. The SDK also has an internal `http.Client` timeout.

* **APIError**
  Error type returned on non-2xx responses with `StatusCode`, `Message`, optional `Code`, and raw `Body` fields.

---

## Tables, browsing, pagination

* **List Tables (ListTables)**
  Returns the array of table names within a project.

* **Browse Rows (BrowseRows)**
  Retrieves rows with `limit` and `offset` pagination.

* **Pagination**
  Iterative retrieval pattern for large datasets. The `Pager` type encapsulates repeated `BrowseRows` calls.

---

## Chain-backed fields (commit/status)

* **TxDigest**
  Transaction digest associated with a commit operation.

* **WriterPassID**
  Identifier related to write authorization artifacts when `IncludePass` is enabled.

* **BlobID**
  Reference to stored payload data associated with a commit.

* **CSVHashHex / DigestHex / SignatureHex**
  Hash and signature material emitted by commit operations, represented as hex strings.

---

## Logging & observability

* **Logger (hook)**
  Optional function invoked with structured metadata for requests and responses. API keys are redacted by the SDK before emission.

* **BeforeHooks / AfterHooks**
  Optional per-request and per-response callbacks for tracing, metrics, or additional headers.

---

## CLI

* **`warlotctl`**
  Single-binary CLI that wraps the SDK. Commands include `resolve`, `init`, `issue-key`, `sql`, `tables list`, `tables rows`, `schema`, `count`, `status`, and `commit`.

---

## Abbreviations & codes

| Term        | Meaning                       |
| ----------- | ----------------------------- |
| DDL         | Data Definition Language      |
| DML         | Data Manipulation Language    |
| UA          | User-Agent                    |
| UUID        | Universally Unique Identifier |
| JSON        | JavaScript Object Notation    |
| 2xx/4xx/5xx | HTTP status code classes      |

---

## Verification snippets (compact tests)

> These minimal definitions confirm common glossary expectations. Place alongside SDK tests as needed.

**1) Headers applied**

```go
func TestGlossary_AuthHeaders(t *testing.T) {
  s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("x-api-key") == "" ||
       r.Header.Get("x-holder-id") == "" ||
       r.Header.Get("x-project-name") == "" {
      t.Fatalf("required headers missing")
    }
    io.WriteString(w, `{"ok":true,"row_count":0}`)
  }))
  defer s.Close()
  cl := warlot.New(
    warlot.WithBaseURL(s.URL),
    warlot.WithAPIKey("k"),
    warlot.WithHolderID("h"),
    warlot.WithProjectName("n"),
  )
  if _, err := cl.ExecSQL(context.Background(), "P", warlot.SQLRequest{SQL: "CREATE TABLE t(x)"}); err != nil {
    t.Fatal(err)
  }
}
```

**2) Typed mapping**

```go
func TestGlossary_QueryTyped(t *testing.T) {
  s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    io.WriteString(w, `{"ok":true,"rows":[{"id":1,"name":"A","price":2.5}]}`)
  }))
  defer s.Close()
  cl := warlot.New(warlot.WithBaseURL(s.URL))
  p := cl.Project("P")
  type Row struct { ID int `json:"id"`; Name string `json:"name"`; Price float64 `json:"price"` }
  got, err := warlot.Query[Row](context.Background(), p, "SELECT ...", nil)
  if err != nil || len(got) != 1 { t.Fatalf("unexpected: %v %#v", err, got) }
}
```

**3) Idempotency forwarded**

```go
func TestGlossary_IdempotencyHeader(t *testing.T) {
  var key string
  s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    key = r.Header.Get("x-idempotency-key")
    io.WriteString(w, `{"ok":true,"row_count":1}`)
  }))
  defer s.Close()
  cl := warlot.New(warlot.WithBaseURL(s.URL))
  _, err := cl.ExecSQL(context.Background(), "P",
    warlot.SQLRequest{SQL:"INSERT INTO t(x) VALUES (?)", Params:[]any{1}},
    warlot.WithIdempotencyKey("k-1"),
  )
  if err != nil || key != "k-1" { t.Fatalf("unexpected: %q %v", key, err) }
}
```

**4) Retry-After honored**

```go
func TestGlossary_RetryAfter(t *testing.T) {
  first := true
  when := time.Now().Add(1 * time.Second).UTC().Format(http.TimeFormat)
  s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if first {
      first = false
      w.Header().Set("Retry-After", when)
      http.Error(w, `{"message":"rate"}`, http.StatusTooManyRequests)
      return
    }
    io.WriteString(w, `{"ok":true,"row_count":0}`)
  }))
  defer s.Close()
  cl := warlot.New(warlot.WithBaseURL(s.URL), warlot.WithRetries(2))
  start := time.Now()
  _, err := cl.ExecSQL(context.Background(), "P", warlot.SQLRequest{SQL:"CREATE TABLE t(x)"})
  if err != nil || time.Since(start) < 900*time.Millisecond {
    t.Fatalf("retry-after not honored or unexpected error: %v", err)
  }
}
```

---

## Cross-references

* Authentication: `03-authentication.md`
* Configuration: `04-configuration.md`
* Projects: `05-projects.md`
* SQL: `06-sql.md`
* Streaming & pagination: `07-streaming-pagination.md`
* Migrations: `08-migrations.md`
* Errors: `09-errors.md`
* Retries & rate limits: `10-retries-rate-limits.md`
* CLI: `11-cli.md`
* Types: `12-types.md`
* Testing: `13-testing.md`
* Security: `15-security.md`
* Versioning & changelog: `16-versioning-changelog.md`
