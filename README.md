# Warlot Go SDK & CLI (warlotdev) 🌌

A production-grade Go SDK and companion CLI for the Warlot SQL Database API.
This repository enables fast database-backed development with project isolation, encrypted key issuance, parameterized SQL, streaming, migrations, and operational tooling—shipped as an idiomatic Go package and a single-binary CLI.

---

## Highlights ✨

* **Purpose-built SDK** for the Warlot SQL API, designed for reliability and clarity.
* **Single-binary CLI (warlotdev)** for development, automation, and operations.
* **Project lifecycle tools**: resolve or initialize projects, then issue API keys.
* **SQL first-class support**: DDL, DML, SELECT, parameter binding, typed decoding, and streaming for large result sets.
* **Operational introspection**: list tables, browse rows, inspect schemas, fetch status, and commit changes.
* **Resilience**: retry with exponential backoff, rate-limit awareness, idempotency keys, and request timeouts.
* **Migrations**: idempotent, ordered, and tracked via a dedicated metadata table.
* **Security posture**: encrypted API keys, header-based auth, redaction-aware logging hooks.

---

## Repository Structure 🗂️

* **warlot-go/warlot**
  Go SDK: client configuration, request/response models, SQL helpers, streaming, migrations, pagination, and error handling.

* **warlot-go/cmd/warlotdev**
  CLI application: commands mirror the SDK surfaces for project lifecycle, SQL execution, table operations, status, and commit.

* **sdk documentation**
  Developer documentation: installation, quickstart, authentication, configuration, SQL, streaming, migrations, errors, retries, CLI usage, types, testing, troubleshooting, security, versioning, FAQ, glossary, support, and flowcharts.

---

## Core Capabilities 🚀

### Project Lifecycle

* Initialize a new project with project metadata and policy flags.
* Resolve an existing project via holder identifier and human-readable name.
* Issue encrypted API keys with project-scoped access.

### SQL Execution & Data Access

* Execute DDL and DML with parameterized queries.
* Run SELECT statements with flexible decoding of row sets.
* Use typed mapping helpers to bind result rows into strongly-typed Go structs.
* Stream large result sets to avoid excessive memory usage.

### Tables & Introspection

* List tables for a project.
* Browse rows with limit and offset pagination.
* Fetch per-table schema details.
* Retrieve aggregate table counts.

### Status & Commit

* Retrieve project status information for diagnostics.
* Commit project changes to chain-backed storage when required.

---

## Reliability & Performance ⚙️

* **Context-aware timeouts** to keep workflows responsive.
* **Retry with jittered backoff** for transient errors and rate limiting.
* **Rate-limit surface** with Retry-After parsing to adapt pause durations.
* **Idempotency keys** for write safety across retries.
* **Pluggable logging hooks** with API key redaction to protect sensitive values.
* **Pagination helpers** to traverse large tables predictably.

---

## Migrations 🧭

* Apply ordered migrations sourced from an embedded or file-backed filesystem.
* Idempotent execution guarded by a metadata table that tracks applied steps.
* Safe to re-run; only unapplied migrations are executed.

---

## Security Considerations 🔐

* Encrypted API keys designed for storage on chain-backed systems.
* Multi-header authentication pattern (API key, holder, project name).
* Parameterized SQL to mitigate injection risks.
* Optional logging hooks with strict redaction of secrets.

Security posture depends on correct API-key handling, stable project metadata, and consistent use of parameterization across all SQL execution points.

---

## CLI (warlotdev) 🧰

A single-binary tool that mirrors SDK features for local development and automation.

**Primary Commands**

* Project operations: resolve, init, issue-key
* SQL operations: execute queries, manage schema, fetch data
* Table operations: list, browse, schema, count
* Status and commit: operational checks and persistence

**Behavior**

* JSON output to standard output; errors to standard error
* Non-zero exit codes on failures
* Global flags for base URL, credentials, timeouts, retries, and backoff config
* Verbose mode for diagnostics with secret redaction

---

## Installation & Setup 📦

Multiple installation channels are supported:

* **Go toolchain**: install from the submodule path at a published tag; ensure the resulting binary directory is included in the system PATH.
* **Prebuilt binaries**: download platform-specific archives from the project’s Releases page, extract, and place the executable in a PATH directory.
* **Package managers**: Homebrew (macOS/Linux) and Scoop (Windows), if configured in the release pipeline.
* **Containers**: a minimal image with the CLI as entrypoint can be provided for CI or sealed environments.

Runtime configuration is supplied via environment variables for base URL, API key, holder identifier, project name, and networking/retry tunables.

---

## Typical Workflows 🧩

* **Bootstrap a project**
  Resolve an existing project by holder and name; if unresolved, initialize a new project and issue an API key.

* **Define schema and ingest data**
  Create or evolve tables, insert and update records with parameterized SQL, query results with or without streaming.

* **Operational visibility**
  List tables, browse rows, and inspect schema for audits and diagnostics. Fetch project status and commit changes.

* **Automation and CI**
  Run CLI tasks inside pipelines for provisioning, migrations, and verification steps, with idempotency keys safeguarding retries.

---

## Versioning & Tags 🏷️

* The Go SDK and CLI reside in a submodule. Tags follow the submodule-scoped convention:
  warlot-go/vX.Y.Z.
* Semantic Versioning is observed.
  v1 and below do not require a module path suffix; v2 and above require a /v2 suffix in the module path.

---

## Documentation Map 📚

* Installation and Quickstart
* Authentication and Configuration
* Projects, SQL, Streaming & Pagination
* Migrations and Idempotency
* Errors, Retries & Rate Limits
* CLI Reference
* Types and Testing
* Troubleshooting and Security
* Versioning, Changelog, FAQ, Glossary, Support
* Flowcharts illustrating project lifecycle, auth issuance, streaming, and migrations

Each topic is maintained in the sdk documentation directory with official guidance and examples.

---

## Contributing 🤝

Issues and pull requests are welcome. Proposed changes should maintain the security posture, typed API surface, resilience guarantees, and documentation quality established by this repository. Before submitting changes, ensure tests pass and public interfaces remain stable or are semantically versioned.

---

## License ⚖️

Distributed under the license specified in the repository’s LICENSE file.

---

## Support 😎

For general questions, integration concerns, or operational incidents, refer to the Support section in the documentation. Production-impacting issues should be reported with context such as API base, project identifier, timestamp, and observed error payloads where possible.

---

## Acknowledgements 🌱

This project integrates Go best practices for HTTP clients, robust retries, typed mapping, and streaming to deliver a pragmatic developer experience for the Warlot SQL Database API.
