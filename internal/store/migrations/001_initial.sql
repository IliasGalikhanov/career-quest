CREATE TABLE datasets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    revision INTEGER NOT NULL DEFAULT 1,
    payload TEXT NOT NULL CHECK(json_valid(payload)),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE recommendation_runs (
    id INTEGER PRIMARY KEY,
    dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    revision INTEGER NOT NULL,
    employee_id TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    payload TEXT NOT NULL CHECK(json_valid(payload)),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX recommendation_runs_dataset ON recommendation_runs(dataset_id, id);
