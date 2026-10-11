package domain_test

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"testing"
)

func TestCompletionIsConservativeAndDoesNotReduceLevel(t *testing.T) {
	b := demo.Dataset()
	e := b.Employees[0]
	e.Skills["architecture"] = 4
	b.History = append(b.History, domain.Activity{ID: "ambiguous", EmployeeID: e.ID, EventID: "architecture-lab", Date: "2026-01-20", Status: "completed", Completion: 100, AssignedBy: "self"})
	levels, uncertain := domain.EffectiveLevels(b, e)
	if levels["architecture"] != 4 || uncertain != 1 {
		t.Fatal("ambiguous self-paced date changed assessment")
	}
	if got := domain.ApplyGain(5, domain.Gain{Gain: 1, MaxLevel: 3}); got != 5 {
		t.Fatalf("level reduced to %d", got)
	}
	b.History[len(b.History)-1].CompletedAt = "2026-01-21"
	levels, uncertain = domain.EffectiveLevels(b, e)
	if levels["architecture"] != 4 || uncertain != 0 {
		t.Fatal("cap or known completion handled incorrectly")
	}
}

func TestReplayUsesFrozenGainsAndReviewBoundary(t *testing.T) {
	b := demo.Dataset()
	e := b.Employees[0]
	h := domain.Activity{ID: "done", EmployeeID: e.ID, EventID: "architecture-lab", Date: "2026-01-10", CompletedAt: "2026-01-10", Status: "completed", Completion: 100, AssignedBy: "self", Gains: []domain.Gain{{SkillID: "architecture", Gain: 1, MaxLevel: 5}}}
	b.History = []domain.Activity{h}
	levels, _ := domain.EffectiveLevels(b, e)
	if levels["architecture"] != 2 {
		t.Fatal("review-day completion double counted")
	}
	b.History[0].CompletedAt = "2026-01-11"
	b.Events[0].Gains[0].Gain = 3
	levels, _ = domain.EffectiveLevels(b, e)
	if levels["architecture"] != 3 {
		t.Fatal("catalog change rewrote a frozen completion")
	}
}

func TestSameDayReplayPreservesRecordedOrder(t *testing.T) {
	b := demo.Dataset()
	b.History = []domain.Activity{
		{ID: "z-first", EmployeeID: "person-01", EventID: "architecture-lab", Date: "2026-01-20", CompletedAt: "2026-01-20", Status: "completed", Gains: []domain.Gain{{SkillID: "architecture", Gain: 1, MaxLevel: 3}}},
		{ID: "a-second", EmployeeID: "person-01", EventID: "architecture-deep-dive", Date: "2026-01-20", CompletedAt: "2026-01-20", Status: "completed", Gains: []domain.Gain{{SkillID: "architecture", Gain: 1, MaxLevel: 5}}},
	}
	levels, _ := domain.EffectiveLevels(b, b.Employees[0])
	if levels["architecture"] != 4 {
		t.Fatalf("ID order changed gain replay: %v", levels)
	}
}

func TestEligibilityAndMissingGoal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*domain.Bundle)
		want   int
	}{
		{"no time", func(b *domain.Bundle) { b.Settings.MaxHours = 0.5 }, 0},
		{"no sessions", func(b *domain.Bundle) { b.Settings.Date = "2027-01-01"; b.Settings.Format = "online" }, 0},
		{"lead without goal", func(b *domain.Bundle) { b.Employees[0].Grade = "Lead"; b.Employees[0].Goal = nil }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := demo.Dataset()
			tc.change(&b)
			a, _ := domain.Analyze(b, b.Employees[0].ID)
			if len(a.Recommendations) != tc.want {
				t.Fatalf("unexpected recommendations %+v", a.Recommendations)
			}
		})
	}
	b := demo.Dataset()
	a, _ := domain.Analyze(b, b.Employees[0].ID)
	for _, c := range a.Recommendations {
		if c.Event.Mandatory || len(c.Blocked) > 0 || c.Benefit <= 0 {
			t.Fatalf("ineligible recommendation: %+v", c)
		}
	}
}

func TestValidationReportsBadReferencesAndDuplicateCompletions(t *testing.T) {
	b := demo.Dataset()
	b.Employees[0].Skills["not-a-skill"] = 8
	b.History = append(b.History, b.History[3])
	issues := domain.Validate(b)
	codes := map[string]bool{}
	for _, issue := range issues {
		codes[issue.Code] = true
	}
	for _, code := range []string{"reference", "range", "duplicate", "duplicate_completion"} {
		if !codes[code] {
			t.Errorf("missing %s issue", code)
		}
	}
}
