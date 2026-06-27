# cfb-manhole

CFB Manhole is a small two-application project for tracking college football pick'em rankings.

- `backend-cfb-manhole/` is a Go API backed by Neon Postgres for player scores and player teams.
- `frontend-cfb-manhole/` is a Next.js UI that fetches and displays the leaderboard.

## Repository Layout

- `backend-cfb-manhole/`: HTTP API, scoring logic, and ranking data.
- `frontend-cfb-manhole/`: App Router frontend built with React and Chakra UI.
- `ARCHITECTURE.md`: repository architecture and request-flow breakdown.

## Quick Start

Run the backend first:

```bash
cd backend-cfb-manhole
export NEON_DATABASE_URL="postgres://..."
go run .
```

Then run the frontend in a second terminal:

```bash
cd frontend-cfb-manhole
npm install
npm run dev
```

Open `http://localhost:3000` in the browser. The frontend calls the backend at `http://127.0.0.1:8000` by default.

## Configuration

The backend requires `NEON_DATABASE_URL` or `DATABASE_URL` and bootstraps its schema on startup.

The frontend can target a different backend by setting `NEXT_PUBLIC_API_BASE_URL` before starting the dev server.

## Documentation

- See `ARCHITECTURE.md` for the system design.
- See `backend-cfb-manhole/README.md` for backend-specific details.
- See `frontend-cfb-manhole/README.md` for frontend-specific details.