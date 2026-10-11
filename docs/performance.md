# Performance and limits

Recorded on 11 October 2026. All records are independently authored synthetic fixtures.
No timings are learning-outcome or recommendation-quality metrics.

## Storage comparison

Windows x86-64, AMD Ryzen 5 5600H, Go 1.25.5, local SQLite on the same host.
Three benchmark runs, 10 operations per run; table shows the median of the three
reported means. Allocation is **bytes allocated per operation**, not peak RSS.
Twelve catalog events, ten skills and three history rows per profile. Runs are
sequential, concurrency 1, without the race detector. Wall-clock timings are sensitive
to the filesystem, antivirus, scheduling and other activity on the machine.

The baseline uses the storage implementation and migration from commit `2184118`,
but the **same current domain types, exercises, scaled fixture and benchmark** as the
new implementation. This isolates the storage change rather than comparing different
demo content. Reproduce with `python scripts/compare_storage.py` from a full clone.
Raw outputs: [before](benchmarks/before.txt), [after](benchmarks/after.txt).

| Profiles | Operation | Before ms | After ms | Before allocated MB | After allocated MB |
| ---: | --- | ---: | ---: | ---: | ---: |
| 12 | goal_update | 3.19 | 6.49 | 0.38 | 0.55 |
| 12 | read | 1.15 | 1.86 | 0.24 | 0.28 |
| 12 | recommendation_with_evidence | 3.33 | 4.43 | 0.36 | 0.30 |
| 500 | goal_update | 30.26 | 37.02 | 6.42 | 9.18 |
| 500 | read | 17.03 | 24.78 | 3.82 | 4.83 |
| 500 | recommendation_with_evidence | 17.57 | 3.25 | 3.98 | 0.30 |
| 2000 | goal_update | 130.22 | 151.99 | 30.79 | 38.04 |
| 2000 | read | 60.97 | 113.75 | 17.19 | 21.17 |
| 2000 | recommendation_with_evidence | 64.97 | 3.06 | 17.37 | 0.30 |

Selective recommendation reads improve substantially at 2,000 profiles because the
server loads only one profile and its history. The composite employee/history-order
index avoids scanning all history to preserve replay order. Evidence persistence is
included in this operation.

Complete reads regress because SQLite returns and decodes many separate rows. Generic
goal updates still assemble a complete bundle, capture immutable entity snapshots,
and compare rows; this also costs more than rewriting a single JSON blob at these
sizes, despite avoiding unchanged database writes. The schema is a data-integrity and
selective-read improvement, **not a universal speedup**. Targeted write commands and
paged snapshots are sensible future changes if larger workspaces become a requirement.

## HTTP under bounded concurrency

Docker Desktop with Linux containers on the same Windows host. Standard-library
Python client, loopback HTTP, 40 measured requests per endpoint, four workers, one
warm-up request. Each endpoint is exercised separately; these are not independent
users or a sustained production soak. Timings include HTTP, JSON transfer/decoding,
and SQLite evidence writes for recommendations. Import is one measurement per size.

| Profiles | History | Import ms | Recommendation p95 ms | HR p95 ms | Dataset p95 ms | Errors |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 12 | 36 | 41.50 | 42.00 | 31.61 | 33.64 | 0 |
| 500 | 1500 | 612.37 | 33.85 | 264.52 | 272.33 | 0 |
| 2000 | 6000 | 3045.53 | 38.07 | 1313.35 | 1099.34 | 0 |

[Full response measurements](benchmarks/http-load.json) include maxima. The 2,000-profile
HR maximum was **2,172.61 ms**, exceeding two seconds in this test. Import of that set
took **3,045.53 ms**. There is no claim of a universal two-second interaction bound.

Reproduce against a running Docker service with `python scripts/load_check.py`.
The script imports its own temporary fixtures and deletes only those fixtures after
each size. It writes a fresh report. Run it separately from other benchmarks.

The largest tested fixture has 2,000 profiles, 6,000 history rows and a roughly 2.1 MB
input. Upload limits of 5 MiB / 5,000 profiles / 100,000 history records are validation
ceilings, not demonstrated capacity. Catalog growth, sustained contention, slow disks,
remote networks and multi-process writers have not been characterized.
