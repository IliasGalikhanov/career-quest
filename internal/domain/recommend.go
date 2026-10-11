package domain

import (
	"math"
	"slices"
	"sort"
	"time"
)

const AlgorithmVersion = "weighted-gap-v1"

type Gap struct {
	SkillID           string `json:"skill_id"`
	Assessed          int    `json:"assessed"`
	Current           int    `json:"current"`
	Required          int    `json:"required"`
	Critical          bool   `json:"critical"`
	MissingAssessment bool   `json:"missing_assessment"`
}
type Improvement struct {
	SkillID  string `json:"skill_id"`
	From     int    `json:"from"`
	To       int    `json:"to"`
	Required int    `json:"required"`
	Critical bool   `json:"critical"`
}
type Candidate struct {
	Event             Event         `json:"event"`
	Score             float64       `json:"score"`
	Benefit           float64       `json:"benefit"`
	CriticalGain      int           `json:"critical_gain"`
	HistoryCompleted  int           `json:"history_completed"`
	HistoryUnfinished int           `json:"history_unfinished"`
	HistoryFactor     float64       `json:"history_factor"`
	Improvements      []Improvement `json:"improvements"`
	Blocked           []string      `json:"blocked"`
	SessionDate       string        `json:"session_date"`
}
type Analysis struct {
	Employee             Employee       `json:"employee"`
	Goal                 *Goal          `json:"goal"`
	SuggestedGoal        bool           `json:"suggested_goal"`
	Levels               map[string]int `json:"levels"`
	Gaps                 []Gap          `json:"gaps"`
	Coverage             *float64       `json:"coverage"`
	AssessedCoverage     *float64       `json:"assessed_coverage"`
	Recommendations      []Candidate    `json:"recommendations"`
	Catalog              []Candidate    `json:"catalog"`
	History              []Activity     `json:"history"`
	UncertainCompletions int            `json:"uncertain_completions"`
	Algorithm            string         `json:"algorithm"`
}

func ApplyGain(level int, g Gain) int { return max(level, min(5, g.MaxLevel, level+g.Gain)) }

func EffectiveLevels(b Bundle, e Employee) (map[string]int, int) {
	levels := map[string]int{}
	for k, v := range e.Skills {
		levels[k] = v
	}
	events := map[string]Event{}
	for _, event := range b.Events {
		events[event.ID] = event
	}
	type completion struct {
		date string
		row  Activity
	}
	rows := []completion{}
	uncertain := 0
	for _, h := range b.History {
		if h.EmployeeID != e.ID || h.Status != "completed" || h.Date > b.Settings.Date {
			continue
		}
		date := h.CompletedAt
		if date == "" {
			if events[h.EventID].Format != "self_paced" {
				date = h.Date
			} else {
				uncertain++
				if b.Settings.AssumeCompletionDate {
					date = h.Date
				}
			}
		}
		if date != "" && date > e.ReviewDate && date <= b.Settings.Date {
			rows = append(rows, completion{date, h})
		}
	}
	// Date-only imports have no within-day timestamp. Preserve recorded order for
	// ties; sorting random IDs could change the result of differently capped gains.
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].date < rows[j].date })
	for _, r := range rows {
		gains := r.row.Gains
		if gains == nil {
			gains = events[r.row.EventID].Gains
		}
		for _, g := range gains {
			levels[g.SkillID] = ApplyGain(levels[g.SkillID], g)
		}
	}
	return levels, uncertain
}

