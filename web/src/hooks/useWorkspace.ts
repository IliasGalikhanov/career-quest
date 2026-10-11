import { useEffect, useState } from "react";
import { api, json, problemOf } from "../api";
import { translator } from "../i18n";
import type { Analysis, Dataset, Language, Problem, Summary } from "../types";
import type { Run } from "../views/shared";
import { useSection } from "./useSection";
function saved(key: string, fallback: string) {
  try {
    return localStorage.getItem(key) ?? fallback;
  } catch {
    return fallback;
  }
}
function save(key: string, value: string) {
  try {
    localStorage.setItem(key, value);
  } catch {
    /* The workspace remains usable without browser storage. */
  }
}

export function useWorkspace() {
  const [language, setLanguage] = useState<Language>(() => {
    const value = saved("cq-language", "ru");
    return value === "en" || value === "kk" ? value : "ru";
  });
  const [dark, setDark] = useState(() => saved("cq-theme", "light") === "dark");
  const [summaries, setSummaries] = useState<Summary[]>([]);
  const [datasetId, setDatasetId] = useState("");
  const [dataset, setDataset] = useState<Dataset | null>(null);
  const [employeeId, setEmployeeId] = useState(() => saved("cq-employee", ""));
  useEffect(() => {
    save("cq-employee", employeeId);
  }, [employeeId]);
  const [analysis, setAnalysis] = useState<Analysis | null>(null);
  const [page, setPage] = useSection();
  const [initialLoading, setInitialLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState<Problem | null>(null);
  const [notice, setNotice] = useState("");
  const [refresh, setRefresh] = useState(0);
  const t = translator(language);
  const selectedEmployee = dataset?.bundle.employees.some(
    (e) => e.employee_id === employeeId,
  )
    ? employeeId
    : (dataset?.bundle.employees[0]?.employee_id ?? "");

  useEffect(() => {
    document.documentElement.lang = language;
    save("cq-language", language);
  }, [language]);
  useEffect(() => {
    document.documentElement.dataset.theme = dark ? "dark" : "light";
    save("cq-theme", dark ? "dark" : "light");
  }, [dark]);
  useEffect(() => {
    if (!notice) return;
    const timeout = setTimeout(() => setNotice(""), 4500);
    return () => clearTimeout(timeout);
  }, [notice]);
  useEffect(() => {
    let active = true;
    api<Summary[]>("/datasets")
      .then((rows) => {
        if (active) {
          setSummaries(rows);
          const remembered = saved("cq-dataset", "");
          setDatasetId(
            rows.some((r) => r.id === remembered)
              ? remembered
              : (rows[0]?.id ?? ""),
          );
          setInitialLoading(false);
        }
      })
      .catch((error) => {
        if (active) {
          setProblem(problemOf(error));
          setInitialLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, []);
  useEffect(() => {
    if (!datasetId) return;
    let active = true;
    api<Dataset>(`/datasets/${datasetId}`)
      .then((d) => {
        if (active) {
          setDataset((current) =>
            current?.id === d.id && current.revision > d.revision ? current : d,
          );
          save("cq-dataset", d.id);
        }
      })
      .catch((error) => {
        if (active) setProblem(problemOf(error));
      });
    return () => {
      active = false;
    };
  }, [datasetId, refresh]);
  useEffect(() => {
    if (!dataset || !selectedEmployee) return;
    let active = true;
    api<Analysis>(
      `/datasets/${dataset.id}/employees/${selectedEmployee}/analysis`,
    )
      .then((a) => {
        if (active) setAnalysis(a);
      })
      .catch((error) => {
        if (active) setProblem(problemOf(error));
      });
    return () => {
      active = false;
    };
  }, [dataset, selectedEmployee]);

  const update = (d: Dataset) => {
    setAnalysis(null);
    setDataset(d);
    setDatasetId(d.id);
    save("cq-dataset", d.id);
    setSummaries((rows) => [
      ...rows.filter((row) => row.id !== d.id),
      { id: d.id, name: d.name, revision: d.revision },
    ]);
  };
  const run: Run = async (action, message) => {
    setBusy(true);
    setProblem(null);
    try {
      await action();
      if (message) setNotice(message);
    } catch (error) {
      setProblem(problemOf(error));
    } finally {
      setBusy(false);
    }
  };
  const createDemo = () =>
    void run(async () => {
      update(await api<Dataset>("/demo", { method: "POST" }));
      setPage("overview");
    });
  const start = (event: string) =>
    void run(async () => {
      if (!dataset) return;
      update(
        await api<Dataset>(
          `/datasets/${dataset.id}/participation`,
          json("POST", {
            revision: dataset.revision,
            employee_id: selectedEmployee,
            event_id: event,
          }),
        ),
      );
    }, "started");
  const removed = (id: string) => {
    const remaining = summaries.filter((item) => item.id !== id);
    setSummaries(remaining);
    setDataset(null);
    setAnalysis(null);
    setDatasetId(remaining[0]?.id ?? "");
    save("cq-dataset", remaining[0]?.id ?? "");
  };
  return {
    language,
    setLanguage,
    dark,
    setDark,
    summaries,
    datasetId,
    setDatasetId,
    dataset,
    setDataset,
    employeeId,
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
  };
}
