export type Language = "ru" | "kk" | "en";
export type Localized = Partial<Record<Language, string>>;
export interface Goal {
  target_role: string;
  target_grade: string;
}
export interface Skill {
  skill_id: string;
  name: string;
  names?: Localized;
  type: string;
}
export interface Employee {
  employee_id: string;
  full_name: string;
  department: string;
  role: string;
  grade: string;
  skills: Record<string, number>;
  career_goal: Goal | null;
  last_review_date: string;
}
export interface RoleProfile {
  role: string;
  grade: string;
  required_skills: Record<string, number>;
  critical_skills: string[];
}
export interface Event {
  event_id: string;
  title: string;
  titles?: Localized;
  description: string;
  descriptions?: Localized;
  type: string;
  format: string;
  duration_hours: number;
  mandatory: boolean;
  prerequisites: Record<string, number>;
  upcoming_sessions: string[];
}
export interface Activity {
  record_id: string;
  employee_id: string;
  event_id: string;
  date: string;
  status: string;
  completion_pct: number;
  completed_at?: string;
  session_date?: string;
}
export interface Settings {
  date: string;
  max_hours: number;
  format: string;
  horizon_days: number;
  assume_completion_date: boolean;
}
export interface Bundle {
  meta: { dataset: string; as_of_date: string; synthetic: boolean };
  employees: Employee[];
  skills: Skill[];
  role_profiles: RoleProfile[];
  events: Event[];
  activity_history: Activity[];
  settings: Settings;
}
export interface Summary {
  id: string;
  name: string;
  revision: number;
}
export interface Dataset extends Summary {
  bundle: Bundle;
}
export interface Gap {
  skill_id: string;
  assessed: number;
  current: number;
  required: number;
  critical: boolean;
  missing_assessment: boolean;
}
export interface Candidate {
  event: Event;
  score: number;
  benefit: number;
  history_completed: number;
  history_unfinished: number;
  history_factor: number;
  improvements: {
    skill_id: string;
    from: number;
    to: number;
    required: number;
    critical: boolean;
  }[];
  blocked: string[];
  session_date: string;
}
export interface Analysis {
  employee: Employee;
  goal: Goal | null;
  suggested_goal: boolean;
  levels: Record<string, number>;
  gaps: Gap[];
  coverage: number | null;
  assessed_coverage: number | null;
  recommendations: Candidate[];
  catalog: Candidate[];
  history: Activity[];
  uncertain_completions: number;
  algorithm: string;
}
export interface HR {
  people: number;
  gaps: {
    skill_id: string;
    people: number;
    eligible: number;
    total_gap: number;
  }[];
  without_step: { employee_id: string; name: string; reasons: string[] }[];
  participation: { event_id: string; statuses: Record<string, number> }[];
  statuses: Record<string, number>;
  date: string;
}
export interface Preview {
  name: string;
  employees: number;
  events: number;
  skills: number;
  history: number;
  as_of_date: string;
  uncertain_dates: number;
}
export interface Problem {
  code: string;
  issues?: { path: string; code: string; detail: string }[];
}
