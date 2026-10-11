# Architecture and ranking

Career Quest is a modular monolith. One Go process serves the React production build
and a same-origin JSON API. SQLite stores local state; no external services run at
request time.

```text
React views → HTTP API → domain rules / importer → SQLite store
                             ↑
                 independent synthetic generator
```

## Boundaries

| Module | Responsibility |
| --- | --- |
| `internal/domain` | Portable types, validation, skill replay, ranking, participation transitions |
| `internal/importer` | JSON and CSV parsing, append semantics, validation paths |
| `internal/store` | SQL migrations, aggregate revisions, transactions, evidence retention |
| `internal/httpapi` | HTTP routing, request limits, same-origin write checks, API errors, static assets |
| `internal/demo` | Reproducible fictional scenarios without external data |
| `web/src` | Typed API client, localized views, theme and browser preferences |

Domain functions operate on values and do not depend on SQL or HTTP. React displays
server-calculated results; it never reimplements the recommendation score.

## Storage and concurrency

A dataset is a versioned aggregate serialized as JSON inside SQLite. Catalogs,
assessments, goals, participation, and settings remain distinct fields of that
aggregate. This makes whole-package validation and portable backup straightforward.
At the current local scale, atomicity and simplicity matter more than relational
reporting across millions of records. The tradeoff is that updates rewrite the
aggregate and analytics run in memory. A larger multi-user system should normalize
the records and introduce indexed queries and authorization.

SQLite uses WAL, foreign keys, a busy timeout, and one connection per process.
Every mutation checks the client's dataset revision inside a transaction. Concurrent
edits cannot silently overwrite each other: a stale client receives HTTP 409 and must
refresh. Migrations are embedded SQL files, recorded only after successful commit.

Completion stores its date and a frozen copy of the activity's gain rules. Skill
levels are derived by replay, not by repeatedly incrementing the profile. Retrying
completion cannot award a second gain. Imported records lacking frozen gains use
their dataset catalog; append mode cannot modify existing catalog entries.

Recommendation evidence is stored separately with the algorithm version, dataset
revision, calculation settings, assessed profile, effective levels, candidates, and
score factors. The latest 500 runs per dataset are retained and downloadable. A
dataset backup includes domain state; historical recommendation evidence has a
separate export and is not restored by the dataset importer.

## Recommendation algorithm

1. Use the chosen target role and grade. If no goal exists, suggest the next grade
   only when that role profile is available. Lead without a goal receives no invented
   next grade.
2. Replay eligible completions strictly after the last assessment and no later than
   the calculation date. Missing skills are treated as zero and marked unassessed.
3. Exclude mandatory activities, incompatible current roles/grades, missing
   prerequisites, unavailable sessions, completed nonrepeatable activities, open
   attempts, and activities outside time/format constraints.
4. Calculate the skill gap reduction, using weight 2 for critical target skills and
   1 for other target skills.
5. Apply a bounded history preference and an effort penalty, then return up to three
   alternatives with positive benefit.

```text
next_level = max(level, min(5, activity.max_level, level + activity.gain))
benefit = Σ min(max(0, requirement - level), next_level - level) × critical_weight
history_factor = 0.75 + 0.5 × (completed + 2) / (completed + unfinished + 4)
score = benefit × history_factor / sqrt(max(1, duration_hours))
```

The history window is 180 days. Only voluntary activities of the same type and format
contribute. Unfinished means dropped, declined, or no-show; ongoing work is not a
failure. No relevant history gives factor 1. Equal scores sort by critical gap
reduction, then shorter duration, then activity ID. These are fixed, documented
heuristics, not a fitted probability model.

The UI explains the goal, concrete gap and expected gain, and participation evidence
(including its absence). A missing recommendation is a valid outcome. The library
shows why events were filtered out. Users may manually start a feasible voluntary
activity that does not close a current target gap; it is not advertised as a
recommendation.

## Dates and progress

`completed_at` is authoritative when present. A dated scheduled session can use its
session record date. A self-paced row without `completed_at` cannot reliably locate
completion relative to the last review and does not increase the estimate by default.
An explicit setting can treat its enrollment date as completion; the UI warns that
this may overestimate progress. Review-day completions are treated as already included
in the assessment because no within-day ordering is available.

For multiple completions on the same day, replay preserves their recorded order.
Imported dates cannot be later than the calculation date, and duplicate completions
of the same session are rejected. A calculation date before the snapshot is invalid.

Coverage is `Σ min(level, requirement) / Σ requirement` over the target profile.
Zero total requirements produce no percentage. Critical gaps are shown separately.
Neither coverage nor course completion changes the employee's grade or confirmed
assessment automatically.

Scheduled participation cannot be marked complete before its selected session.
Repeatable self-paced activities can complete at most once per calculation day;
repeatable scheduled activities can complete once per selected session. The first
future session in the planning window is selected automatically.

## Operating limits

- Local single-user access only. No identities, roles, consent workflow, or tenant isolation.
- Imports require a synthetic-data declaration; this is not a personal-data detector.
- Maximum request upload: 5 MiB, including multipart overhead; 5,000 profiles, 500
  skills, 2,000 events, 100,000 history records. These are validation ceilings, not
  demonstrated performance guarantees at maximum size.
- Time limit is total activity effort. Weekly schedules and calendar conflict checks
  cannot be inferred from total duration alone.
- A role change can have no feasible training because admission still uses the current
  role and grade. Missing catalog coverage is not replaced with fabricated events.
- Event text from uploads stays in the source language; only the owned demo has
  three-language skill and activity descriptions.
- No automatic promotion, employee ranking, attrition prediction, or quality claims
  based on synthetic outcomes.
- SQLite snapshots suit a local workspace. This is not a horizontally scalable service.

## Runtime protections

Compose binds only to loopback. The container runs as UID 10001, drops capabilities,
has a read-only root filesystem and a writable database volume. Browser writes must
be same-origin, request sizes are bounded, and SQL uses parameters. React escapes
uploaded text. Responses set content-type, framing, and CSP protections. These measures
do not substitute for authentication if the deployment model changes.
