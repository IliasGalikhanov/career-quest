package demo

import (
	"careerquest/internal/domain"
	"fmt"
)

// Scaled repeats independent fictional profiles for technical load checks only.
// It does not create additional evidence of recommendation quality.
func Scaled(people int) domain.Bundle {
	b := Dataset()
	source := b.Employees
	b.Employees = make([]domain.Employee, 0, people)
	b.History = make([]domain.Activity, 0, people*3)
	for i := 0; i < people; i++ {
		p := source[i%len(source)]
		p.ID = fmt.Sprintf("load-%05d", i)
		p.Name = fmt.Sprintf("Synthetic profile %d", i)
		p.ManagerID = nil
		b.Employees = append(b.Employees, p)
		for j := 0; j < 3; j++ {
			b.History = append(b.History, domain.Activity{ID: fmt.Sprintf("load-history-%d-%d", i, j), EmployeeID: p.ID, EventID: "story-circle", Date: fmt.Sprintf("2026-01-%02d", j+1), Status: "no_show", AssignedBy: "self"})
		}
	}
	b.Meta.Dataset = fmt.Sprintf("Synthetic load fixture: %d profiles", people)
	return b
}