func Analyze(b Bundle, employeeID string) (Analysis, bool) {
	a := Analysis{Gaps: []Gap{}, Recommendations: []Candidate{}, Catalog: []Candidate{}, History: []Activity{}, Algorithm: AlgorithmVersion}
	found := false
	for _, e := range b.Employees {
		if e.ID == employeeID {
			a.Employee = e
			found = true
			break
		}
	}
	if !found {
		return a, false
	}
	a.Goal = a.Employee.Goal
	if a.Goal == nil {
		index := slices.Index(Grades, a.Employee.Grade)
		if index >= 0 && index < len(Grades)-1 {
			goal := Goal{a.Employee.Role, Grades[index+1]}
			for _, p := range b.RoleProfiles {
				if p.Role == goal.Role && p.Grade == goal.Grade {
					a.Goal = &goal
					a.SuggestedGoal = true
					break
				}
			}
		}
	}
	a.Levels, a.UncertainCompletions = EffectiveLevels(b, a.Employee)
	profile := RoleProfile{Required: map[string]int{}}
	if a.Goal != nil {
		for _, p := range b.RoleProfiles {
			if p.Role == a.Goal.Role && p.Grade == a.Goal.Grade {
				profile = p
				break
			}
		}
	}
	current, assessed, total := 0, 0, 0
	for id, required := range profile.Required {
		_, known := a.Employee.Skills[id]
		a.Gaps = append(a.Gaps, Gap{id, a.Employee.Skills[id], a.Levels[id], required, slices.Contains(profile.Critical, id), !known})
		current += min(a.Levels[id], required)
		assessed += min(a.Employee.Skills[id], required)
		total += required
	}
	sort.Slice(a.Gaps, func(i, j int) bool {
		if a.Gaps[i].Critical != a.Gaps[j].Critical {
			return a.Gaps[i].Critical
		}
		return a.Gaps[i].SkillID < a.Gaps[j].SkillID
	})
	if total > 0 {
		coverage := 100 * float64(current) / float64(total)
		baseline := 100 * float64(assessed) / float64(total)
		a.Coverage = &coverage
		a.AssessedCoverage = &baseline
	}
	for _, h := range b.History {
		if h.EmployeeID == employeeID && h.Date <= b.Settings.Date {
			a.History = append(a.History, h)
		}
	}
	sort.SliceStable(a.History, func(i, j int) bool { return a.History[i].Date > a.History[j].Date })
	events := map[string]Event{}
	for _, e := range b.Events {
		events[e.ID] = e
	}
	today, _ := time.Parse(time.DateOnly, b.Settings.Date)
	until := today.AddDate(0, 0, b.Settings.HorizonDays).Format(time.DateOnly)
	since := today.AddDate(0, 0, -180).Format(time.DateOnly)
	for _, event := range b.Events {
		c := Candidate{Event: event, Blocked: []string{}, Improvements: []Improvement{}}
		if event.Mandatory {
			c.Blocked = append(c.Blocked, "mandatory")
		}
		if !slices.Contains(event.Roles, a.Employee.Role) || !slices.Contains(event.Grades, a.Employee.Grade) {
			c.Blocked = append(c.Blocked, "audience")
		}
		for id, level := range event.Prerequisites {
			if a.Levels[id] < level {
				c.Blocked = append(c.Blocked, "prerequisites")
				break
			}
		}
		if event.Format != "self_paced" {
			sessions := slices.Clone(event.Sessions)
			sort.Strings(sessions)
			for _, date := range sessions {
				if date >= b.Settings.Date && date <= until {
					c.SessionDate = date
					break
				}
			}
			if c.SessionDate == "" {
				c.Blocked = append(c.Blocked, "no_session")
			}
		}
		if event.Hours > b.Settings.MaxHours {
			c.Blocked = append(c.Blocked, "time_budget")
		}
		if b.Settings.Format != "any" && b.Settings.Format != event.Format {
			c.Blocked = append(c.Blocked, "format")
		}
		for _, h := range a.History {
			if h.EventID == event.ID {
				if h.Status == "completed" && (!event.Repeatable || (event.Format != "self_paced" && h.SessionDate == c.SessionDate) || (event.Format == "self_paced" && h.CompletedAt == b.Settings.Date)) {
					c.Blocked = append(c.Blocked, "completed")
				}
				if h.Status == "in_progress" {
					c.Blocked = append(c.Blocked, "in_progress")
				}
			}
			past := events[h.EventID]
			date := h.CompletedAt
			if date == "" {
				date = h.Date
			}
			if date < since || date > b.Settings.Date || past.Mandatory || past.Type != event.Type || past.Format != event.Format {
				continue
			}
			switch h.Status {
			case "completed":
				c.HistoryCompleted++
			case "dropped", "declined", "no_show":
				c.HistoryUnfinished++
			}
		}
		for _, g := range event.Gains {
			required := profile.Required[g.SkillID]
			from := a.Levels[g.SkillID]
			to := ApplyGain(from, g)
			delta := min(max(0, required-from), to-from)
			if delta > 0 {
				critical := slices.Contains(profile.Critical, g.SkillID)
				weight := 1
				if critical {
					weight = 2
					c.CriticalGain += delta
				}
				c.Benefit += float64(delta * weight)
				c.Improvements = append(c.Improvements, Improvement{g.SkillID, from, to, required, critical})
			}
		}
		if a.Goal == nil {
			c.Blocked = append(c.Blocked, "no_goal")
		} else if c.Benefit == 0 {
			c.Blocked = append(c.Blocked, "no_gain")
		}
		c.HistoryFactor = 0.75 + 0.5*float64(c.HistoryCompleted+2)/float64(c.HistoryCompleted+c.HistoryUnfinished+4)
		c.Score = c.Benefit * c.HistoryFactor / math.Sqrt(max(1, event.Hours))
		sort.Strings(c.Blocked)
		c.Blocked = slices.Compact(c.Blocked)
		a.Catalog = append(a.Catalog, c)
		if len(c.Blocked) == 0 {
			a.Recommendations = append(a.Recommendations, c)
		}
	}
	sort.Slice(a.Recommendations, func(i, j int) bool {
		x, y := a.Recommendations[i], a.Recommendations[j]
		if x.Score != y.Score {
			return x.Score > y.Score
		}
		if x.CriticalGain != y.CriticalGain {
			return x.CriticalGain > y.CriticalGain
		}
		if x.Event.Hours != y.Event.Hours {
			return x.Event.Hours < y.Event.Hours
		}
		return x.Event.ID < y.Event.ID
	})
	if len(a.Recommendations) > 3 {
		a.Recommendations = a.Recommendations[:3]
	}
	return a, true
}
