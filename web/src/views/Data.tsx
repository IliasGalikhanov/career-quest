import { useState } from "react";
import {
  UploadCloud,
  Download,
  CheckCircle2,
  SlidersHorizontal,
  Database,
} from "lucide-react";
import { api, json, ApiError } from "../api";
import type { T } from "../i18n";
import type { Dataset, Language, Preview } from "../types";
import type { Run } from "./shared";

interface Props {
  dataset: Dataset | null;
  language: Language;
  t: T;
  busy: boolean;
  run: Run;
  onUpdate: (d: Dataset) => void;
  onDemo: () => void;
}

export default function Data(props: Props) {
  const { dataset, t, onDemo, busy } = props;
  return (
    <>
      <header className="page-heading">
        <div>
          <p className="eyebrow">{t("data")}</p>
          <h1>{t("dataTitle")}</h1>
          <p className="muted">{t("dataBody")}</p>
        </div>
      </header>
      <div className="data-columns">
        <Importer key={`${dataset?.id}-${dataset?.revision}`} {...props} />
        <div className="form-stack">
          {dataset && (
            <SettingsForm
              key={`${dataset.id}-${dataset.revision}`}
              {...props}
              dataset={dataset}
            />
          )}
          <section className="panel" style={{ marginTop: 0 }}>
            <div className="section-heading">
              <h2>{t("backup")}</h2>
              <Database size={19} />
            </div>
            <div className="backup-actions">
              {dataset && (
                <>
                  <a
                    className="button secondary"
                    href={`/api/datasets/${dataset.id}/export`}
                    download
                  >
                    <Download size={16} />
                    {t("backup")}
                  </a>
                  <a
                    className="button secondary"
                    href={`/api/datasets/${dataset.id}/evidence`}
                    download
                  >
                    <Download size={16} />
                    {t("evidence")}
                  </a>
                </>
              )}
              <a className="button secondary" href="/api/example" download>
                <Download size={16} />
                {t("downloadExample")}
              </a>
              <button className="text-button" disabled={busy} onClick={onDemo}>
                {t("newDemo")}
              </button>
            </div>
          </section>
        </div>
      </div>
    </>
  );
}

function Importer({ dataset, t, busy, run, onUpdate }: Props) {
  const [files, setFiles] = useState<File[]>([]);
  const [synthetic, setSynthetic] = useState(false);
  const [append, setAppend] = useState(false);
  const [preview, setPreview] = useState<Preview | null>(null);
  const makeForm = () => {
    if (files.reduce((sum, file) => sum + file.size, 0) > 5 * 1024 * 1024)
      throw new ApiError({ code: "invalid_upload" });
    const form = new FormData();
    files.forEach((file) => form.append("files", file));
    form.set("synthetic", String(synthetic));
    if (append && dataset) {
      form.set("append_to", dataset.id);
      form.set("revision", String(dataset.revision));
    }
    return form;
  };
  return (
    <section className="panel" style={{ marginTop: 0 }}>
      <div className="section-heading">
        <h2>{t("upload")}</h2>
        <UploadCloud size={20} />
      </div>
      <p className="panel-intro">{t("uploadHelp")}</p>
      <div className="form-stack">
        <label className="upload-zone">
          <UploadCloud size={28} />
          {t("files")}
          <input
            type="file"
            multiple
            accept=".json,.csv"
            onChange={(e) => {
              setFiles(Array.from(e.target.files ?? []));
              setPreview(null);
            }}
          />
        </label>
        <label className="check-label">
          <input
            type="checkbox"
            checked={synthetic}
            onChange={(e) => {
              setSynthetic(e.target.checked);
              setPreview(null);
            }}
          />
          {t("syntheticConfirm")}
        </label>
        {dataset && (
          <div>
            <label className="check-label">
              <input
                type="checkbox"
                checked={append}
                onChange={(e) => {
                  setAppend(e.target.checked);
                  setPreview(null);
                }}
              />
              {t("append")}
            </label>
            {append && <p className="small muted">{t("appendHelp")}</p>}
          </div>
        )}
        <button
          className="button secondary"
          disabled={busy || !files.length || !synthetic}
          onClick={() =>
            void run(async () =>
              setPreview(
                await api<Preview>("/import/preview", {
                  method: "POST",
                  body: makeForm(),
                }),
              ),
            )
          }
        >
          {t("preview")}
        </button>
        {preview && (
          <div className="preview-box">
            <h3>
              <CheckCircle2 size={17} /> {t("previewReady")}
            </h3>
            <p className="small">
              {preview.name} · {preview.as_of_date}
            </p>
            <div className="preview-stats">
              <span>
                {preview.employees} {t("profiles")}
              </span>
              <span>
                {preview.events} {t("activities")}
              </span>
              <span>
                {preview.skills} {t("skills").toLocaleLowerCase()}
              </span>
              <span>
                {preview.history} {t("records")}
              </span>
            </div>
            {preview.uncertain_dates > 0 && (
              <p className="small">
                {preview.uncertain_dates} {t("uncertainDates")}
              </p>
            )}
          </div>
        )}
        {preview && (
          <button
            className="button primary"
            disabled={busy || !synthetic}
            onClick={() =>
              void run(async () => {
                const d = await api<Dataset>("/import/commit", {
                  method: "POST",
                  body: makeForm(),
                });
                onUpdate(d);
                setPreview(null);
                setFiles([]);
              }, "importSuccess")
            }
          >
            {t("commitImport")}
          </button>
        )}
      </div>
    </section>
  );
}

