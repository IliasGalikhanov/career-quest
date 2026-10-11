// Package demo builds independent, explicitly fictional development scenarios.
package demo

import (
	"careerquest/internal/domain"
	"fmt"
	"math/rand"
)

func text(en, ru, kk string) domain.Localized { return domain.Localized{"en": en, "ru": ru, "kk": kk} }

// Dataset is deterministic. It does not read or transform any external dataset.
func Dataset() domain.Bundle {
	b := domain.Bundle{Meta: domain.Meta{Dataset: "Northstar Studio · synthetic demo", Version: "1.0", AsOfDate: "2026-02-02", Synthetic: true}, Settings: domain.DefaultSettings("2026-02-02"), History: []domain.Activity{}}
	definitions := []struct{ id, en, ru, kk, kind string }{
		{"architecture", "Architecture", "Архитектура", "Архитектура", "hard"},
		{"api", "API design", "Проектирование API", "API жобалау", "hard"},
		{"testing", "Testing", "Тестирование", "Тестілеу", "hard"},
		{"communication", "Communication", "Коммуникация", "Қарым-қатынас", "soft"},
		{"research", "User research", "Исследования пользователей", "Пайдаланушыларды зерттеу", "hard"},
		{"visual", "Visual design", "Визуальный дизайн", "Визуалды дизайн", "hard"},
		{"analysis", "Data analysis", "Анализ данных", "Деректерді талдау", "hard"},
		{"querying", "Data querying", "Запросы к данным", "Деректерге сұрау жасау", "hard"},
		{"mentoring", "Mentoring", "Наставничество", "Тәлімгерлік", "soft"},
		{"planning", "Planning", "Планирование", "Жоспарлау", "soft"},
	}
	for _, s := range definitions {
		b.Skills = append(b.Skills, domain.Skill{ID: s.id, Name: s.en, Type: s.kind, Category: s.kind, Description: s.en, Names: text(s.en, s.ru, s.kk)})
	}
	roles := []string{"Platform Engineer", "Product Designer", "Analytics Specialist"}
	tracks := [][]string{{"architecture", "api", "testing"}, {"research", "visual", "communication"}, {"analysis", "querying", "planning"}}
	for i, role := range roles {
		for j, grade := range domain.Grades {
			required := map[string]int{"communication": min(4, j+1), "planning": min(4, j+1)}
			for _, id := range tracks[i] {
				required[id] = min(5, j+2)
			}
			if j == 3 {
				required["mentoring"] = 3
			}
			b.RoleProfiles = append(b.RoleProfiles, domain.RoleProfile{Role: role, Grade: grade, Required: required, Critical: []string{tracks[i][0]}})
		}
	}
	rng := rand.New(rand.NewSource(73))
	names := []string{"Alex Demo", "Dana Demo", "Sam Demo", "Mira Demo", "Noah Demo", "Ari Demo", "Kai Demo", "Rin Demo", "Robin Demo", "Lee Demo", "Ash Demo", "Kim Demo"}
	for i, name := range names {
		r := i % 3
		grade := domain.Grades[(i/3)%4]
		levels := map[string]int{}
		for _, s := range b.Skills {
			levels[s.ID] = rng.Intn(3) + 1
		}
		var goal *domain.Goal
		if grade != "Lead" {
			goal = &domain.Goal{Role: roles[r], Grade: domain.Grades[(i/3)+1]}
		}
		b.Employees = append(b.Employees, domain.Employee{ID: fmt.Sprintf("person-%02d", i+1), Name: name, Department: []string{"Platform", "Experience", "Insights"}[r], Role: roles[r], Grade: grade, HireDate: "2024-06-01", TenureMonths: 20, WorkFormat: "hybrid", Language: "en", Goal: goal, Skills: levels, ReviewDate: "2026-01-10"})
	}
	b.Employees[0].Grade = "Middle"
	b.Employees[0].Goal = &domain.Goal{Role: roles[0], Grade: "Senior"}
	b.Employees[0].Skills["architecture"] = 2
	b.Employees[0].Skills["api"] = 3
	b.Employees[0].Skills["testing"] = 3
	b.Employees[0].Skills["communication"] = 1
	b.Employees[2].Goal = &domain.Goal{Role: roles[0], Grade: "Middle"}
	b.Employees[4].Goal = nil
	add := func(id, en, ru, kk string, role []string, skill string, gain, cap int, hours float64, format string) {
		b.Events = append(b.Events, domain.Event{ID: id, Title: en, Titles: text(en, ru, kk), Description: "A focused practice session with a concrete deliverable and a short review.", Descriptions: text("Practice a focused skill, produce a small deliverable, and reflect on the result.", "Отработайте навык, подготовьте небольшой результат и проведите самооценку.", "Дағдыны жаттықтырып, шағын жұмыс дайындаңыз және нәтижені бағалаңыз."), Type: "workshop", Format: format, Hours: hours, Roles: role, Grades: domain.Grades, Gains: []domain.Gain{{SkillID: skill, Gain: gain, MaxLevel: cap}}, Prerequisites: map[string]int{}, Sessions: []string{}})
	}
	add("architecture-lab", "Design a resilient service", "Спроектировать устойчивый сервис", "Тұрақты сервисті жобалау", []string{roles[0]}, "architecture", 1, 4, 4, "self_paced")
	add("api-clinic", "Review an API contract", "Ревью контракта API", "API келісімшартын тексеру", []string{roles[0]}, "api", 1, 5, 3, "self_paced")
	add("test-workbench", "Build a testing strategy", "Разработать стратегию тестирования", "Тестілеу стратегиясын құру", []string{roles[0]}, "testing", 1, 4, 5, "self_paced")
	add("research-sprint", "Plan a discovery interview", "Подготовить исследовательское интервью", "Зерттеу сұхбатын жоспарлау", []string{roles[1]}, "research", 1, 5, 4, "self_paced")
	add("design-critique", "Run a design critique", "Провести дизайн-ревью", "Дизайнды талқылау", []string{roles[1]}, "visual", 1, 4, 2, "online")
	b.Events[len(b.Events)-1].Sessions = []string{"2026-02-10", "2026-03-10"}
	add("analysis-lab", "Explain an experiment", "Объяснить результат эксперимента", "Эксперимент нәтижесін түсіндіру", []string{roles[2]}, "analysis", 1, 4, 4, "self_paced")
	add("query-lab", "Investigate a slow query", "Исследовать медленный запрос", "Баяу сұрауды зерттеу", []string{roles[2], roles[0]}, "querying", 1, 4, 3, "self_paced")
	add("story-circle", "Share a technical story", "Рассказать о техническом решении", "Техникалық шешімді таныстыру", roles, "communication", 1, 4, 2, "offline")
	b.Events[len(b.Events)-1].Sessions = []string{"2026-02-02", "2026-02-16", "2026-03-02"}
	b.Events[len(b.Events)-1].Repeatable = true
	add("planning-room", "Plan a small delivery", "Спланировать небольшой проект", "Шағын жобаны жоспарлау", roles, "planning", 1, 4, 3, "self_paced")
	add("mentor-studio", "Practice a mentoring session", "Потренировать наставническую встречу", "Тәлімгерлік кездесуді жаттықтыру", roles, "mentoring", 1, 4, 6, "online")
	b.Events[len(b.Events)-1].Sessions = []string{"2026-02-20"}
	b.Events[len(b.Events)-1].Prerequisites = map[string]int{"communication": 3}
	add("architecture-deep-dive", "Review distributed tradeoffs", "Разобрать распределённую архитектуру", "Таратылған архитектураны талдау", []string{roles[0]}, "architecture", 1, 5, 8, "self_paced")
	b.Events[len(b.Events)-1].Prerequisites = map[string]int{"architecture": 4}
	add("orientation", "Workspace orientation", "Знакомство с рабочей средой", "Жұмыс ортасымен танысу", roles, "planning", 1, 2, 1, "self_paced")
	b.Events[len(b.Events)-1].Mandatory = true
	b.Events[len(b.Events)-1].Gains = []domain.Gain{}
	for i, date := range []string{"2025-11-03", "2025-12-01", "2026-01-05"} {
		b.History = append(b.History, domain.Activity{ID: fmt.Sprintf("past-%d", i), EmployeeID: b.Employees[0].ID, EventID: "story-circle", Date: date, Status: "no_show", AssignedBy: "self"})
	}
	b.History = append(b.History, domain.Activity{ID: "past-4", EmployeeID: b.Employees[1].ID, EventID: "research-sprint", Date: "2026-01-12", CompletedAt: "2026-01-14", Status: "completed", Completion: 100, AssignedBy: "self"})
	b.History = append(b.History, domain.Activity{ID: "past-5", EmployeeID: b.Employees[2].ID, EventID: "analysis-lab", Date: "2026-01-13", Status: "completed", Completion: 100, AssignedBy: "self"})
	return b
}
