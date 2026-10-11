import {
  ArrowUpRight,
  Clock3,
  Compass,
  CalendarDays,
  CheckCircle2,
} from "lucide-react";
import { localize, type T } from "../i18n";
import type { Candidate, Language, Skill, Goal } from "../types";

interface Props {
  candidate: Candidate;
  language: Language;
  t: T;
  skills: Skill[];
  goal: Goal | null;
  busy: boolean;
  onStart: (id: string) => void;
  compact?: boolean;
}

export default function ActivityCard({
  candidate: c,
  language,
  t,
  skills,
  goal,
  busy,
  onStart,
  compact,
}: Props) {
  const name = (id: string) => {
    const skill = skills.find((s) => s.skill_id === id);
    return localize(skill?.names, skill?.name ?? id, language);
  };
  const canStart = c.blocked.every(
    (reason) => reason === "no_gain" || reason === "no_goal",
  );
  return (
    <article
      className={`activity-card ${compact ? "compact" : ""}`}
      data-testid={`activity-${c.event.event_id}`}
    >
      <div className="card-top">
        <span className="activity-icon">
          <Compass size={21} />
        </span>
        <span className="tag">{t(c.event.format)}</span>
      </div>
      <h3>{localize(c.event.titles, c.event.title, language)}</h3>
      <p className="muted">
        {localize(c.event.descriptions, c.event.description, language)}
      </p>
      <div className="activity-meta">
        <span>
          <Clock3 size={15} />
          {c.event.duration_hours} {t("hours")}
        </span>
        {c.session_date && (
          <span>
            <CalendarDays size={15} />
            {c.session_date}
          </span>
        )}
      </div>
      {!compact && (
        <div className="reason-box">
          <h4>{t("why")}</h4>
          <p>
            <CheckCircle2 size={14} />
            <span>
              {t("goalFactor")}: {goal?.target_role} · {goal?.target_grade}
            </span>
          </p>
          {c.improvements.map((g) => (
            <p key={g.skill_id}>
              <ArrowUpRight size={14} />
              <span>
                {name(g.skill_id)}:{" "}
                <strong>
                  {g.from} → {g.to}
                </strong>{" "}
                / {g.required}
                {g.critical ? ` · ${t("critical")}` : ""}
              </span>
            </p>
          ))}
          <p>
            <Clock3 size={14} />
            <span>
              {c.history_completed + c.history_unfinished === 0
                ? t("historyNone")
                : `${t("historyFactor")}: ${c.history_completed} ${t("completedCount")}, ${c.history_unfinished} ${t("unfinishedCount")}.`}
            </span>
          </p>
          {c.history_unfinished > 0 && <small>{t("historyNeutral")}</small>}
        </div>
      )}
      {compact && Object.keys(c.event.prerequisites).length > 0 && (
        <p className="small muted">
          {t("requirements")}:{" "}
          {Object.entries(c.event.prerequisites)
            .map(([id, level]) => `${name(id)} ≥ ${level}`)
            .join(", ")}
        </p>
      )}
      {c.blocked.length > 0 && (
        <ul className="blocked-reasons">
          {c.blocked.map((reason) => (
            <li key={reason}>{t(reason)}</li>
          ))}
        </ul>
      )}
      <button
        className={compact ? "button secondary full" : "button primary full"}
        disabled={busy || !canStart}
        onClick={() => onStart(c.event.event_id)}
      >
        {t("start")}
        <ArrowUpRight size={17} />
      </button>
    </article>
  );
}
