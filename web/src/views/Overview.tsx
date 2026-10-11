import { useState } from "react";
import {
  ArrowRight,
  ArrowUpRight,
  Flag,
  TrendingUp,
  Check,
  CircleDashed,
} from "lucide-react";
import { api, json } from "../api";
import { localize } from "../i18n";
import ActivityCard from "../components/ActivityCard";
import type { Analysis, Dataset } from "../types";
import type { ViewProps } from "./shared";

interface Props extends ViewProps {
  analysis: Analysis;
  onStart: (id: string) => void;
}

export default function Overview({
  dataset,
  analysis: a,
  language,
  t,
  busy,
  run,
  onUpdate,
  onStart,
}: Props) {
  const [editGoal, setEditGoal] = useState(false);
  const [role, setRole] = useState(a.goal?.target_role ?? a.employee.role);
  const [grade, setGrade] = useState(a.goal?.target_grade ?? a.employee.grade);
  const [historyLimit, setHistoryLimit] = useState(10);
  const name = (id: string) => {
    const skill = dataset.bundle.skills.find((s) => s.skill_id === id);
    return localize(skill?.names, skill?.name ?? id, language);
  };
  const title = (id: string) => {
    const event = dataset.bundle.events.find((e) => e.event_id === id);
    return localize(event?.titles, event?.title ?? id, language);
  };
  const roles = [...new Set(dataset.bundle.role_profiles.map((p) => p.role))];
  const grades = dataset.bundle.role_profiles
    .filter((p) => p.role === role)
    .map((p) => p.grade);
  const finish = (record: string, status: string) =>
    void run(
      async () => {
        onUpdate(
          await api<Dataset>(
            `/datasets/${dataset.id}/participation/${record}`,
            json("PUT", { revision: dataset.revision, status }),
          ),
        );
      },
      status === "completed" ? "finished" : "stopped",
    );
  const ongoing = a.history.filter((h) => h.status === "in_progress");
  return (
    <>
      <header className="page-heading">
        <div>
          <p className="eyebrow">
            {t("hello")} · {a.employee.full_name}
          </p>
          <h1>{t("nextTitle")}</h1>
          <p className="muted">{t("nextBody")}</p>
        </div>
        <span className="date-pill">
          {t("date")}
          <strong>{dataset.bundle.settings.date}</strong>
        </span>
      </header>
      <section className="trajectory-panel">
        <div className="trajectory-content">
          <span className="eyebrow">
            <Flag size={14} />
            {t("target")}
            {a.suggested_goal ? ` · ${t("suggested")}` : ""}
          </span>
          <div className="career-route">
            <div>
              <span>{a.employee.role}</span>
              <strong>{a.employee.grade}</strong>
            </div>
            <ArrowRight className="route-arrow" />
            <div>
              <span>{a.goal?.target_role ?? "—"}</span>
              <strong>{a.goal?.target_grade ?? "—"}</strong>
            </div>
          </div>
          <button
            className="text-button"
            onClick={() => setEditGoal(!editGoal)}
          >
            {t("changeGoal")}
            <ArrowUpRight size={16} />
          </button>
          {editGoal && (
            <form
              className="goal-form"
              onSubmit={(event) => {
                event.preventDefault();
                void run(async () => {
                  onUpdate(
                    await api<Dataset>(
                      `/datasets/${dataset.id}/employees/${a.employee.employee_id}/goal`,
                      json("PUT", {
                        revision: dataset.revision,
                        goal: { target_role: role, target_grade: grade },
                      }),
                    ),
                  );
                  setEditGoal(false);
                }, "success");
              }}
            >
              <label>
                {t("role")}
                <select
                  value={role}
                  onChange={(e) => {
                    setRole(e.target.value);
                    setGrade(
                      dataset.bundle.role_profiles.find(
                        (p) => p.role === e.target.value,
                      )?.grade ?? "Junior",
                    );
                  }}
                >
                  {roles.map((r) => (
                    <option key={r}>{r}</option>
                  ))}
                </select>
              </label>
              <label>
                {t("grade")}
                <select
                  value={grade}
                  onChange={(e) => setGrade(e.target.value)}
                >
                  {grades.map((g) => (
                    <option key={g}>{g}</option>
                  ))}
                </select>
              </label>
              <button className="button primary" disabled={busy}>
                {t("save")}
              </button>
            </form>
          )}
          {!a.goal && <p>{t("noGoal")}</p>}
        </div>
        <div className="coverage-block">
          <div
            className="coverage-ring"
            style={{
              background: `conic-gradient(var(--accent) ${(a.coverage ?? 0) * 3.6}deg, var(--track) 0)`,
            }}
          >
            <div>
              <strong data-testid="coverage">
                {a.coverage === null ? "—" : `${Math.round(a.coverage)}%`}
              </strong>
              <span>{t("estimated")}</span>
            </div>
          </div>
          <div>
            <h3>{t("coverage")}</h3>
            <p>
              {t("assessed")}:{" "}
              {a.assessed_coverage === null
                ? "—"
                : `${Math.round(a.assessed_coverage)}%`}
            </p>
            <small>{t("notPromotion")}</small>
          </div>
        </div>
      </section>
      <div className="stat-strip">
        <div>
          <span className="stat-symbol">
            <Flag size={19} />
          </span>
          <strong>
            {a.gaps.filter((g) => g.critical && g.current < g.required).length}
          </strong>
          <span>{t("criticalGaps")}</span>
        </div>
        <div>
          <span className="stat-symbol">
            <TrendingUp size={19} />
          </span>
          <strong>{a.recommendations.length}</strong>
          <span>{t("nextSteps")}</span>
        </div>
        <div>
          <span className="stat-symbol">
            <Check size={19} />
          </span>
          <strong>
            {a.history.filter((h) => h.status === "completed").length}
          </strong>
          <span>{t("completed")}</span>
        </div>
      </div>
      {a.uncertain_completions > 0 && (
        <p className="notice">
          {a.uncertain_completions} · {t("uncertain")}
        </p>
      )}
      {ongoing.length > 0 && (
        <section>
          <div className="section-heading">
            <h2>{t("inProgress")}</h2>
            <span className="count">{ongoing.length}</span>
          </div>
          <div className="ongoing-list">
            {ongoing.map((h) => (
              <article className="ongoing" key={h.record_id}>
                <CircleDashed size={23} />
                <div>
                  <h3>{title(h.event_id)}</h3>
                  <p className="small muted">
                    {h.session_date
                      ? `${t("session")}: ${h.session_date}`
                      : t("self_paced")}
                  </p>
                </div>
                <div className="button-row">
                  <button
                    className="button secondary"
                    disabled={busy}
                    onClick={() => finish(h.record_id, "dropped")}
                  >
                    {t("drop")}
                  </button>
                  <button
                    className="button primary"
                    disabled={
                      busy ||
                      (!!h.session_date &&
                        h.session_date > dataset.bundle.settings.date)
                    }
                    onClick={() => finish(h.record_id, "completed")}
                  >
                    {t("finish")}
                    <Check size={16} />
                  </button>
                </div>
              </article>
            ))}
          </div>
        </section>
      )}
      <section>
        <div className="section-heading">
          <h2>{t("recommendations")}</h2>
          <span className="subtle-tag">01 — 03</span>
        </div>
        {a.recommendations.length ? (
          <div className="cards-grid">
            {a.recommendations.map((c) => (
              <ActivityCard
                key={c.event.event_id}
                candidate={c}
                language={language}
                t={t}
                skills={dataset.bundle.skills}
                goal={a.goal}
                busy={busy}
                onStart={onStart}
              />
            ))}
          </div>
        ) : (
          <div className="empty-state">
            <CompassPlaceholder />
            <h3>{t("noRecommendations")}</h3>
            <p>{t("noRecommendationsBody")}</p>
          </div>
        )}
      </section>
      <section className="panel">
        <div className="section-heading">
          <h2>{t("skills")}</h2>
          <span className="small muted">
            {t("assessed")} · {a.employee.last_review_date}
          </span>
        </div>
        <div className="skill-table">
          <div className="skill-row table-labels">
            <span>{t("skill")}</span>
            <span>{t("assessed")}</span>
            <span>{t("estimated")}</span>
            <span>{t("required")}</span>
          </div>
          {a.gaps.map((g) => (
            <div className="skill-row" key={g.skill_id}>
              <div>
                <strong>{name(g.skill_id)}</strong>
                {g.critical && (
                  <span className="critical-tag">{t("critical")}</span>
                )}
                {g.missing_assessment && <small>{t("notAssessed")}</small>}
              </div>
              <span>{g.assessed}</span>
              <div className="skill-meter">
                <progress
                  aria-label={`${name(g.skill_id)} · ${t("skillProgress")}`}
                  max={5}
                  value={g.current}
                />
                <strong data-testid={`level-${g.skill_id}`}>{g.current}</strong>
              </div>
              <span className="target-level">{g.required}</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <div className="section-heading">
          <h2>{t("history")}</h2>
          <span className="small muted">
            {a.history.length} {t("records")}
          </span>
        </div>
        {a.history.length === 0 ? (
          <p className="muted">{t("noHistory")}</p>
        ) : (
          <>
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>{t("event")}</th>
                    <th>{t("historyDate")}</th>
                    <th>{t("status")}</th>
                  </tr>
                </thead>
                <tbody>
                  {a.history.slice(0, historyLimit).map((h) => (
                    <tr key={h.record_id}>
                      <td>{title(h.event_id)}</td>
                      <td>{h.completed_at ?? h.date}</td>
                      <td>
                        <span className={`status ${h.status}`}>
                          {t(h.status)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {historyLimit < a.history.length && (
              <button
                className="text-button"
                onClick={() => setHistoryLimit(historyLimit + 20)}
              >
                {t("showing")} {historyLimit} / {a.history.length} · +20
              </button>
            )}
          </>
        )}
      </section>
    </>
  );
}

function CompassPlaceholder() {
  return <TrendingUp size={28} />;
}
