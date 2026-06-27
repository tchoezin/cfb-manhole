# Backend Refactor Guide: Move JSON API Calls Out of `main.go`

This guide explains how to separate API/JSON handling logic from `main.go` into a dedicated file while keeping behavior the same.

## Goal

Keep `main.go` focused on server bootstrap only, and move request handling and JSON response logic into an API handler file.

## Suggested File Layout

- `main.go` (server startup and route registration)
- `handlers.go` (HTTP handlers and routing switch)
- `responses.go` (JSON and CORS helper functions)
- `scoring.go` (existing ranking/scoring logic; unchanged)

You can combine `handlers.go` and `responses.go` into one file if you prefer fewer files.

## Step-by-Step

1. Create a new file for handlers
- Add `handlers.go` in the same package (`package main`).
- Move these items from `main.go` to `handlers.go`:
  - `rankingEntry` type
  - `rerank(...)`
  - `handleRequest(...)`

2. Create a new file for JSON response helpers
- Add `responses.go` in `package main`.
- Move these helper functions from `main.go` to `responses.go`:
  - `setCommonHeaders(...)`
  - `sendJSON(...)`

3. Keep shared package access
- Because all files are in `package main`, moved functions can still call:
  - `getOrderedRankings()` from `scoring.go`
  - `http`, `json`, `strings`, and `url` imports in their new files

4. Simplify `main.go`
- Keep only startup concerns:
  - `main()`
  - `http.NewServeMux()`
  - `mux.HandleFunc("/", handleRequest)`
  - `http.ListenAndServe(...)`
- Remove imports no longer used in `main.go`.

5. Rebuild and run
- From `backend-cfb-manhole`, run:
  - `go run .`
- Confirm endpoints still respond:
  - `/api/rankings`
  - `/api/rankings/divisions`
  - `/api/rankings/divisions/{divisionId}`
  - `/health`

## Optional Improvement (Clean Architecture Lite)

If you want one extra separation layer, introduce a small service boundary between HTTP and scoring logic.

### Why this helps

- Handlers become thin and predictable (parse request -> call service -> map result to status code).
- Business rules can be unit tested without `net/http` setup.
- Future data source changes (static list, DB, API) can happen behind one interface.

### Recommended split

- `handlers.go`
- Responsibility: HTTP-only concerns (path parsing, status codes, response shape).
- Calls service methods and converts errors to JSON responses.

- `service.go`
- Responsibility: ranking use-cases and orchestration.
- Returns pure Go values and typed errors.

- `repository.go` (optional now, useful later)
- Responsibility: data access abstraction.
- First implementation can wrap existing in-memory/scoring functions.

### Example contracts

Use small interfaces and structs in `service.go`:

```go
type RankingRepository interface {
  OrderedRankings() []rankingEntry
}

type RankingService struct {
  repo RankingRepository
}

func NewRankingService(repo RankingRepository) *RankingService {
  return &RankingService{repo: repo}
}

func (s *RankingService) GetAll() []rankingEntry
func (s *RankingService) GetByDivision(name string) ([]rankingEntry, error)
func (s *RankingService) GetGroupedByDivision() map[int][]rankingEntry
```

Define typed errors in service layer:

```go
var ErrDivisionNotFound = errors.New("division not found")
var ErrInvalidDivision = errors.New("invalid division")
```

Then map errors in handlers:

- `ErrInvalidDivision` -> `400`
- `ErrDivisionNotFound` -> `404`
- anything else -> `500`

### Minimal first repository implementation

You can start with a local implementation that reuses current behavior:

```go
type scoringRepository struct{}

func (scoringRepository) OrderedRankings() []rankingEntry {
  return getOrderedRankings()
}
```

This keeps behavior unchanged while creating a clean seam for future DB/API storage.

### Incremental migration path (safe rollout)

1. Introduce `RankingService` and wire it in `main.go`.
2. Update only `/api/rankings` handler path to use service.
3. Verify response parity.
4. Migrate `/divisions` and `/division/{name}` endpoints.
5. Remove duplicated ranking logic from handlers after all routes use service.

### Suggested tests after this improvement

- `service_test.go`
- validates grouping/reranking behavior.
- validates division lookup success and not-found cases.

- `handlers_test.go`
- verifies status-code mapping and response JSON for success/error paths.
- can use a fake service to force each error condition.

This gives you cleaner boundaries now without introducing heavy framework or folder complexity.

## Quick Verification Checklist

- `main.go` has only bootstrap code.
- API behavior and JSON output format remain unchanged.
- CORS headers still present on `GET` and `OPTIONS` responses.
- `go run .` starts successfully.
- Existing frontend calls still work with no URL changes.
