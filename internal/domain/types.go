// Package domain defines the portable dataset contract and development rules.
package domain

type Localized map[string]string

type Meta struct {
	Dataset   string `json:"dataset"`
	Version   string `json:"version"`
	AsOfDate  string `json:"as_of_date"`
	Synthetic bool   `json:"synthetic"`
}

type Skill struct {
	ID          string    `json:"skill_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Names       Localized `json:"names,omitempty"`
}

type RoleProfile struct {
	Role     string         `json:"role"`
	Grade    string         `json:"grade"`
	Required map[string]int `json:"required_skills"`
	Critical []string       `json:"critical_skills"`
}

type Goal struct {
	Role  string `json:"target_role"`
	Grade string `json:"target_grade"`
}

type Employee struct {
	ID           string         `json:"employee_id"`
	Name         string         `json:"full_name"`
	Department   string         `json:"department"`
	Role         string         `json:"role"`
	Grade        string         `json:"grade"`
	ManagerID    *string        `json:"manager_id"`
	HireDate     string         `json:"hire_date"`
	TenureMonths int            `json:"tenure_months"`
	WorkFormat   string         `json:"work_format"`
	Language     string         `json:"preferred_language"`
	Goal         *Goal          `json:"career_goal"`
	Skills       map[string]int `json:"skills"`
	ReviewDate   string         `json:"last_review_date"`
}

type Gain struct {
	SkillID  string `json:"skill_id"`
	Gain     int    `json:"gain"`
	MaxLevel int    `json:"max_level"`
}

type Event struct {
	Deliverable   Localized      `json:"deliverable,omitempty"`
	Criteria      []Localized    `json:"criteria,omitempty"`
	ID            string         `json:"event_id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Titles        Localized      `json:"titles,omitempty"`
	Descriptions  Localized      `json:"descriptions,omitempty"`
	Type          string         `json:"type"`
	Format        string         `json:"format"`
	Hours         float64        `json:"duration_hours"`
	Mandatory     bool           `json:"mandatory"`
	Roles         []string       `json:"target_roles"`
	Grades        []string       `json:"target_grades"`
	Gains         []Gain         `json:"develops_skills"`
	Prerequisites map[string]int `json:"prerequisites"`
	Sessions      []string       `json:"upcoming_sessions"`
	Repeatable    bool           `json:"repeatable"`
}

type Activity struct {
	Reflection  string `json:"reflection,omitempty"`
	ID          string `json:"record_id"`
	EmployeeID  string `json:"employee_id"`
	EventID     string `json:"event_id"`
	Date        string `json:"date"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status"`
	Completion  int    `json:"completion_pct"`
	Score       *int   `json:"score"`
	Rating      *int   `json:"feedback_rating"`
	AssignedBy  string `json:"assigned_by"`
	CompletedAt string `json:"completed_at,omitempty"`
	// Gains freeze the catalog rule at completion; later imports cannot rewrite it.
	Gains       []Gain `json:"applied_gains,omitempty"`
	SessionDate string `json:"session_date,omitempty"`
}

type Settings struct {
	Date                 string  `json:"date"`
	MaxHours             float64 `json:"max_hours"`
	Format               string  `json:"format"`
	HorizonDays          int     `json:"horizon_days"`
	AssumeCompletionDate bool    `json:"assume_completion_date"`
}

type Bundle struct {
	Meta         Meta          `json:"meta"`
	Skills       []Skill       `json:"skills"`
	RoleProfiles []RoleProfile `json:"role_profiles"`
	Employees    []Employee    `json:"employees"`
	Events       []Event       `json:"events"`
	History      []Activity    `json:"activity_history"`
	Settings     Settings      `json:"settings"`
}

var Grades = []string{"Junior", "Middle", "Senior", "Lead"}

func DefaultSettings(date string) Settings {
	return Settings{Date: date, MaxHours: 40, Format: "any", HorizonDays: 90}
}

type Issue struct {
	Path   string `json:"path"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}
