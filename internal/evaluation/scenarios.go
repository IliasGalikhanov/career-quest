// Package evaluation contains authored product expectations, not measured outcomes.
package evaluation

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"math"
	"sort"
)

type Case struct {
	Name, Rationale, WantTop string
	Prepare                  func(*domain.Bundle)
}

func Cases() []Case {
	return []Case{
		{"critical_gap", "Prefer the critical architecture gap over a shorter communication activity.", "architecture-lab", func(*domain.Bundle) {}},
		{"time_budget", "Do not propose work that exceeds the available effort.", "", func(b *domain.Bundle) { b.Settings.MaxHours = .5 }},
		{"missing_prerequisite", "Do not offer advanced architecture before the prerequisite level.", "", func(b *domain.Bundle) { b.Events = []domain.Event{b.Events[10]} }},
		{"role_change_admission", "A new goal does not grant admission reserved for the target role.", "", func(b *domain.Bundle) {
			b.Employees[0].Goal = &domain.Goal{Role: "Product Designer", Grade: "Senior"}
			b.Events = []domain.Event{b.Events[3]}
		}},
		{"no_goal_at_lead", "Do not invent a promotion after Lead.", "", func(b *domain.Bundle) { b.Employees[0].Grade = "Lead"; b.Employees[0].Goal = nil }},
		{"requirements_met", "Do not recommend courses when all target requirements are met.", "", func(b *domain.Bundle) {
			for k := range b.Employees[0].Skills {
				b.Employees[0].Skills[k] = 5
			}
		}},
		{"no_history", "Sparse history must not prevent a useful first recommendation.", "architecture-lab", func(b *domain.Bundle) { b.History = nil }},
		{"catalog_gap", "Return no step when the catalog cannot improve the target skills.", "", func(b *domain.Bundle) {
			b.Events = []domain.Event{b.Events[9]}
			b.Employees[0].Skills["communication"] = 3
		}},
	}
}

type Result struct {
	Name             string            `json:"name"`
	Rationale        string            `json:"rationale"`
	Expected         string            `json:"expected_top"`
	Actual           string            `json:"actual_top"`
	Passed           bool              `json:"passed"`
	ShortestFeasible string            `json:"shortest_feasible_baseline"`
	Sensitivity      map[string]string `json:"critical_weight_sensitivity"`
}

func top(c []domain.Candidate) string {
	if len(c) == 0 {
		return ""
	}
	return c[0].Event.ID
}
func Run() []Result {
	out := []Result{}
	for _, tc := range Cases() {
		b := demo.Dataset()
		tc.Prepare(&b)
		a, _ := domain.Analyze(b, b.Employees[0].ID)
		eligible := []domain.Candidate{}
		for _, c := range a.Catalog {
			if len(c.Blocked) == 0 && c.Benefit > 0 {
				eligible = append(eligible, c)
			}
		}
		sort.Slice(eligible, func(i, j int) bool {
			if eligible[i].Event.Hours == eligible[j].Event.Hours {
				return eligible[i].Event.ID < eligible[j].Event.ID
			}
			return eligible[i].Event.Hours < eligible[j].Event.Hours
		})
		r := Result{Name: tc.Name, Rationale: tc.Rationale, Expected: tc.WantTop, Actual: top(a.Recommendations), ShortestFeasible: top(eligible), Sensitivity: map[string]string{}}
		r.Passed = r.Expected == r.Actual
		for label, weight := range map[string]float64{"1": 1, "2": 2, "3": 3} {
			variants := append([]domain.Candidate{}, eligible...)
			for i := range variants {
				benefit := float64(0)
				for _, g := range variants[i].Improvements {
					w := 1.0
					if g.Critical {
						w = weight
					}
					benefit += float64(min(g.Required-g.From, g.To-g.From)) * w
				}
				variants[i].Score = benefit * variants[i].HistoryFactor / math.Sqrt(math.Max(1, variants[i].Event.Hours))
			}
			sort.Slice(variants, func(i, j int) bool {
				if variants[i].Score == variants[j].Score {
					if variants[i].CriticalGain != variants[j].CriticalGain {
						return variants[i].CriticalGain > variants[j].CriticalGain
					}
					if variants[i].Event.Hours != variants[j].Event.Hours {
						return variants[i].Event.Hours < variants[j].Event.Hours
					}
					return variants[i].Event.ID < variants[j].Event.ID
				}
				return variants[i].Score > variants[j].Score
			})
			r.Sensitivity[label] = top(variants)
		}
		out = append(out, r)
	}
	return out
}
