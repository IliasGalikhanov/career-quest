package domain_test

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"errors"
	"testing"
)

func TestFutureSessionAndRepeatableCompletion(t *testing.T) {
	b := demo.Dataset()
	b.Settings.Date = "2026-02-03"
	if err := domain.Start(&b, "person-01", "story-circle", "new"); err != nil {
		t.Fatal(err)
	}
	if err := domain.Finish(&b, "new", "completed"); !errors.Is(err, domain.ErrFutureSession) {
		t.Fatalf("future session allowed: %v", err)
	}
	b.Settings.Date = "2026-02-16"
	if err := domain.Finish(&b, "new", "completed"); err != nil {
		t.Fatal(err)
	}
	if err := domain.Start(&b, "person-01", "story-circle", "duplicate"); !errors.Is(err, domain.ErrIneligible) {
		t.Fatalf("same session repeated: %v", err)
	}
	b.Settings.Date = "2026-02-17"
	if err := domain.Start(&b, "person-01", "story-circle", "next"); err != nil {
		t.Fatal("next session should be available", err)
	}
	if issues := domain.Validate(b); len(issues) > 0 {
		t.Fatal(issues)
	}
}

func TestValidationRejectsFutureHistoryAndDuplicateSessions(t *testing.T) {
	b := demo.Dataset()
	b.History = append(b.History, domain.Activity{ID: "future", EmployeeID: "person-01", EventID: "api-clinic", Date: "2026-02-02", CompletedAt: "2027-01-01", Status: "completed", Completion: 100, AssignedBy: "self"})
	if len(domain.Validate(b)) == 0 {
		t.Fatal("future completion accepted")
	}
	b = demo.Dataset()
	h := domain.Activity{ID: "repeat-a", EmployeeID: "person-01", EventID: "story-circle", Date: "2026-02-02", CompletedAt: "2026-02-02", SessionDate: "2026-02-02", Status: "completed", Completion: 100, AssignedBy: "self"}
	b.History = append(b.History, h)
	h.ID = "repeat-b"
	b.History = append(b.History, h)
	if len(domain.Validate(b)) == 0 {
		t.Fatal("duplicate same-session completion accepted")
	}
}
