<!--
Sync Impact Report
Version change: template -> 1.0.0
Modified principles:
- Placeholder principle 1 -> I. Clean Code Is the Default
- Placeholder principle 2 -> II. UX Must Stay Simple and Responsive
- Placeholder principle 3 -> III. Testing Is Mandatory Across Layers
- Placeholder principle 4 -> IV. Documentation Ships With the Change
- Placeholder principle 5 -> V. Prefer Simplicity, Justify Complexity
Added sections:
- Engineering Standards
- Delivery Workflow
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/spec-template.md
- ✅ .specify/templates/tasks-template.md
- ✅ README.md
- ✅ frontend-cfb-manhole/README.md
Follow-up TODOs:
- None
-->
# cfb-manhole Constitution

## Core Principles

### I. Clean Code Is the Default
Every production change MUST favor small, readable units with obvious names,
single-purpose functions, and explicit data flow. Public modules, exported
components, and reusable functions MUST include concise documentation that
states purpose, inputs, outputs, and notable side effects. Dead code,
speculative abstractions, and duplicated logic MUST be removed or justified in
the implementation plan because maintainability depends on code being easy to
understand at a glance.

### II. UX Must Stay Simple and Responsive
User-facing behavior MUST optimize for the shortest path to the primary task.
Interfaces MUST define expected behavior for mobile and desktop layouts, loading
states, empty states, and error states before implementation begins. Responsive
design and accessibility are release gates, not polish items, because a simple
experience fails if it is confusing, brittle, or unusable on common screen
sizes.

### III. Testing Is Mandatory Across Layers
Every feature and bug fix MUST include the minimum effective mix of automated
tests: unit tests for core logic, integration tests for boundaries between
modules or services, and end-to-end coverage for critical user journeys. When a
layer is not applicable, the plan MUST state why. Tests MUST be written before
or alongside implementation and MUST fail meaningfully without the change,
because unverified behavior is not accepted work.

### IV. Documentation Ships With the Change
Every change that affects behavior, setup, or operations MUST update the
relevant documentation in the same delivery. This includes user-facing usage
guidance, developer setup notes, quickstarts, and inline function or component
documentation where readers need immediate context. A feature is incomplete if
future contributors or users cannot tell how to run it, use it, or extend it
safely.

### V. Prefer Simplicity, Justify Complexity
Designs MUST begin with the simplest approach that can satisfy current
requirements. New dependencies, abstractions, or architectural layers require an
explicit tradeoff statement in the plan and a documented reason simpler options
were rejected. Complexity is acceptable only when it measurably improves
correctness, usability, or delivery risk.

## Engineering Standards

- Plans MUST identify how the backend and frontend responsibilities are split
	when a change touches both surfaces.
- Linting, type checks where available, and automated tests MUST pass before a
	change is considered ready for review.
- UX work MUST include responsive acceptance criteria for small and large
	screens and MUST avoid hidden critical actions or ambiguous labels.
- Public functions, utilities, and shared UI building blocks MUST include brief
	descriptive documentation close to the code.

## Delivery Workflow

- Specifications MUST include testable user stories, explicit edge cases,
	documentation impact, and non-functional requirements for responsiveness and
	clarity.
- Implementation plans MUST complete a Constitution Check covering code
	simplicity, responsive UX scope, required test layers, and documentation
	updates.
- Tasks MUST schedule documentation work and required unit, integration, and
	end-to-end tests as first-class deliverables rather than optional follow-up.
- Reviews MUST reject changes that add undocumented behavior, omit required test
	coverage, or ship user flows that are not responsive.

## Governance

This constitution overrides conflicting local habits or template defaults. Any
amendment MUST be made in the same change set as the template or documentation
updates needed to keep the workflow aligned. Versioning follows semantic
versioning for governance: MAJOR for incompatible principle changes or removals,
MINOR for new principles or materially stronger requirements, and PATCH for
clarifications that do not change obligations. Every review, plan, and task list
MUST include an explicit compliance check against this constitution.

**Version**: 1.0.0 | **Ratified**: 2026-04-19 | **Last Amended**: 2026-04-19
