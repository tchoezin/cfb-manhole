# cfb-manhole

cfb-manhole contains a small backend scoring sandbox and a Next.js frontend for
presenting the experience. The repository now follows a Spec Kit constitution
that treats clean code, simple responsive UX, automated testing, and complete
documentation as release requirements.

## Repository Layout

- `backend-cfb-manhole/`: Python scripts and scoring logic.
- `frontend-cfb-manhole/`: Next.js 15 frontend using React 19 and Chakra UI.
- `.specify/`: Spec Kit templates, workflow memory, and automation metadata.

## Local Development

### Frontend

From `frontend-cfb-manhole/`:

```bash
npm install
npm run dev
```

The development server runs on `http://localhost:3000` by default.

### Backend

The backend directory currently contains standalone Python modules rather than a
packaged service. Any change that makes the backend user-runnable MUST document
its entry point, required environment, and example invocation in this README or
in a feature quickstart.

## Delivery Expectations

- Every change must keep public functions, shared components, and reusable
	modules documented close to the code.
- User-facing changes must define mobile and desktop behavior plus loading,
	empty, and error states.
- Features must include the right mix of unit, integration, and end-to-end
	tests unless a plan explicitly explains why a layer does not apply.
- Usage and setup documentation must be updated in the same change as the code.

## Working Agreement

The source of truth for delivery rules is `.specify/memory/constitution.md`.
Specs, plans, and tasks generated for this repo are expected to enforce those
rules rather than treating them as optional follow-up work.