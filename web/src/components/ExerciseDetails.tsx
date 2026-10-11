import type { Event, Language } from "../types";
import { localize, type T } from "../i18n";

export default function ExerciseDetails({ event, language, t }: { event?: Event; language: Language; t: T }) {
  if (!event?.deliverable && !event?.criteria?.length) return null;
  return <details className="exercise-details">
    <summary>{t("exerciseDetails")}</summary>
    <strong>{t("deliverable")}</strong>
    <p>{localize(event.deliverable, "", language)}</p>
    <strong>{t("selfCheck")}</strong>
    <ul>{event.criteria?.map((item, i) => <li key={i}>{localize(item, "", language)}</li>)}</ul>
    <small>{t("selfReported")}</small>
  </details>;
}
