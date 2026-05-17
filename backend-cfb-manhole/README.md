# Backend CFB Manhole API (Go)

A minimal read-only Go API for leaderboard data.

## Requirements

- Go 1.22+

## Run

From this folder:

```bash
go run .
```

Server starts at `http://127.0.0.1:8000`.

## Build

```bash
go build ./...
```

## Endpoints

- `GET /health`
  - Returns service status.
- `GET /api/rankings`
  - Returns players ordered by descending score.
- `GET /api/rankings/divisions`
  - Returns all divisions with division-specific rankings.
- `GET /api/rankings/division/{divisionName}`
  - Returns a single division leaderboard.
  - Example: `/api/rankings/division/Division%201`

## CORS

This API includes permissive CORS headers for frontend wiring:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

### Example response

```json
{
  "count": 12,
  "rankings": [
    {
      "rank": 1,
      "player": "JR",
      "score": 7,
      "division": "Division 1"
    }
  ]
}
```
