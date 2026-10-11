-- Split the legacy aggregate without changing the portable backup format.
ALTER TABLE datasets ADD COLUMN meta TEXT NOT NULL DEFAULT '{}' CHECK(json_valid(meta));
ALTER TABLE datasets ADD COLUMN settings TEXT NOT NULL DEFAULT '{}' CHECK(json_valid(settings));
UPDATE datasets SET meta=json_extract(payload,'$.meta'), settings=json_extract(payload,'$.settings');
CREATE TABLE skills (
 dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
 id TEXT NOT NULL,
 ordinal INTEGER NOT NULL CHECK(ordinal>=0),
 payload TEXT NOT NULL CHECK(json_valid(payload)),
 PRIMARY KEY(dataset_id,id)
);
INSERT INTO skills(dataset_id,id,ordinal,payload) SELECT d.id,json_extract(j.value,'$.skill_id'),CAST(j.key AS INTEGER),j.value FROM datasets d,json_each(d.payload,'$.skills') j;
CREATE INDEX skills_order ON skills(dataset_id,ordinal);
CREATE TABLE role_profiles (
 dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
 id TEXT NOT NULL,
 ordinal INTEGER NOT NULL CHECK(ordinal>=0),
 payload TEXT NOT NULL CHECK(json_valid(payload)),
 PRIMARY KEY(dataset_id,id)
);
INSERT INTO role_profiles(dataset_id,id,ordinal,payload) SELECT d.id,json_array(json_extract(j.value,'$.role'),json_extract(j.value,'$.grade')),CAST(j.key AS INTEGER),j.value FROM datasets d,json_each(d.payload,'$.role_profiles') j;
CREATE INDEX role_profiles_order ON role_profiles(dataset_id,ordinal);
CREATE TABLE events (
 dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
 id TEXT NOT NULL,
 ordinal INTEGER NOT NULL CHECK(ordinal>=0),
 payload TEXT NOT NULL CHECK(json_valid(payload)),
 PRIMARY KEY(dataset_id,id)
);
INSERT INTO events(dataset_id,id,ordinal,payload) SELECT d.id,json_extract(j.value,'$.event_id'),CAST(j.key AS INTEGER),j.value FROM datasets d,json_each(d.payload,'$.events') j;
CREATE INDEX events_order ON events(dataset_id,ordinal);
CREATE TABLE employees (
 dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
 id TEXT NOT NULL,
 ordinal INTEGER NOT NULL CHECK(ordinal>=0),
 payload TEXT NOT NULL CHECK(json_valid(payload)),
 department TEXT GENERATED ALWAYS AS (json_extract(payload,'$.department')) STORED,
 role TEXT GENERATED ALWAYS AS (json_extract(payload,'$.role')) STORED,
 PRIMARY KEY(dataset_id,id)
);
INSERT INTO employees(dataset_id,id,ordinal,payload) SELECT d.id,json_extract(j.value,'$.employee_id'),CAST(j.key AS INTEGER),j.value FROM datasets d,json_each(d.payload,'$.employees') j;
CREATE INDEX employees_order ON employees(dataset_id,ordinal);
CREATE TABLE activity_history (
 dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
 id TEXT NOT NULL,
 ordinal INTEGER NOT NULL CHECK(ordinal>=0),
 payload TEXT NOT NULL CHECK(json_valid(payload)),
 employee_id TEXT GENERATED ALWAYS AS (json_extract(payload,'$.employee_id')) STORED NOT NULL,
 event_id TEXT GENERATED ALWAYS AS (json_extract(payload,'$.event_id')) STORED NOT NULL,
 status TEXT GENERATED ALWAYS AS (json_extract(payload,'$.status')) STORED NOT NULL CHECK(status IN ('completed','declined','dropped','in_progress','no_show','overdue')),
 PRIMARY KEY(dataset_id,id),
 FOREIGN KEY(dataset_id,employee_id) REFERENCES employees(dataset_id,id),
 FOREIGN KEY(dataset_id,event_id) REFERENCES events(dataset_id,id)
);
INSERT INTO activity_history(dataset_id,id,ordinal,payload) SELECT d.id,json_extract(j.value,'$.record_id'),CAST(j.key AS INTEGER),j.value FROM datasets d,json_each(d.payload,'$.activity_history') j;
CREATE INDEX activity_history_order ON activity_history(dataset_id,ordinal);
CREATE INDEX employees_department_role ON employees(dataset_id,department,role);
CREATE INDEX history_employee_status ON activity_history(dataset_id,employee_id,status);
CREATE INDEX history_event_status ON activity_history(dataset_id,event_id,status);
ALTER TABLE datasets DROP COLUMN payload;
