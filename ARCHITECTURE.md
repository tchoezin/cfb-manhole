# Repository Architecture

## Overview

This repository is a small two-tier application for displaying college football pick'em rankings:

- `backend-cfb-manhole/` is a Go HTTP API backed by Neon Postgres.
- `frontend-cfb-manhole/` is a Next.js application that fetches and renders ranking data.

The system remains intentionally small, but the backend now persists player records in Neon Postgres. The frontend and backend are still separate runtimes, and the current leaderboard endpoint remains the main integration point.

## High-Level Topology

```text
Browser
  |
  v
Next.js frontend (React + Chakra UI)
  |
  | GET /api/rankings
  v
Go HTTP API
  |
  v
Neon Postgres
```

## Mermaid Diagram

```mermaid
flowchart TD
    Browser[Browser] --> Frontend[Next.js Frontend\nReact 19 + Chakra UI]
    Frontend --> ApiClient[src/lib/api.ts\nFetchRankings()]
    ApiClient --> Backend[Go HTTP API\n127.0.0.1:8000]
    Backend --> Handler[HTTP Handler\nhandlers.go]
    Handler --> Service[RankingService\nservice.go]
    Handler --> PlayerService[PlayerService\nservice.go]
    Service --> Repo[Postgres repository\ndatabase.go]
    PlayerService --> Repo
    Repo --> Neon[(Neon Postgres)]
    Repo -. seed source .-> DomainData[In-memory seed data\nscoring.go]
```

## Repository Structure

### Root

- `README.md`: minimal top-level project marker.
- `package.json`: top-level JavaScript dependencies only; not the primary runtime entrypoint for the app.
- `backend-cfb-manhole/`: Go backend.
- `frontend-cfb-manhole/`: Next.js frontend.

The backend and frontend are developed and run as separate applications.

## Backend Architecture

### Runtime Model

The backend is a single-process HTTP server started from `backend-cfb-manhole/main.go`.

- It creates a `RankingService`.
- It creates a `PlayerService`.
- It registers one handler on `http.ServeMux`.
- It listens on `127.0.0.1:8000`.

There is no router framework, dependency injection framework, or ORM. Persistence is handled through a small repository built directly on pgx.

### Backend Layers

The backend is organized into lightweight logical layers rather than separate packages.

#### 1. Transport Layer

Defined primarily in `backend-cfb-manhole/handlers.go` and `backend-cfb-manhole/responses.go`.

Responsibilities:

- Accept HTTP requests.
- Enforce allowed methods.
- Parse route parameters.
- Return JSON responses.
- Apply an origin allowlist for browser CORS requests.

Supported endpoints:

- `GET /health`
- `GET /api/rankings`
- `GET /api/rankings/divisions`
- `GET /api/rankings/divisions/{divisionId}`
- `GET /api/players`
- `POST /api/players`
- `GET /api/players/{playerId}`
- `DELETE /api/players/{playerId}`
- `PUT /api/players/{playerId}/score`
- `PUT /api/players/{playerId}/teams`

The handler uses basic path matching with string comparisons and prefix checks instead of a dedicated routing library.

The CORS policy is loaded at startup from `CORS_ALLOWED_ORIGINS`. If unset, the backend allows only the local frontend dev origins on port `3000`. Disallowed preflight requests are rejected before they reach the write handlers.

#### 2. Service Layer

Defined in `backend-cfb-manhole/service.go`.

Responsibilities:

- Expose ranking-oriented operations to the handler.
- Expose player read/write operations to the handler.
- Group rankings by division.
- Re-rank entries within a division after partitioning the global list.

Primary methods:

- `GetRankings()`
- `GetDivisions()`
- `GetDivisionByID(id int)`
- `ListPlayers()`
- `GetPlayer(id)`
- `CreatePlayer(...)`
- `UpdatePlayerScore(id, score)`
- `ReplacePlayerTeams(id, teams)`

The service depends on a small repository interface:

```go
type rankingRepository interface {
  OrderedRankings(ctx context.Context) ([]playerEntry, error)
}
```

This remains the main architectural seam in the backend. It keeps transport logic decoupled from persistence.

#### 3. Repository and Domain Data Layer

Defined primarily in `backend-cfb-manhole/database.go`, with seed-source data in `backend-cfb-manhole/scoring.go`.

Responsibilities:

- Bootstrap the Postgres schema on startup.
- Seed an empty database with the current players, teams, divisions, and computed scores.
- Query globally ordered rankings from persisted scores.
- Persist player records and player/team relationships.

Important functions:

- `EnsureSchema()`
- `SeedInitialData()`
- `OrderedRankings()`
- `CreatePlayer()`
- `UpdatePlayerScore()`
- `ReplacePlayerTeams()`

### Scoring Pipeline

The backend computes initial seed scores from in-repo game results, then serves rankings from persisted player state using this sequence:

1. `SeedInitialData()` calls `calculateScores()` only when the `players` table is empty.
2. `calculateScores()` iterates through completed game results.
3. For each result, it checks which players picked the winner.
4. It assigns points based on the game context:
   - `1` point for a correct pick by default.
   - `2` points if the winner and loser are in the same conference.
   - `3` points if another player in the same division picked the loser.
5. Scores are aggregated per player and inserted into Postgres alongside each player's teams.
6. Ranking requests read persisted `current_score` values from Postgres.
7. Rankings are sorted by descending score and then ascending player name.
8. Global ranks are assigned.
9. Division endpoints regroup that ranked list by division and assign division-local ranks.

