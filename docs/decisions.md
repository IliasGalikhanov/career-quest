# Engineering decisions

Career Quest helps a person choose a feasible learning activity for a career goal,
record the work they did, and distinguish that estimated progress from an assessment.
The local portfolio release uses independently authored fictional data throughout.

## Go modular monolith

One process serves the API and compiled React application. Pure domain functions
own admission, replay and ranking; HTTP and persistence stay outside those rules.
This makes the most important behavior testable without a browser or database.
Go's standard HTTP server, context cancellation and race detector fit this scope.
Microservices would add deployment and consistency costs without a separate scaling
or ownership requirement.

## SQLite and entity tables

The first version stored a complete bundle in one JSON column. It made atomic import
and backup simple, but every recommendation decoded every profile. The second schema
separates profiles, events, skills, role profiles and history into related tables.
Composite foreign keys constrain history to its dataset. Indexed stored columns
support employee, department, role and history selection. Unchanged entity rows are
not rewritten when a goal changes.

This is a **hybrid relational/document schema**, not a claim that every nested field
is normalized. Localized text, skill maps, prerequisites and frozen gains remain
JSON inside an entity row. Their semantic references are checked by the importer;
the database enforces dataset ownership and history's employee/event references.
Further normalization would become worthwhile for SQL reports across those nested
fields. There is no second PostgreSQL service just to enlarge the stack.

Measurements show the tradeoff: selective recommendations become much cheaper, while
complete snapshot reads and generic aggregate updates cost more. See
[the measured comparison](performance.md), including the regressions.

## Revisions and replay

Mutations read the current revision and change state inside a transaction. An old
revision returns 409, including for deletion. Read transactions keep metadata and
entity rows in one snapshot. The local process uses one database connection;
horizontal writes from multiple server processes are outside the supported model.

Completion freezes gain rules. Estimated levels are reconstructed from the assessment
and eligible completions, so retrying a completion cannot repeatedly add skills.
Date-only same-day history keeps recorded order. Unknown completion dates stay
uncertain. A work note records self-reported evidence; it is not graded automatically.

## Rules instead of an unvalidated model

There are no real learning outcomes, reference rankings or promotion labels to train
and evaluate against. Fixed weights make the system explainable and reproducible.
The scenario suite checks explicit product expectations and compares a shortest-step
baseline. Sensitivity tests expose when a different critical-skill weight changes
the answer. They do not establish which weight improves real employee outcomes.

## Frontend boundaries

`useWorkspace` owns asynchronous loading, mutations, errors and browser preferences.
`useSection` owns section URLs and Back/Forward behavior. Views render server results;
the ranking formula is not duplicated in TypeScript. Dataset forms have distinct
identity keys so replacing a dataset revision cannot leave duplicate form instances.
The browser suite includes delayed-read races, deletion, deep links and escaped notes.

## Operations appropriate to a local release

The process emits JSON request logs with server-generated request IDs, route patterns,
status, duration and response size. It does not log request bodies, query strings,
profile names or reflections. OpenAPI is generated from domain types, validated as a
document, reproduced in CI, and checked against live response schemas.

The Docker service stays on loopback, runs without root, and stores data in a named
volume. No user identities or HR access control are implied by navigation. Public
hosting and real employee data require a different security and governance design.
