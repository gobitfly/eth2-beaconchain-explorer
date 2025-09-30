### Guidelines for LLM-based Coding Contributions

This document explains how a coding agent should produce changes that fit the structure, patterns, and expectations of this repository.

#### 1) Project layout and where to add code

- Binaries / entry points
    - Add new executables under `cmd/<name>/main.go`.
    - Reuse the existing CLI pattern: parse flags, show `version` when requested, read config via `utils.ReadConfig`, set `utils.Config`.
- HTTP features
    - Add new API endpoints in `handlers/` and register routes in `cmd/explorer/main.go`.
    - Add view templates under `templates/` and static assets under `static/`.
    - Use `utils.GetTemplateFuncs()` for common template helpers.
- Libraries / shared logic
    - New reusable packages should live at the repo root as top-level folders (e.g., `price`, `rpc`, `services`, `utils`).
- Database
    - Use the `db` package for connection management; do not build new connection pools in feature packages.
    - Distinguish Writer vs. Reader DB usage; use writer only for mutations.

#### 2) Configuration

- Always load config via `utils.ReadConfig(cfg, path)` and assign `utils.Config = cfg`.
- Validate critical config fields early (fatal if missing), especially chain parameters:
    - `SlotsPerEpoch`, `SecondsPerSlot`, `GenesisTimestamp`.
- Gate optional subsystems (metrics, pprof) behind config flags and verify `utils.Config` before use.
- If introducing new config, extend `types.Config` and ensure YAML/env parsing is supported by `utils.ReadConfig`.

#### 3) Logging and error handling

- Use `logrus` with module-scoped entries:
    - `logger := logrus.New().WithField("module", "<package or binary>")` or `logrus.StandardLogger().WithField(...)`.
- Fail fast on unrecoverable init errors with context:
    - `logrus.Fatalf(...)` or `utils.LogFatal(err, "message", 0)`.
- Prefer returning errors from library functions; only binaries should `Fatal`.
- Add context fields to logs (e.g., entity IDs, config source, batch counts) to aid debugging.

#### 4) HTTP routing, middleware, and security

- Routing
    - Use `gorilla/mux` and explicitly set HTTP methods, e.g. `.Methods("GET", "OPTIONS")`.
- Middleware
    - For APIs, ensure CORS via `utils.CORSMiddleware` when appropriate.
    - Use CSRF protection (`gorilla/csrf`) on state-changing POST endpoints rendered in the web app.
    - Keep gzip compression via `negroni-gzip` enabled for HTML/API where applicable.
- Health and monitoring
    - Expose module health through existing mechanisms (e.g., `service_status` table) and handlers in `handlers/api.go`.

#### 5) Database access patterns

- Connections
    - Initialize DBs via `db.MustInitDB`/`db.MustInitFrontendDB` (and `db.MustInitClickhouseDB` when needed) in the main binary, not inside libraries.
- Queries
    - Use parameterized SQL; avoid building raw SQL strings with user input.
    - For batch inserts/updates, use array parameters and `UNNEST`, with `ON CONFLICT` for idempotent upserts.
    - Example (pattern):
      ```sql
      INSERT INTO some_table(col_a, col_b)
      SELECT UNNEST($1::bytea[]), UNNEST($2::text[])
      ON CONFLICT (col_a) DO UPDATE SET col_b = EXCLUDED.col_b;
      ```
- Reader vs writer
    - Use `db.ReaderDb` for SELECTs and `db.WriterDb` for INSERT/UPDATE/DELETE.
- Ignore error to unresolved database functions and unknown tables in sql statement that are contained in a go file
- Ignore semantic errors related to an unresolved database table
#### 6) Data validation and normalization

- Normalize hex inputs: strip `0x`/`0X` prefixes and validate hex length before decoding.
- Enforce length constraints before DB writes (truncate or reject with error, depending on field semantics).
- Validate addresses/transactions/withdrawal credentials via helpers:
    - `utils.IsValidEth1Address`, `utils.IsValidEth1Tx`, `utils.IsValidWithdrawalCredentials`.
- Sanitize strings for templates; rely on `html/template` escaping and use safe helpers when returning `template.HTML`.

#### 7) Concurrency and performance

- Startup
    - Initialize independent subsystems concurrently using `sync.WaitGroup`.
- Long-running/batch jobs
    - Process in batches (e.g., 10k rows) and log progress with counts and duration.
    - Use bulk inserts with arrays/unnest for performance.
- Caching and rate limiting
    - Reuse existing packages (`cache`, `ratelimit`) where applicable.

