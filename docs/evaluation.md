# Recommendation scenario evaluation

Run `go run ./cmd/evaluate`. The command prints JSON and returns nonzero if an
authored expectation fails. CI reproduces [the recorded results](evaluation.json).

| Scenario | Expected behavior |
| --- | --- |
| Critical skill gap | Architecture first, despite a shorter communication activity |
| Insufficient time | No recommendation |
| Missing prerequisite | No advanced activity until admission requirements are met |
| Change of role | A target role does not bypass current-role admission rules |
| Lead without a goal | No invented next grade |
| All requirements met | No activity advertised as reducing a nonexistent gap |
| No history | A useful first recommendation remains possible |
| Catalog lacks coverage | No fabricated activity |

All eight expectations pass. **These are authored regression scenarios**, deliberately
constructed to encode product policy. They are not independent relevance labels,
an unbiased benchmark, a held-out test set, or an accuracy estimate on real people.

The baseline selects the shortest feasible activity with a positive target gain.
In the critical-gap scenario it selects `story-circle`, while the weighted rule
selects `architecture-lab`. They agree on the empty-result scenarios because both
respect the same admission rules; those cases do not demonstrate ranking superiority.

The report recalculates the ranking at critical weights 1, 2 and 3. At weight 1,
communication comes first in the critical-gap example; at weights 2 and 3,
architecture does. The default 2 expresses a product preference, not an empirically
optimal parameter. History remains a bounded preference signal, not an employee score.

Separate invariant tests shuffle catalog order, fuzz gain bounds, verify frozen
completion effects, and reject invalid imports. A future real-world evaluation would
need permissioned data, independently collected relevance judgments, explicit outcome
definitions and appropriate separation between parameter selection and evaluation.
