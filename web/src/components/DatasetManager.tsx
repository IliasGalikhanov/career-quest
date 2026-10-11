import { useState } from "react";
import { api, json } from "../api";
import type { Dataset } from "../types";
import type { T } from "../i18n";
import type { Run } from "../views/shared";

export default function DatasetManager({
  dataset,
  t,
  busy,
  run,
  onUpdate,
  onRemove,
}: {
  dataset: Dataset;
  t: T;
  busy: boolean;
  run: Run;
  onUpdate: (d: Dataset) => void;
  onRemove: (id: string) => void;
}) {
  const [name, setName] = useState(dataset.name);
  const [confirm, setConfirm] = useState(false);
  return (
    <section className="panel">
      <h2>{t("manageDataset")}</h2>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void run(
            async () =>
              onUpdate(
                await api<Dataset>(
                  `/datasets/${dataset.id}`,
                  json("PATCH", { revision: dataset.revision, name }),
                ),
              ),
            "success",
          );
        }}
      >
        <label>
          {t("datasetName")}
          <input
            value={name}
            maxLength={120}
            required
            onChange={(e) => setName(e.target.value)}
          />
        </label>
        <button className="button secondary" disabled={busy || !name.trim()}>
          {t("renameDataset")}
        </button>
      </form>
      {confirm ? (
        <div
          className="delete-confirm"
          role="group"
          aria-label={t("deleteDataset")}
        >
          <p>
            {t("deleteWarning")} <strong>{dataset.name}</strong>
          </p>
          <a href={`/api/datasets/${dataset.id}/export`} download>
            {t("backup")}
          </a>
          <div className="button-row">
            <button
              className="button secondary"
              disabled={busy}
              onClick={() => setConfirm(false)}
            >
              {t("cancelAction")}
            </button>
            <button
              className="button secondary"
              disabled={busy}
              onClick={() =>
                void run(async () => {
                  await api(
                    `/datasets/${dataset.id}`,
                    json("DELETE", { revision: dataset.revision }),
                  );
                  onRemove(dataset.id);
                }, "datasetDeleted")
              }
            >
              {t("confirmDelete")}
            </button>
          </div>
        </div>
      ) : (
        <button
          className="text-button"
          disabled={busy}
          onClick={() => setConfirm(true)}
        >
          {t("deleteDataset")}
        </button>
      )}
    </section>
  );
}
