# Verification

Recorded on **11 October 2026**. All application scenarios use the independently
authored synthetic dataset. No real employee records or measured learning outcomes
were used.

## Functional checks

| Check | Result |
| --- | --- |
| `go vet ./cmd/... ./internal/...` | Passed |
| `go test -race -cover ./cmd/... ./internal/...` | Passed, no race reports |
| React TypeScript and production build | Passed |
| ESLint with zero allowed warnings | Passed |
| Chromium browser suite | 13 scenarios passed, 40.1 seconds total on the local run |
| Accessibility | No axe violations for the trajectory, catalog, HR and data screens in light/dark themes using WCAG 2 A/AA and 2.1 AA tags |
| Keyboard | Skip link transfers focus to the main content |
| Mobile | 390 × 844 viewport, no horizontal page overflow; Russian/Kazakh switching and dark theme verified |
| Docker | Production image built; nonroot service reached healthy state |
| Container recreation | Dataset, revision, history, skill levels, and coverage preserved after `down` followed by `up` without deleting volumes |
| Clean GitHub clone | Built and started from a separate clone, with isolated project name, port and initially empty volume |
| Legacy upgrade | All 4 existing Docker datasets retained byte-equivalent JSON values and revisions after automatic schema migration |
| API contract | OpenAPI document valid; live response schemas and correlated error IDs passed |
| Recommendation evaluation | 8 authored expectations passed; baseline and weight sensitivity recorded |
| Repository exclusions | Private analysis, working brief, databases, imports and secrets are not tracked |

The browser suite covers delayed-read races, completion and reload, goal changes,
unavailable activities, import preview/commit/export, invalid import preservation,
HR filtering, localization, mobile layout, persisted time limits, keyboard access,
work notes, workspace rename/delete, URL Back/Forward, escaped text, and automated accessibility checks. The server suite additionally checks concurrent
writes, same-day replay order, duplicate completion, future sessions, ambiguous
history, CSV parsing, append conflicts, rollback, and SQLite reopening.

The main workflow was also exercised manually through the in-app browser against
the Docker service: start a recommendation, complete it, and observe architecture
increase from 2 to 3 while the assessed level remains 2.

The local checks used Windows on x86-64, Go 1.25.5, Node 24.11.0, Docker Engine 29.4.0
with Linux containers, and Playwright Chromium. Docker base images are pinned by
digest. The separate-clone check reused Docker's dependency cache; it verifies that
no untracked project files are required, not a completely cold download benchmark.

## Initial release latency (historical)

These measurements describe the original aggregate-storage release. For the current
entity schema, allocation comparison and concurrent HTTP check, see
[performance and limits](performance.md).

Measured with `python scripts/benchmark.py` against the Docker service on loopback.
Each endpoint received 3 warm-up calls and 30 measured requests, concurrency 1.
The dataset contained 12 profiles, 12 events, and 10 skills. Recommendation timing
includes the SQLite evidence write. Timings include HTTP transport and JSON decoding.

| Endpoint | Median | p95 | Maximum |
| --- | ---: | ---: | ---: |
| Dataset | 5.84 ms | 20.84 ms | 24.83 ms |
| Recommendation and evidence write | 10.35 ms | 28.81 ms | 34.28 ms |
| HR summary | 4.68 ms | 17.91 ms | 24.79 ms |

This is a small synthetic local benchmark. It does not establish performance at the
upload limits, under concurrent users, or over a remote network. Browser scenario
durations include multiple actions and are not individual interaction latency SLAs.
No recommendation accuracy, retention effect, or real-world engagement improvement
is claimed.

## Continuous verification

[GitHub Actions](https://github.com/IliasGalikhanov/career-quest/actions/workflows/ci.yml)
executes static checks, race-enabled tests, deterministic demo reproduction, frontend
build/lint, Chromium scenarios, Docker startup, and container-recreation persistence
checks on a clean Ubuntu checkout. Use the workflow attached to the relevant commit
for its exact result rather than assuming every historical commit was green.

## Remaining limits

- Automated accessibility coverage is limited to the tested screens and states; it is
  not a complete assistive-technology audit.
- Kazakh text is implemented but has not had an independent native-speaker review.
- In-browser self-reported completion is not a verified assessment or LMS integration.
- No authorization, public deployment, sustained multi-user soak test, or real-data evaluation
  is part of this local release.
- For self-paced history without completion dates, the actual assessment boundary
  cannot be recovered. The conservative default excludes uncertain gains.
- A project-wide source license has not been selected. The original third-party
  materials are not distributed by this repository.
