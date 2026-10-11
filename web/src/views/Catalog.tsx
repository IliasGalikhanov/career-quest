import { useState } from "react";
import { Search } from "lucide-react";
import ActivityCard from "../components/ActivityCard";
import { localize } from "../i18n";
import type { Analysis } from "../types";
import type { ViewProps } from "./shared";

interface Props extends ViewProps {
  analysis: Analysis;
  onStart: (id: string) => void;
}

export default function Catalog({
  dataset,
  analysis,
  language,
  t,
  busy,
  onStart,
}: Props) {
  const [query, setQuery] = useState("");
  const [format, setFormat] = useState("any");
  const [available, setAvailable] = useState(false);
  const matches = analysis.catalog.filter(
    (c) =>
      (!query ||
        localize(c.event.titles, c.event.title, language)
          .toLocaleLowerCase()
          .includes(query.toLocaleLowerCase())) &&
      (format === "any" || c.event.format === format) &&
      (!available || c.blocked.length === 0),
  );
  return (
    <>
      <header className="page-heading">
        <div>
          <p className="eyebrow">{t("nextSteps")}</p>
          <h1>{t("catalog")}</h1>
          <p className="muted">{t("noRecommendationsBody")}</p>
        </div>
        <span className="count">
          {matches.length} / {analysis.catalog.length}
        </span>
      </header>
      <div className="filters">
        <label className="search-field">
          {t("search")}
          <Search size={16} />
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("search")}
          />
        </label>
        <label>
          {t("preferredFormat")}
          <select value={format} onChange={(e) => setFormat(e.target.value)}>
            {["any", "self_paced", "online", "offline"].map((f) => (
              <option key={f} value={f}>
                {t(f)}
              </option>
            ))}
          </select>
        </label>
        <label className="check-label">
          <input
            type="checkbox"
            checked={available}
            onChange={(e) => setAvailable(e.target.checked)}
          />
          {t("available")}
        </label>
      </div>
      {matches.length ? (
        <div className="cards-grid">
          {matches.map((c) => (
            <ActivityCard
              key={c.event.event_id}
              candidate={c}
              language={language}
              t={t}
              skills={dataset.bundle.skills}
              goal={analysis.goal}
              busy={busy}
              onStart={onStart}
              compact
            />
          ))}
        </div>
      ) : (
        <div className="empty-state">{t("searchEmpty")}</div>
      )}
    </>
  );
}
