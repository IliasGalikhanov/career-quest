package httpapi

import (
	"careerquest/internal/domain"
	"net/http"
	"sort"
)

type skillGap struct {
	SkillID  string `json:"skill_id"`
	People   int    `json:"people"`
	Eligible int    `json:"eligible"`
	TotalGap int    `json:"total_gap"`
}
type noStep struct {
	EmployeeID string   `json:"employee_id"`
	Name       string   `json:"name"`
	Reasons    []string `json:"reasons"`
}
type participation struct {
	EventID  string         `json:"event_id"`
	Statuses map[string]int `json:"statuses"`
}

func (s *Server) hr(w http.ResponseWriter, r *http.Request) {
	d, err := s.store.Get(r.Context(), r.PathValue("id"))
	if handleError(w, err) {
		return
	}
	gaps := map[string]*skillGap{}
	people := map[string]bool{}
	missing := []noStep{}
	count := 0
	for _, e := range d.Bundle.Employees {
		if department := r.URL.Query().Get("department"); department != "" && department != e.Department {
			continue
		}
		if role := r.URL.Query().Get("role"); role != "" && role != e.Role {
			continue
		}
		people[e.ID] = true
		count++
		a, _ := domain.Analyze(d.Bundle, e.ID)
		for _, gap := range a.Gaps {
			if gap.Required == 0 {
				continue
			}
			g := gaps[gap.SkillID]
			if g == nil {
				g = &skillGap{SkillID: gap.SkillID}
				gaps[gap.SkillID] = g
			}
			g.Eligible++
			if gap.Current < gap.Required {
				g.People++
				g.TotalGap += gap.Required - gap.Current
			}
		}
		if len(a.Recommendations) == 0 {
			reasons := map[string]bool{}
			for _, c := range a.Catalog {
				for _, reason := range c.Blocked {
					reasons[reason] = true
				}
			}
			list := []string{}
			for reason := range reasons {
				list = append(list, reason)
			}
			sort.Strings(list)
			missing = append(missing, noStep{e.ID, e.Name, list})
		}
	}
	gapList := []skillGap{}
	for _, g := range gaps {
		gapList = append(gapList, *g)
	}
	sort.Slice(gapList, func(i, j int) bool {
		if gapList[i].People != gapList[j].People {
			return gapList[i].People > gapList[j].People
		}
		return gapList[i].SkillID < gapList[j].SkillID
	})
	events := map[string]*participation{}
	statuses := map[string]int{}
	for _, h := range d.Bundle.History {
		if !people[h.EmployeeID] || h.Date > d.Bundle.Settings.Date {
			continue
		}
		p := events[h.EventID]
		if p == nil {
			p = &participation{EventID: h.EventID, Statuses: map[string]int{}}
			events[h.EventID] = p
		}
		p.Statuses[h.Status]++
		statuses[h.Status]++
	}
	participationList := []participation{}
	for _, p := range events {
		participationList = append(participationList, *p)
	}
	sort.Slice(participationList, func(i, j int) bool { return participationList[i].EventID < participationList[j].EventID })
	respond(w, 200, map[string]any{"people": count, "gaps": gapList, "without_step": missing, "participation": participationList, "statuses": statuses, "date": d.Bundle.Settings.Date})
}
