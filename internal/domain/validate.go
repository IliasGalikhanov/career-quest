package domain

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

func ValidDate(value string) bool {
	t, err := time.Parse(time.DateOnly, value)
	return err == nil && t.Format(time.DateOnly) == value
}

func ValidateSettings(s Settings) []Issue {
	issues := []Issue{}
	if !ValidDate(s.Date) {
		issues = append(issues, Issue{"settings.date", "date", "Use YYYY-MM-DD"})
	}
	if s.MaxHours <= 0 || s.MaxHours > 1000 || math.IsNaN(s.MaxHours) || math.IsInf(s.MaxHours, 0) {
		issues = append(issues, Issue{"settings.max_hours", "range", "Expected a number in (0, 1000]"})
	}
	if s.HorizonDays < 1 || s.HorizonDays > 730 {
		issues = append(issues, Issue{"settings.horizon_days", "range", "Expected 1–730 days"})
	}
	if !slices.Contains([]string{"any", "online", "offline", "self_paced"}, s.Format) {
		issues = append(issues, Issue{"settings.format", "enum", "Unknown format"})
	}
	return issues
}

// Validate reports paths instead of silently repairing an inconsistent upload.
func Validate(b Bundle) []Issue {
	issues := []Issue{}
	add := func(path, code, detail string) {
		if len(issues) < 100 {
			issues = append(issues, Issue{path, code, detail})
		}
	}
	if !b.Meta.Synthetic {
		add("meta.synthetic", "synthetic_required", "Only synthetic datasets are supported")
	}
	if strings.TrimSpace(b.Meta.Dataset) == "" || utf8.RuneCountInString(b.Meta.Dataset) > 120 {
		add("meta.dataset", "required", "Use a dataset name of 1–120 characters")
	}
	if !ValidDate(b.Meta.AsOfDate) {
		add("meta.as_of_date", "date", "Use YYYY-MM-DD")
	}
	if b.Settings.Date < b.Meta.AsOfDate {
		add("settings.date", "date", "Calculation cannot precede the assessment snapshot")
	}
	for _, issue := range ValidateSettings(b.Settings) {
		add(issue.Path, issue.Code, issue.Detail)
	}
	if len(b.Employees) == 0 || len(b.Employees) > 5000 {
		add("employees", "size", "Expected 1–5000 employees")
	}
	if len(b.Skills) == 0 || len(b.Skills) > 500 || len(b.Events) > 2000 || len(b.History) > 100000 {
		add("bundle", "size", "Maximum: 500 skills, 2000 events, 100000 history rows")
	}
	ids := func(id, path string, seen map[string]bool) {
		if strings.TrimSpace(id) == "" || len(id) > 100 || strings.ContainsAny(id, "/\\\x00") {
			add(path, "id", "Use a nonempty ID without slashes, up to 100 characters")
		}
		if seen[id] {
			add(path, "duplicate", id)
		}
		seen[id] = true
	}
	skills := map[string]bool{}
	for i, s := range b.Skills {
		p := fmt.Sprintf("skills[%d]", i)
		ids(s.ID, p+".skill_id", skills)
		if s.Name == "" {
			add(p+".name", "required", "Name is required")
		}
		if s.Type != "hard" && s.Type != "soft" {
			add(p+".type", "enum", "Expected hard or soft")
		}
	}
	levels := func(values map[string]int, path string) {
		for id, level := range values {
			if !skills[id] {
				add(path+"."+id, "reference", "Unknown skill")
			}
			if level < 0 || level > 5 {
				add(path+"."+id, "range", "Expected level 0–5")
			}
		}
	}
	profiles := map[string]bool{}
	roles := map[string]bool{}
	for i, p := range b.RoleProfiles {
		path := fmt.Sprintf("role_profiles[%d]", i)
		key := p.Role + "/" + p.Grade
		if profiles[key] {
			add(path, "duplicate", key)
		}
		profiles[key] = true
		roles[p.Role] = true
		if p.Role == "" || !slices.Contains(Grades, p.Grade) {
			add(path, "enum", "Invalid role or grade")
		}
		levels(p.Required, path+".required_skills")
		for _, id := range p.Critical {
			if _, ok := p.Required[id]; !ok {
				add(path+".critical_skills", "reference", "Critical skill must have a requirement")
			}
		}
	}
	employees := map[string]bool{}
	for i, e := range b.Employees {
		p := fmt.Sprintf("employees[%d]", i)
		ids(e.ID, p+".employee_id", employees)
		if e.Name == "" {
			add(p+".full_name", "required", "Name is required")
		}
		if !profiles[e.Role+"/"+e.Grade] {
			add(p, "reference", "Unknown role and grade")
		}
		levels(e.Skills, p+".skills")
		if !ValidDate(e.HireDate) || !ValidDate(e.ReviewDate) || e.HireDate > e.ReviewDate || e.ReviewDate > b.Meta.AsOfDate {
			add(p, "date", "Expected hire_date <= last_review_date <= as_of_date")
		}
		if e.Goal != nil && !profiles[e.Goal.Role+"/"+e.Goal.Grade] {
			add(p+".career_goal", "reference", "Unknown target role and grade")
		}
	}
	for i, e := range b.Employees {
		if e.ManagerID != nil && (!employees[*e.ManagerID] || *e.ManagerID == e.ID) {
			add(fmt.Sprintf("employees[%d].manager_id", i), "reference", "Invalid manager")
		}
	}
	eventIDs := map[string]bool{}
	events := map[string]Event{}
	checkGains := func(gains []Gain, path string) {
		seen := map[string]bool{}
		for _, g := range gains {
			if !skills[g.SkillID] {
				add(path, "reference", "Unknown developed skill")
			}
			if seen[g.SkillID] {
				add(path, "duplicate", "Repeated skill gain")
			}
			seen[g.SkillID] = true
			if g.Gain < 1 || g.Gain > 5 || g.MaxLevel < 1 || g.MaxLevel > 5 {
				add(path, "range", "Gain and max_level must be 1–5")
			}
		}
	}
	for i, e := range b.Events {
		p := fmt.Sprintf("events[%d]", i)
		ids(e.ID, p+".event_id", eventIDs)
		events[e.ID] = e
		if e.Title == "" {
			add(p+".title", "required", "Title is required")
		}
		if e.Hours <= 0 || e.Hours > 1000 {
			add(p+".duration_hours", "range", "Expected (0, 1000]")
		}
		if !slices.Contains([]string{"online", "offline", "self_paced"}, e.Format) {
			add(p+".format", "enum", "Unknown format")
		}
		if len(e.Roles) == 0 || len(e.Grades) == 0 {
			add(p, "required", "Event audience is required")
		}
		for _, r := range e.Roles {
			if !roles[r] {
				add(p+".target_roles", "reference", "Unknown role")
			}
		}
		for _, g := range e.Grades {
			if !slices.Contains(Grades, g) {
				add(p+".target_grades", "enum", "Unknown grade")
			}
		}
		levels(e.Prerequisites, p+".prerequisites")
		checkGains(e.Gains, p+".develops_skills")
		for _, date := range e.Sessions {
			if !ValidDate(date) {
				add(p+".upcoming_sessions", "date", "Use YYYY-MM-DD")
			}
		}
	}
	historyIDs := map[string]bool{}
	completed := map[string]bool{}
	open := map[string]bool{}
	sessions := map[string]bool{}
	for i, h := range b.History {
		if utf8.RuneCountInString(h.Reflection) > 2000 {
			add(fmt.Sprintf("activity_history[%d].reflection", i), "size", "Maximum 2000 characters")
		}
		p := fmt.Sprintf("activity_history[%d]", i)
		ids(h.ID, p+".record_id", historyIDs)
		if !employees[h.EmployeeID] || !eventIDs[h.EventID] {
			add(p, "reference", "Unknown employee or event")
		}
		if !ValidDate(h.Date) || h.Date > b.Settings.Date {
			add(p+".date", "date", "Expected YYYY-MM-DD no later than the calculation date")
		}
		if h.CompletedAt != "" && (!ValidDate(h.CompletedAt) || h.CompletedAt < h.Date || h.CompletedAt > b.Settings.Date) {
			add(p+".completed_at", "date", "Expected enrollment <= completion <= calculation date")
		}
		if h.SessionDate != "" && (!ValidDate(h.SessionDate) || (h.Status == "completed" && h.SessionDate > h.CompletedAt)) {
			add(p+".session_date", "date", "Completed sessions require a valid session date no later than completed_at")
		}
		if h.DueDate != "" && !ValidDate(h.DueDate) {
			add(p+".due_date", "date", "Use YYYY-MM-DD")
		}
		if !slices.Contains([]string{"completed", "in_progress", "dropped", "declined", "no_show", "overdue"}, h.Status) {
			add(p+".status", "enum", "Unknown status")
		}
		if h.Completion < 0 || h.Completion > 100 || (h.Status == "completed" && h.Completion != 100) || (h.Status != "completed" && h.Completion == 100) || ((h.Status == "declined" || h.Status == "no_show") && h.Completion != 0) {
			add(p+".completion_pct", "range", "Percentage does not match status")
		}
		if (h.Status == "dropped" && h.Completion < 5) || (h.Status != "completed" && h.Completion > 95) {
			add(p+".completion_pct", "range", "Incomplete participation is at most 95%; dropped work is at least 5%")
		}
		if h.Status == "no_show" && events[h.EventID].Format == "self_paced" {
			add(p+".status", "enum", "Self-paced activities do not have no-show outcomes")
		}
		if h.Status == "overdue" && (!events[h.EventID].Mandatory || h.DueDate == "" || h.DueDate >= b.Settings.Date) {
			add(p+".due_date", "date", "Overdue requires a mandatory activity and a past due date")
		}
		if h.Score != nil && (*h.Score < 0 || *h.Score > 100) {
			add(p+".score", "range", "Expected 0–100")
		}
		if h.Rating != nil && (*h.Rating < 1 || *h.Rating > 5) {
			add(p+".feedback_rating", "range", "Expected 1–5")
		}
		if !slices.Contains([]string{"self", "manager", "hr"}, h.AssignedBy) {
			add(p+".assigned_by", "enum", "Unknown initiator")
		}
		if h.Status != "completed" && (h.CompletedAt != "" || len(h.Gains) > 0) {
			add(p, "status", "Only completed records can carry completion effects")
		}
		checkGains(h.Gains, p+".applied_gains")
		key := h.EmployeeID + "/" + h.EventID
		if h.Status == "completed" {
			if completed[key] && !events[h.EventID].Repeatable {
				add(p, "duplicate_completion", "Nonrepeatable event completed more than once")
			}
			completed[key] = true
			date := h.SessionDate
			if date == "" {
				date = h.CompletedAt
			}
			if date == "" {
				date = h.Date
			}
			sessionKey := key + "/" + date
			if sessions[sessionKey] {
				add(p, "duplicate_completion", "A session can be completed only once")
			}
			sessions[sessionKey] = true
		}
		if h.Status == "in_progress" {
			if open[key] {
				add(p, "duplicate", "Multiple open attempts")
			}
			open[key] = true
		}
	}
	for _, h := range b.History {
		if h.Status == "in_progress" && completed[h.EmployeeID+"/"+h.EventID] && !events[h.EventID].Repeatable {
			add("activity_history", "duplicate_completion", "A completed nonrepeatable event cannot also be in progress")
		}
	}
	return issues
}
