package demo

import (
	"careerquest/internal/domain"
	"testing"
)

func TestDemoContract(t *testing.T) {
	if issues := domain.Validate(Dataset()); len(issues) > 0 {
		t.Fatalf("invalid demo: %+v", issues)
	}
	a, ok := domain.Analyze(Dataset(), "person-01")
	if !ok || len(a.Recommendations) == 0 || a.Recommendations[0].Event.ID != "architecture-lab" {
		t.Fatalf("expected critical architecture gap to lead: %+v", a.Recommendations)
	}
}