After seeding, score updates overwrite the latest state in the database rather than appending history.

### Backend Data Characteristics

- Player IDs are generated by the backend as app-level UUIDs.
- Player teams are stored in a separate `player_teams` table.
- Player scores are stored as current state in `players.current_score`.
- The in-memory scoring dataset is now seed data, not the runtime source of truth.

This approach keeps the code simple, but it also means:

- score changes overwrite the current value rather than preserving history,
- player teams are replaceable by API but expected to remain stable over a season,
- there is still no authentication or authorization layer around player writes,
- CORS now reduces accidental cross-site writes from arbitrary browser origins, but it is not a substitute for auth or CSRF protection,
- scaling beyond the current simple query set would likely benefit from migrations and stronger contract validation.

## Frontend Architecture

### Runtime Model

The frontend is a Next.js App Router application in `frontend-cfb-manhole/` using:

- React 19
- Next.js 15
- Chakra UI 3

The current UI is client-rendered for the home page leaderboard.

### Frontend Layers

#### 1. Application Shell

Defined in `frontend-cfb-manhole/src/app/layout.tsx`.

Responsibilities:

- Define root layout metadata.
- Load global CSS.
- Wrap the app in the shared UI provider.

#### 2. UI Provider Layer

Defined in `frontend-cfb-manhole/src/components/ui/provider.tsx` and related UI helpers.

Responsibilities:

- Initialize Chakra UI.
- Provide color mode support.

This layer centralizes UI framework setup so page components do not own provider wiring.

#### 3. Data Access Layer

Defined in `frontend-cfb-manhole/src/lib/api.ts`.

Responsibilities:

- Define frontend TypeScript types for ranking payloads.
- Call the backend using `fetch`.
- Validate success via HTTP status handling.
- Return parsed ranking data to page components.

The frontend expects the backend base URL from:

- `NEXT_PUBLIC_API_BASE_URL`, or
- defaults to `http://127.0.0.1:8000`

This is the main integration contract between the two applications.

#### 4. Page Layer

Defined in `frontend-cfb-manhole/src/app/page.tsx`.

Responsibilities:

- Trigger ranking fetch on mount.
- Manage `loading`, `error`, and `rankings` state.
- Render the leaderboard table.

The page currently consumes only the global rankings endpoint. Division-specific endpoints exist on the backend but are not yet surfaced in the UI.

## End-to-End Request Flow

### Global Rankings

1. A browser requests the home page from the Next.js app.
2. The home page mounts on the client.
3. `FetchRankings()` requests `GET /api/rankings` from the Go backend.
4. The backend handler calls `RankingService.GetRankings()`.
5. The ranking service reads ordered player rows from Postgres.
6. The backend returns JSON with `count` and `rankings`.
7. The frontend stores the parsed array in component state.
8. Chakra UI table components render the leaderboard.

### Division Rankings

Division ranking flow is available server-side but not currently wired into the frontend:

1. The client would call either `GET /api/rankings/divisions` or `GET /api/rankings/divisions/{id}`.
2. The handler delegates to `RankingService`.
3. The service loads persisted global ranking entries from Postgres.
4. Each division is re-ranked from `1..n` within that division.
5. JSON is returned to the caller.

### Player Writes

1. A client calls `POST /api/players` to create a player or `PUT /api/players/{id}/score` to overwrite a score.
2. The handler validates and decodes JSON.
3. `PlayerService` normalizes inputs and applies basic validation.
4. The Postgres repository writes the `players` and `player_teams` rows.
5. The updated player record is returned as JSON.

## Architectural Boundaries

### Stable Boundaries

- Backend and frontend are fully separate deployable units.
- The backend exposes a small JSON API contract.
- The service layer depends on an interface, not a concrete data source.
- The frontend keeps API access in a dedicated library module.
- Ranking reads and player writes share one repository, which keeps persistence concerns localized.

### Tight Couplings

- Initial seed data and seed-score logic are still tightly coupled in `scoring.go`.
- Backend route handling and application composition live in the same package.
- Frontend types mirror backend JSON informally rather than from generated schema.
- There is no shared contract generation such as OpenAPI or codegen.

## Current Design Tradeoffs

### Strengths

- Very low operational complexity.
- Easy local development.
- Persistent player/team storage.
- Clear separation between UI and API.
- Small service interface that can support future storage changes.

### Limitations

- No authentication or authorization.
- No score history; only latest score state is retained.
- No caching layer on top of database reads.
- No schema validation layer for requests or responses.
- Frontend currently uses only a subset of the backend capabilities.

## Evolution Paths

If this repository grows, the most natural next architectural steps are:

1. Add explicit SQL migrations rather than bootstrapping schema in application code.
2. Split backend packages by concern such as `transport`, `service`, and `repository`.
3. Introduce contract documentation or schema generation for the API.
4. Add integration tests against a disposable Postgres instance for repository behavior.
5. Extend the frontend to support division views and richer ranking exploration.
6. Consider server-side data fetching in Next.js if SEO, initial load performance, or caching become important.

## Summary

The repository uses a deliberately small architecture: a standalone Go API persists player scores and teams in Neon Postgres, and a standalone Next.js frontend fetches and displays those rankings. The main architectural seam is still the backend service-to-repository boundary, which now carries both leaderboard reads and player writes without forcing a larger framework redesign.