#### 8) Testing, observability, and operations

- Metrics
    - When adding new services or critical code paths, expose counters/histograms via the existing `metrics` package.
- Logging
    - Include module names and key context fields; prefer structured logs.
- Feature flags
    - Use config-based toggles to enable/disable optional features.
- Pprof
    - If adding CPU/memory-intensive features, consider enabling pprof via config for debugging.

#### 9) Templates and frontend integration

- Place new pages under `templates/` and reference common layout patterns.
- Register template functions via `utils.GetTemplateFuncs()` rather than per-template reinventing.
- Keep client-side assets under `static/` and avoid bundling external CDNs unless necessary.

#### 10) Public APIs and backward compatibility

- Maintain existing routes and response shapes unless a deprecation path is documented.
- For new routes, version under a logical prefix if stability is a concern (e.g., `/api/v1/...`).
- Validate and sanitize any user inputs; avoid leaking internal errors or stack traces in responses.

#### 11) Code style and structure

- Go idioms
    - Keep functions small and cohesive; split large files by domain.
    - Prefer returning `(T, error)` over panics; libraries should not call `os.Exit`/`Fatal`.
    - Use standard naming (exported identifiers capitalized with doc comments when public).
- Module logging
    - Create a `logger` variable in each package: `var logger = logrus.StandardLogger().WithField("module", "<pkg>")`.
- Comments and docs
    - Add package comments and function docstrings for exported APIs; include usage examples if nontrivial.

#### 12) Example patterns to follow

- CSV/batch ingestion (like `cmd/validator-tagger`):
  ```go
  reader := csv.NewReader(file)
  reader.FieldsPerRecord = -1 // allow variable fields

  batchSize := 10_000
  batch := make(map[string]SomeValue, batchSize)
  for {
      row, err := reader.Read()
      if err == io.EOF { break }
      if err != nil { return fmt.Errorf("csv read: %w", err) }
      // validate, normalize, add to batch
      if len(batch) >= batchSize { flush(batch); batch = make(map[string]SomeValue, batchSize) }
  }
  flush(batch)
  ```

- DB bulk upsert with arrays/UNNEST:
  ```go
  _, err := db.WriterDb.Exec(`
      INSERT INTO validator_names (publickey, name)
      SELECT UNNEST($1::bytea[]), UNNEST($2::text[])
      ON CONFLICT (publickey) DO NOTHING
  `, pq.ByteaArray(pubkeys), pq.Array(names))
  if err != nil { return fmt.Errorf("upsert validator_names: %w", err) }
  ```

- Router setup for new handlers:
  ```go
  r := mux.NewRouter()
  r.HandleFunc("/api/v1/resource", handlers.ListResource).Methods("GET", "OPTIONS")
  r.HandleFunc("/api/v1/resource", handlers.CreateResource).Methods("POST", "OPTIONS")
  // add CORS/CSRF as needed
  ```

#### 13) What NOT to do

- Do not create ad-hoc DB connections inside handlers or utility packages.
- Do not log sensitive data (API keys, passwords, secrets).
- Do not concatenate SQL strings with user input; always use parameters.
- Do not bypass existing middleware for security (CORS/CSRF) on new routes.
- Do not introduce global variables for mutable state beyond established patterns (`utils.Config`, `db.*`).
- Do not use regex or pattern matching for extracting information from json objects
- Do not use abbreviated, very short or single letter variable names!!!!!!! Make sure every variable is named clearly and consistently.

#### 14) Pull request checklist for the agent

- [ ] New code placed in the correct package/folder (cmd vs handlers vs lib).
- [ ] Configuration added to `types.Config` and parsed via `utils.ReadConfig`.
- [ ] Reader/Writer/Frontend DB usage correct; no ad-hoc connections.
- [ ] Input validation and normalization implemented.
- [ ] Logging has module field and adequate context.
- [ ] Batch operations use arrays/UNNEST where appropriate.
- [ ] Routes registered with methods, protected with CORS/CSRF where relevant.
- [ ] Templates use `html/template` and helpers; no XSS risks introduced.
- [ ] Metrics or health checks updated if adding critical paths.
- [ ] Tests or at least manual testing notes included; pprof considered for heavy features.

#### 15) Plan before taking action
IMPORTANT: ALWAYS devise an INITIAL plan for implementing tasks and ask for user confirmation before proceeding!!!
If you have any questions, number them with letters to allow me to easily address them one by one
---

By following the above, contributions will be consistent with existing patterns and operational practices in this codebase.