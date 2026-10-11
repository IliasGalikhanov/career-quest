package importer

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"encoding/json"
	"testing"
)

func marshal(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func fourFiles(t *testing.T) map[string][]byte {
	b := demo.Dataset()
	return map[string][]byte{
		"employees.json":       marshal(t, map[string]any{"meta": b.Meta, "employees": b.Employees}),
		"skills.json":          marshal(t, map[string]any{"meta": b.Meta, "skills": b.Skills, "role_profiles": b.RoleProfiles}),
		"events.json":          marshal(t, map[string]any{"meta": b.Meta, "events": b.Events}),
		"activity_history.csv": []byte("record_id,employee_id,event_id,date,due_date,status,completion_pct,score,feedback_rating,assigned_by\nr1,person-01,architecture-lab,2026-01-12,,completed,100,85,4,self\n"),
	}
}
func TestFourFileImportAndConservativeHistory(t *testing.T) {
	b, issues := Parse(fourFiles(t), nil)
	if len(issues) > 0 {
		t.Fatal(issues)
	}
	a, _ := domain.Analyze(b, "person-01")
	if a.Levels["architecture"] != 2 || a.UncertainCompletions != 1 {
		t.Fatal("undated course incorrectly raised skill")
	}
}
func TestCSVDoesNotCoerceMalformedNumbers(t *testing.T) {
	files := fourFiles(t)
	files["activity_history.csv"] = []byte("record_id,employee_id,event_id,date,due_date,status,completion_pct,score,feedback_rating,assigned_by\nr1,person-01,architecture-lab,2026-01-12,,completed,not-a-number,85,4,self\n")
	_, issues := Parse(files, nil)
	if len(issues) != 1 || issues[0].Code != "parse" {
		t.Fatal("invalid integer accepted")
	}
}
func TestAppendPreservesExistingStateAndRejectsIDConflicts(t *testing.T) {
	base := demo.Dataset()
	person := base.Employees[0]
	person.ID = "new-person"
	files := map[string][]byte{"employees.json": marshal(t, map[string]any{"meta": base.Meta, "employees": []domain.Employee{person}}), "activity_history.csv": []byte("record_id,employee_id,event_id,date,due_date,status,completion_pct,score,feedback_rating,assigned_by\n")}
	merged, issues := Parse(files, &base)
	if len(issues) > 0 || len(merged.Employees) != len(base.Employees)+1 || len(merged.History) != len(base.History) {
		t.Fatalf("append failed: %v", issues)
	}
	files["employees.json"] = marshal(t, map[string]any{"meta": base.Meta, "employees": []domain.Employee{base.Employees[0]}})
	_, issues = Parse(files, &base)
	if len(issues) == 0 {
		t.Fatal("duplicate employee accepted")
	}
}
func TestBundleRoundTrip(t *testing.T) {
	original := demo.Dataset()
	restored, issues := Parse(map[string][]byte{"backup.json": marshal(t, original)}, nil)
	if len(issues) > 0 {
		t.Fatal(issues)
	}
	if string(marshal(t, restored)) != string(marshal(t, original)) {
		t.Fatal("export contract not preserved")
	}
}
