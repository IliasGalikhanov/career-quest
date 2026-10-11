// Package importer reads portable JSON bundles and the four-file JSON/CSV format.
package importer

import (
	"bytes"
	"careerquest/internal/domain"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

func decode(data []byte, out any) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("UTF-8 encoding required")
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	d := json.NewDecoder(bytes.NewReader(data))
	if err := d.Decode(out); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("expected one JSON document")
	}
	return nil
}

func Parse(files map[string][]byte, base *domain.Bundle) (domain.Bundle, []domain.Issue) {
	b := domain.Bundle{History: []domain.Activity{}}
	fail := func(path string, err error) (domain.Bundle, []domain.Issue) {
		return b, []domain.Issue{{Path: path, Code: "parse", Detail: err.Error()}}
	}
	if len(files) == 1 && base == nil {
		for name, data := range files {
			if !strings.HasSuffix(strings.ToLower(name), ".json") {
				return fail(name, fmt.Errorf("expected a JSON bundle"))
			}
			if err := decode(data, &b); err != nil {
				return fail(name, err)
			}
		}
	} else {
		if base != nil {
			b = *base
		}
		allowed := map[string]bool{"employees.json": true, "activity_history.csv": true}
		if base == nil {
			allowed["skills.json"] = true
			allowed["events.json"] = true
		}
		for name := range files {
			if !allowed[name] {
				return fail(name, fmt.Errorf("unexpected file for selected import mode"))
			}
		}
		for name := range allowed {
			if _, ok := files[name]; !ok {
				return fail(name, fmt.Errorf("required file is missing"))
			}
		}
		var employees struct {
			Meta      domain.Meta       `json:"meta"`
			Employees []domain.Employee `json:"employees"`
		}
		if err := decode(files["employees.json"], &employees); err != nil {
			return fail("employees.json", err)
		}
		if base == nil {
			b.Meta = employees.Meta
			b.Meta.Synthetic = true
			var skills struct {
				Meta     domain.Meta          `json:"meta"`
				Skills   []domain.Skill       `json:"skills"`
				Profiles []domain.RoleProfile `json:"role_profiles"`
			}
			if err := decode(files["skills.json"], &skills); err != nil {
				return fail("skills.json", err)
			}
			var events struct {
				Meta   domain.Meta    `json:"meta"`
				Events []domain.Event `json:"events"`
			}
			if err := decode(files["events.json"], &events); err != nil {
				return fail("events.json", err)
			}
			if skills.Meta.AsOfDate != b.Meta.AsOfDate || events.Meta.AsOfDate != b.Meta.AsOfDate {
				return fail("meta.as_of_date", fmt.Errorf("all files must describe the same snapshot"))
			}
			b.Skills = skills.Skills
			b.RoleProfiles = skills.Profiles
			b.Events = events.Events
			b.Settings = domain.DefaultSettings(b.Meta.AsOfDate)
		} else if employees.Meta.AsOfDate != b.Meta.AsOfDate {
			return fail("employees.json", fmt.Errorf("append requires the same snapshot date"))
		}
		b.Employees = append(b.Employees, employees.Employees...)
		history, err := parseCSV(files["activity_history.csv"])
		if err != nil {
			return fail("activity_history.csv", err)
		}
		b.History = append(b.History, history...)
	}
	if b.Settings.Date == "" {
		b.Settings = domain.DefaultSettings(b.Meta.AsOfDate)
	}
	return b, domain.Validate(b)
}

func parseCSV(data []byte) ([]domain.Activity, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("UTF-8 encoding required")
	}
	r := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})))
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	columns := map[string]int{}
	for i, name := range header {
		if _, exists := columns[name]; exists {
			return nil, fmt.Errorf("duplicate column %s", name)
		}
		columns[name] = i
	}
	for _, name := range []string{"record_id", "employee_id", "event_id", "date", "due_date", "status", "completion_pct", "score", "feedback_rating", "assigned_by"} {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("missing column %s", name)
		}
	}
	out := []domain.Activity{}
	for line := 2; ; line++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		get := func(name string) string {
			if i, ok := columns[name]; ok {
				return row[i]
			}
			return ""
		}
		number := func(name string, required bool) (*int, error) {
			value := get(name)
			if value == "" && !required {
				return nil, nil
			}
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("row %d, %s: expected an integer", line, name)
			}
			return &n, nil
		}
		pct, err := number("completion_pct", true)
		if err != nil {
			return nil, err
		}
		score, err := number("score", false)
		if err != nil {
			return nil, err
		}
		rating, err := number("feedback_rating", false)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.Activity{ID: get("record_id"), EmployeeID: get("employee_id"), EventID: get("event_id"), Date: get("date"), DueDate: get("due_date"), Status: get("status"), Completion: *pct, Score: score, Rating: rating, AssignedBy: get("assigned_by"), CompletedAt: get("completed_at"), SessionDate: get("session_date")})
	}
	return out, nil
}
