# Frontend

This package contains the Next.js frontend for cfb-manhole. It is expected to
follow the repository constitution: clean component boundaries, clear function
documentation where code is reused, responsive behavior across common viewport
sizes, and automated coverage for critical user journeys.

## Commands

```bash
npm install
npm run dev
npm run build
npm run lint
```

`npm run dev` starts the app locally on `http://localhost:3000`.

## Development Rules

- Keep primary user flows obvious and low-friction on both mobile and desktop.
- Define and implement loading, empty, validation, and error states as part of
	the feature, not as later polish.
- Document exported utilities, shared UI primitives, and other reusable code so
	their purpose is clear on first read.
- Update the root README, feature quickstarts, or other usage docs whenever the
	frontend changes how the app is started or used.

## Testing Expectation

Frontend changes should come with the right level of automated coverage for the
risk involved. That usually means unit coverage for isolated logic, integration
tests for component behavior and boundaries, and end-to-end coverage for the
main user path when the change affects the full experience.
