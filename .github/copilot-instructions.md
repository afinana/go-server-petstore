<!-- Auto-generated guidance for AI coding agents on this repo -->
# Copilot / AI agent instructions — Go Petstore (Mongo)

Purpose: brief, actionable notes to help an AI agent make productive edits quickly.

- Repo shape: a small Go HTTP API generated from Swagger. Key code lives under `petstore/` and app entry is `main.go`.

- Architecture (big picture):
  - `main.go` wires dependencies: creates a Mongo client, builds `PetModel`/`StoreModel`/`UserModel` (each holds a `*mongo.Collection`) and passes them to `NewLog()` which returns an `*Application`.
  - `Application.NewRouter()` (in `petstore/routers.go`) registers routes (prefix `/petstore/v2/`) implemented as `Application` methods in `petstore/*.go` (e.g., `api_pet.go`, `api_store.go`, `api_user.go`).
  - Request middleware: `Logger()` in `petstore/logger.go` wraps handlers; Prometheus metrics are provided by `petstore/metrics.go` (route `/metrics` registered in router).
  - Data flow: HTTP handler -> `Application` method -> model (`PetModel`, `StoreModel`, `UserModel`) -> Mongo collection.

- Running & debugging:
  - Run locally: `go run main.go`.
  - Flags supported (see `main.go`): `-serverAddr`, `-mongoURI`, `-mongoDatabase`, `-enableCredentials`.
  - When `-enableCredentials=true`, supply `MONGODB_USERNAME` and `MONGODB_PASSWORD` environment variables.
  - Docker build/run examples are in `README.md`.

- Project-specific conventions and patterns (do NOT break these):
  - Routes are defined in `petstore/routers.go` using `github.com/gorilla/mux`. Paths are explicitly declared (no auto-discovery).
  - CORS: handlers check `OPTIONS` and call `app.enableCors(&w, r)` (see `petstore/application.go`). Preserve this pattern when adding endpoints.
  - Error/not-found: model methods return an error whose text equals `ErrNoDocuments` to signal not-found. Handlers check `err.Error() == "ErrNoDocuments"` — keep that string behavior when changing models or handlers.
  - ID handling: some models use Mongo `ObjectID` for `_id` (hex strings), while `Pet` also supports a numeric `id` field (handlers convert string to int). Be careful when changing ID parsing.
  - JSON responses: handlers set `Content-Type: Application/json; charset=UTF-8` and then call `json.NewEncoder(w).Encode(...)`. Preserve that header ordering.
  - Logging: basic `log` package is used; `infoLog` / `errLog` are passed from `main.go`. Use them consistently via `app.infoLog`/`app.errorLog`.

- Metrics & instrumentation:
  - `petstore/metrics.go` registers a Prometheus `http_requests_total` counter and exposes `/metrics` via `MetricsHandler()`.
  - `RecordMetrics(path, method, status string)` exists but is not wired into middleware. When adding metrics, increment using `RecordMetrics(routePattern, r.Method, strconv.Itoa(statusCode))` (or integrate inside `Logger()` to capture status).

- Database & models:
  - Models use `go.mongodb.org/mongo-driver`. Use `context.TODO()` in simple model methods for now to match existing style; prefer passing contextual `context.Context` if making broader changes.
  - Insert/Update sometimes return `InsertOneResult` even for updates (existing code). If you change semantics, update all callers.

- Files to consult when editing behavior:
  - `main.go` — startup, flags, Mongo client wiring, server config.
  - `petstore/routers.go` — route registration and metrics route.
  - `petstore/application.go` — `Application` struct, CORS and error helper.
  - `petstore/logger.go` — request logging middleware.
  - `petstore/metrics.go` — Prometheus setup and helper.
  - `petstore/mongo_*_model.go` — model patterns for DB access.
  - `petstore/api_*.go` — HTTP handlers & examples of error/response handling.

- Quick examples (copyable):
  - Start server locally:

    go run main.go -serverAddr=0.0.0.0:8080 -mongoURI=mongodb://localhost:27017 -mongoDatabase=petstore

  - Example metric increment (if adding manually in a handler):

    import "strconv"
    // ... after writing response status
    RecordMetrics("/petstore/v2/pet/{petId}", r.Method, strconv.Itoa(http.StatusOK))

- Tests & CI: repository has no tests or CI config. Any automated changes should include small smoke checks (e.g., run `go build` and run a sample request) before PR.

- PR guidance for agents: make minimal, focused changes; update `README.md` or this instructions file when adding infra or altering run flags; preserve backward-compatible JSON and route behavior.

If anything above is unclear or you'd like more detail (examples of model changes, wiring metrics into middleware, or adjusting ID handling), tell me which area to expand.
