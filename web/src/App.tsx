import { useWorkspace } from "./hooks/useWorkspace";
import {
  ArrowRight,
  ArrowUpRight,
  BookOpen,
  CheckCircle2,
  Compass,
  Database,
  LayoutDashboard,
  Moon,
  Sun,
  Sprout,
  Users,
  X,
  AlertCircle,
  Target,
  TrendingUp,
} from "lucide-react";
import type { Language } from "./types";
import Overview from "./views/Overview";
import Catalog from "./views/Catalog";
import People from "./views/People";
import Data from "./views/Data";

export default function App() {
  const {
    language,
    setLanguage,
    dark,
    setDark,
    summaries,
    datasetId,
    setDatasetId,
    dataset,
    setDataset,
    setEmployeeId,
    analysis,
    setAnalysis,
    page,
    setPage,
    initialLoading,
    busy,
    problem,
    setProblem,
    notice,
    setRefresh,
    t,
    selectedEmployee,
    update,
    run,
    createDemo,
    start,
    removed,
  } = useWorkspace();
  const nav = [
    { key: "overview", icon: LayoutDashboard },
    { key: "catalog", icon: BookOpen },
    { key: "hr", icon: Users },
    { key: "data", icon: Database },
  ] as const;
  const readyAnalysis =
    analysis?.employee.employee_id === selectedEmployee ? analysis : null;
  const props = dataset
    ? { dataset, language, t, busy, run, onUpdate: update }
    : null;

  return (
    <div className="app-shell">
      <a className="skip-link" href="#main">
        {t("overview")}
      </a>
      <aside className="sidebar">
        <a
          className="brand"
          href="#"
          onClick={(e) => {
            e.preventDefault();
            setPage("overview");
          }}
        >
          <span className="brand-mark">
            <Sprout size={25} />
          </span>
          <span>
            career<span className="brand-light">quest</span>
            <small>{t("workspace")}</small>
          </span>
        </a>
        <span className="nav-caption">WORKSPACE</span>
        <nav aria-label={t("workspace")}>
          {nav.map(({ key, icon: Icon }) => (
            <button
              key={key}
              className={page === key ? "nav-item active" : "nav-item"}
              aria-current={page === key ? "page" : undefined}
              disabled={!dataset && key !== "data" && key !== "overview"}
              onClick={() => setPage(key)}
            >
              <Icon size={19} />
              <span>{t(key)}</span>
              {page === key && <span className="nav-dot" />}
            </button>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <div className="local-card">
            <span className="live-dot" />
            <strong>{t("local")}</strong>
            <p>{t("demoOnly")}</p>
          </div>
          <span className="sidebar-footer">CAREER QUEST / 01</span>
        </div>
      </aside>
      <div className="main-shell">
        <header className="topbar">
          <div className="breadcrumb">
            <span>Career Quest</span>
            <span>/</span>
            <strong>{t(page)}</strong>
          </div>
          <div className="topbar-controls">
            <span className="synthetic-badge">
              <span />
              {t("synthetic")}
            </span>
            <select
              aria-label={t("language")}
              className="language-select"
              value={language}
              onChange={(e) => setLanguage(e.target.value as Language)}
            >
              <option value="ru">RU</option>
              <option value="kk">ҚАЗ</option>
              <option value="en">EN</option>
            </select>
            <button
              className="icon-button"
              aria-label={t("theme")}
              onClick={() => setDark(!dark)}
            >
              {dark ? <Sun size={20} /> : <Moon size={20} />}
            </button>
          </div>
        </header>
        {dataset && (
          <div className="workspace-selector">
            <label>
              {t("dataset")}
              <select
                value={datasetId}
                disabled={busy}
                onChange={(e) => {
                  setDataset(null);
                  setAnalysis(null);
                  setDatasetId(e.target.value);
                }}
              >
                {summaries.map((d) => (
                  <option value={d.id} key={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              {t("profile")}
              <select
                value={selectedEmployee}
                disabled={busy}
                onChange={(e) => {
                  setAnalysis(null);
                  setEmployeeId(e.target.value);
                }}
              >
                {dataset.bundle.employees.map((e) => (
                  <option value={e.employee_id} key={e.employee_id}>
                    {e.full_name} · {e.grade}
                  </option>
                ))}
              </select>
            </label>
            <span className="small muted revision-label">
              {t("revision")} {dataset.revision}
            </span>
          </div>
        )}
        <main id="main" tabIndex={-1}>
          {problem && (
            <div className="error-box" role="alert">
              <AlertCircle size={22} />
              <div>
                <strong>{t("errorTitle")}</strong>
                <p>{t(problem.code)}</p>
                {problem.request_id && (
                  <small>
                    Request ID: <code>{problem.request_id}</code>
                  </small>
                )}
                {problem.issues && (
                  <ul>
                    {problem.issues.map((issue, i) => (
                      <li key={i}>
                        <code>{issue.path}</code> — {t(issue.code)}
                        <small>{issue.detail}</small>
                      </li>
                    ))}
                  </ul>
                )}
                <button
                  className="text-button"
                  onClick={() => {
                    setProblem(null);
                    setAnalysis(null);
                    if (datasetId) setRefresh((n) => n + 1);
                    else window.location.reload();
                  }}
                >
                  {t("retry")}
                </button>
              </div>
              <button
                className="icon-button"
                aria-label={t("close")}
                onClick={() => setProblem(null)}
              >
                <X size={18} />
              </button>
            </div>
          )}
          {initialLoading ? (
            <div className="loading">
              <span className="spinner" />
              {t("loading")}
            </div>
          ) : !dataset && !datasetId && page !== "data" ? (
            <div className="welcome">
              <section className="welcome-hero">
                <div>
                  <p className="eyebrow">
                    <span className="live-dot" />
                    {t("welcomeTag")}
                  </p>
                  <h1>{t("welcome")}</h1>
                  <p className="intro">{t("intro")}</p>
                  <div className="button-row">
                    <button
                      className="button primary large"
                      disabled={busy}
                      onClick={createDemo}
                    >
                      {t("startDemo")}
                      <ArrowRight size={19} />
                    </button>
                    <button
                      className="button secondary large"
                      onClick={() => setPage("data")}
                    >
                      {t("importOwn")}
                      <ArrowUpRight size={18} />
                    </button>
                  </div>
                  <p className="small muted">{t("demoNote")}</p>
                </div>
                <div className="journey-art" aria-hidden="true">
                  <div className="orbit orbit-one" />
                  <div className="orbit orbit-two" />
                  <div className="art-core">
                    <Sprout size={50} />
                  </div>
                  <span className="art-chip art-chip-one">
                    <Target size={23} />
                    01
                  </span>
                  <span className="art-chip art-chip-two">
                    <TrendingUp size={23} />
                    02
                  </span>
                  <span className="art-chip art-chip-three">
                    <CheckCircle2 size={23} />
                    03
                  </span>
                  <span className="art-label">YOUR NEXT CHAPTER</span>
                </div>
              </section>
              <div className="welcome-features">
                {[
                  { icon: Target, title: "discover", body: "discoverBody" },
                  { icon: Compass, title: "act", body: "actBody" },
                  { icon: TrendingUp, title: "reflect", body: "reflectBody" },
                ].map(({ icon: Icon, title, body }, i) => (
                  <article key={title}>
                    <span className="feature-index">0{i + 1}</span>
                    <Icon size={25} />
                    <h2>{t(title)}</h2>
                    <p>{t(body)}</p>
                  </article>
                ))}
              </div>
            </div>
          ) : page === "data" ? (
            <Data
              dataset={dataset}
              language={language}
              t={t}
              busy={busy}
              run={run}
              onUpdate={update}
              onDemo={createDemo}
              onRemove={removed}
            />
          ) : (
            props &&
            (page === "hr" ? (
              <People {...props} />
            ) : readyAnalysis ? (
              page === "overview" ? (
                <Overview
                  key={`${dataset?.id}-${selectedEmployee}-${dataset?.revision}`}
                  {...props}
                  analysis={readyAnalysis}
                  onStart={start}
                />
              ) : (
                <Catalog {...props} analysis={readyAnalysis} onStart={start} />
              )
            ) : (
              <div className="loading">
                <span className="spinner" />
                {t("loading")}
              </div>
            ))
          )}
        </main>
        <footer className="main-footer">
          <span>Career Quest</span>
          <span>{t("auditNote")}</span>
        </footer>
      </div>
      {notice && (
        <div className="toast" role="status">
          <CheckCircle2 size={20} />
          {t(notice)}
        </div>
      )}
    </div>
  );
}