function SettingsForm({
  dataset,
  t,
  busy,
  run,
  onUpdate,
}: Props & { dataset: Dataset }) {
  const [settings, setSettings] = useState(dataset.bundle.settings);
  return (
    <section className="panel" style={{ marginTop: 0 }}>
      <div className="section-heading">
        <h2>{t("calculation")}</h2>
        <SlidersHorizontal size={19} />
      </div>
      <form
        className="form-stack"
        onSubmit={(e) => {
          e.preventDefault();
          void run(
            async () =>
              onUpdate(
                await api<Dataset>(
                  `/datasets/${dataset.id}/settings`,
                  json("PUT", { revision: dataset.revision, settings }),
                ),
              ),
            "success",
          );
        }}
      >
        <div className="form-grid">
          <label>
            {t("date")}
            <input
              type="date"
              required
              min={dataset.bundle.meta.as_of_date}
              value={settings.date}
              onChange={(e) =>
                setSettings({ ...settings, date: e.target.value })
              }
            />
          </label>
          <label>
            {t("maxHours")}
            <input
              type="number"
              required
              min="0.5"
              max="1000"
              step="0.5"
              value={settings.max_hours}
              onChange={(e) =>
                setSettings({ ...settings, max_hours: Number(e.target.value) })
              }
            />
          </label>
          <label>
            {t("horizon")}
            <input
              type="number"
              required
              min="1"
              max="730"
              value={settings.horizon_days}
              onChange={(e) =>
                setSettings({
                  ...settings,
                  horizon_days: Number(e.target.value),
                })
              }
            />
          </label>
          <label>
            {t("preferredFormat")}
            <select
              value={settings.format}
              onChange={(e) =>
                setSettings({ ...settings, format: e.target.value })
              }
            >
              {["any", "self_paced", "online", "offline"].map((f) => (
                <option key={f} value={f}>
                  {t(f)}
                </option>
              ))}
            </select>
          </label>
        </div>
        <p className="small muted">
          {t("fixedDate")} {t("snapshot")}: {dataset.bundle.meta.as_of_date}.
        </p>
        <div>
          <label className="check-label">
            <input
              type="checkbox"
              checked={settings.assume_completion_date}
              onChange={(e) =>
                setSettings({
                  ...settings,
                  assume_completion_date: e.target.checked,
                })
              }
            />
            {t("assumption")}
          </label>
          <p className="small muted">{t("assumptionHelp")}</p>
        </div>
        <button className="button primary" disabled={busy}>
          {t("save")}
        </button>
      </form>
    </section>
  );
}
