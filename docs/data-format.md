# Data contract

All imported information must be synthetic and usable under your own rights. The
example in `examples/demo.json` is self-contained. Import preview never writes data;
the complete package is parsed and validated again before commit.

## Portable bundle

One UTF-8 JSON object contains:

| Field | Meaning |
| --- | --- |
| `meta` | `dataset`, `version`, `as_of_date` (YYYY-MM-DD), `synthetic: true` |
| `skills` | Skill IDs, names, hard/soft type, category, optional translated names |
| `role_profiles` | Role/grade pairs, required skill levels, critical skill IDs |
| `employees` | Fictional profiles, skill assessments, last review date, optional goals |
| `events` | Audience, prerequisites, gains/caps, duration, format, future sessions |
| `activity_history` | Participation attempts and outcomes |
| `settings` | Calculation date, maximum activity hours, format, planning horizon, date assumption |

Use the example rather than guessing nested field names. `Go` types in
`internal/domain/types.go` are the canonical schema. Unknown skill/reference IDs,
duplicate IDs, invalid dates, noninteger levels, and incompatible status/percentage
combinations are rejected with paths. Up to 100 issues are returned per request.

Grades are ordered `Junior`, `Middle`, `Senior`, `Lead`. Skill levels are integers
0–5. Missing skills count as zero but are marked as unassessed. Goal and current
role/grade pairs must exist in `role_profiles`. All related files use the same snapshot
date; the profile review cannot precede hire or follow the snapshot.

## Four-file import

Upload the following exact filenames together:

- `employees.json`: `{ "meta": { ... }, "employees": [ ... ] }`
- `skills.json`: `{ "meta": { ... }, "skills": [ ... ], "role_profiles": [ ... ] }`
- `events.json`: `{ "meta": { ... }, "events": [ ... ] }`
- `activity_history.csv`: the header below and zero or more rows.

```csv
record_id,employee_id,event_id,date,due_date,status,completion_pct,score,feedback_rating,assigned_by
```

Quoted CSV fields and UTF-8 BOM are accepted. `score` and `feedback_rating` may be empty;
other numeric values must be valid integers, never silently replaced with zero.
An optional `completed_at` column supplies an actual completion date. `date` is the
session date for scheduled activities, enrollment date for self-paced work.

Statuses: `completed`, `in_progress`, `dropped`, `declined`, `no_show`, `overdue`.
Initiators: `self`, `manager`, `hr`. `completed` requires 100%; declined and no-show
require 0%. Other incomplete attempts cannot use 100%.

The synthetic checkbox is required even if metadata states that the data are fictional.
For a four-file package whose legacy metadata omits `synthetic`, this explicit
declaration supplies the flag. For portable bundles, `meta.synthetic` must also be true.

## Append profiles and history

Select append mode, then upload only `employees.json` and `activity_history.csv`.
The snapshot date must match the selected dataset. Employee and record IDs must be
new; history may reference existing or newly added employees and existing events.
Existing profiles, goals, settings, activities, and history are preserved. Conflicting
IDs fail the entire import. Replacing a dataset is performed by importing a new full
bundle, not by overwriting records invisibly.

## Extended fields and backup

`names`, `titles`, and `descriptions` are optional maps with `en`, `ru`, and `kk` keys.
Uploaded plain text falls back to its source language. `repeatable` is an explicit
event property; no event ID receives a hidden hardcoded exception.

New completed records include `completed_at`, `session_date` where applicable, and
`applied_gains`. Frozen gains preserve the original rule even if a future catalog
changes. Exported bundles retain these fields and can be imported as new datasets.
The export is a portable domain backup, not a SQLite file or a copy of the evidence log.

CSV and bundle files are never written to an upload folder by the application. A valid
commit stores the parsed state in SQLite. Raw request bodies are not logged. Database
files, user imports, private notes, and secrets are excluded from Git and the Docker
build context.

## Exercise content and work notes

Events may include `deliverable` (a localized string map) and `criteria` (a list of
localized string maps). Imported events without these fields remain usable; the UI
omits the exercise-details panel. The independent demo supplies both in EN/RU/KK.
Existing datasets retain their original catalogs during migration; create a new demo
to see the new exercises rather than silently rewriting an old catalog.

Participation can include `reflection`, up to 2,000 Unicode characters. Completion
accepts an optional note, stores plain text and displays it escaped. Export/import
preserves it. It is self-reported evidence, not a verified proficiency assessment.
Repeating an already-completed transition does not modify its note or award new gains.

Dataset names support 1–120 Unicode characters. Rename changes both the displayed
name and `meta.dataset`; it does not change IDs or reference relationships.
