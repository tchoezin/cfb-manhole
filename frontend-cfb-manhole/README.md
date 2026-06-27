# Frontend CFB Manhole UI

This frontend renders the CFB Manhole leaderboard using Next.js, React, and Chakra UI.

## Requirements

- Node.js 20+
- A running backend API, defaulting to `http://127.0.0.1:8000`

## Run

Install dependencies and start the development server:

```bash
npm install
npm run dev
```

Open `http://localhost:3000` in the browser.

## Configuration

Set `NEXT_PUBLIC_API_BASE_URL` to point at a different backend before starting the app:

```bash
NEXT_PUBLIC_API_BASE_URL=http://127.0.0.1:8000 npm run dev
```

If unset, the app uses `http://127.0.0.1:8000`.

## Current Behavior

- The home page fetches global rankings from `GET /api/rankings`.
- Loading and error states are handled in the page component.
- Division-specific backend endpoints exist but are not yet exposed in the UI.

## Key Files

- `src/app/layout.tsx`: root layout and provider wiring.
- `src/app/page.tsx`: leaderboard page.
- `src/lib/api.ts`: backend API client.
- `src/components/ui/`: Chakra UI provider and UI helpers.

## Tech Stack

- Next.js 15 App Router
- React 19
- Chakra UI 3
- TypeScript
