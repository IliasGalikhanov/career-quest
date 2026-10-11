package store

import (
	"careerquest/internal/domain"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type reader interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func readEntities[T any](ctx context.Context, q reader, table, id string) ([]T, error) {
	return readRows[T](ctx, q, "SELECT payload FROM "+table+" WHERE dataset_id=? ORDER BY ordinal", id)
}
func readRows[T any](ctx context.Context, q reader, query string, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		var raw []byte
		var item T
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// get uses one read transaction so metadata and entity rows share a snapshot.
func get(ctx context.Context, q reader, id string) (Dataset, error) {
	return getFiltered(ctx, q, id, "", nil)
}
func getFiltered(ctx context.Context, q reader, id, filter string, args []any) (Dataset, error) {
	d := Dataset{ID: id}
	var meta, settings []byte
	err := q.QueryRowContext(ctx, "SELECT name,revision,meta,settings FROM datasets WHERE id=?", id).Scan(&d.Name, &d.Revision, &meta, &settings)
	if errors.Is(err, sql.ErrNoRows) {
		return d, ErrNotFound
	}
	if err != nil {
		return d, err
	}
	if err = json.Unmarshal(meta, &d.Bundle.Meta); err != nil {
		return d, err
	}
	if err = json.Unmarshal(settings, &d.Bundle.Settings); err != nil {
		return d, err
	}
	if d.Bundle.Skills, err = readEntities[domain.Skill](ctx, q, "skills", id); err != nil {
		return d, err
	}
	if d.Bundle.RoleProfiles, err = readEntities[domain.RoleProfile](ctx, q, "role_profiles", id); err != nil {
		return d, err
	}
	if d.Bundle.Events, err = readEntities[domain.Event](ctx, q, "events", id); err != nil {
		return d, err
	}
	if d.Bundle.Employees, err = readRows[domain.Employee](ctx, q, "SELECT payload FROM employees WHERE dataset_id=?"+filter+" ORDER BY ordinal", append([]any{id}, args...)...); err != nil {
		return d, err
	}
	if filter == " AND id=?" {
		d.Bundle.History, err = readRows[domain.Activity](ctx, q, "SELECT payload FROM activity_history WHERE dataset_id=? AND employee_id=? ORDER BY ordinal", id, args[0])
	} else {
		d.Bundle.History, err = readRows[domain.Activity](ctx, q, "SELECT payload FROM activity_history WHERE dataset_id=? AND employee_id IN (SELECT id FROM employees WHERE dataset_id=?"+filter+") ORDER BY ordinal", append([]any{id, id}, args...)...)
	}
	return d, err
}
func (s *Store) Get(ctx context.Context, id string) (Dataset, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Dataset{}, err
	}
	defer tx.Rollback()
	d, err := get(ctx, tx, id)
	if err != nil {
		return d, err
	}
	return d, tx.Commit()
}

type entity struct {
	id      string
	ordinal int
	payload string
}

func rowsOf[T any](items []T, key func(T) string) (map[string]entity, error) {
	out := make(map[string]entity, len(items))
	for i, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		id := key(item)
		out[id] = entity{id, i, string(raw)}
	}
	return out, nil
}

// syncEntities writes only changed records. Table names are private constants.
func syncEntities[T any](ctx context.Context, tx *sql.Tx, table, id string, old map[string]entity, after []T, key func(T) string) error {
	next, err := rowsOf(after, key)
	if err != nil {
		return err
	}
	for _, row := range next {
		if previous, ok := old[row.id]; ok && previous == row {
			continue
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO "+table+"(dataset_id,id,ordinal,payload) VALUES(?,?,?,?) ON CONFLICT(dataset_id,id) DO UPDATE SET ordinal=excluded.ordinal,payload=excluded.payload", id, row.id, row.ordinal, row.payload)
		if err != nil {
			return err
		}
	}
	for key := range old {
		if _, ok := next[key]; !ok {
			if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE dataset_id=? AND id=?", id, key); err != nil {
				return err
			}
		}
	}
	return nil
}
func persist(ctx context.Context, tx *sql.Tx, id string, old map[string]map[string]entity, b domain.Bundle) error {
	steps := []func() error{
		func() error {
			return syncEntities(ctx, tx, "skills", id, old["skills"], b.Skills, func(v domain.Skill) string { return v.ID })
		},
		func() error {
			return syncEntities(ctx, tx, "role_profiles", id, old["role_profiles"], b.RoleProfiles, func(v domain.RoleProfile) string {
				raw, _ := json.Marshal([]string{v.Role, v.Grade})
				return string(raw)
			})
		},
		func() error {
			return syncEntities(ctx, tx, "events", id, old["events"], b.Events, func(v domain.Event) string { return v.ID })
		},
		func() error {
			return syncEntities(ctx, tx, "employees", id, old["employees"], b.Employees, func(v domain.Employee) string { return v.ID })
		},
		func() error {
			return syncEntities(ctx, tx, "activity_history", id, old["activity_history"], b.History, func(v domain.Activity) string { return v.ID })
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) Create(ctx context.Context, b domain.Bundle) (Dataset, error) {
	d := Dataset{ID: NewID(), Name: b.Meta.Dataset, Revision: 1, Bundle: b}
	meta, err := json.Marshal(b.Meta)
	if err != nil {
		return d, err
	}
	settings, err := json.Marshal(b.Settings)
	if err != nil {
		return d, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return d, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO datasets(id,name,meta,settings) VALUES(?,?,?,?)", d.ID, d.Name, string(meta), string(settings)); err != nil {
		return d, err
	}
	if err = persist(ctx, tx, d.ID, nil, b); err != nil {
		return d, err
	}
	return d, tx.Commit()
}
func (s *Store) Update(ctx context.Context, id string, revision int, change func(*domain.Bundle) error) (Dataset, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Dataset{}, err
	}
	defer tx.Rollback()
	d, err := get(ctx, tx, id)
	if err != nil {
		return d, err
	}
	if d.Revision != revision {
		return d, ErrConflict
	}
	before, err := snapshot(d.Bundle)
	if err != nil {
		return d, err
	}
	if err = change(&d.Bundle); err != nil {
		return d, err
	}
	if err = persist(ctx, tx, id, before, d.Bundle); err != nil {
		return d, err
	}
	meta, err := json.Marshal(d.Bundle.Meta)
	if err != nil {
		return d, err
	}
	settings, err := json.Marshal(d.Bundle.Settings)
	if err != nil {
		return d, err
	}
	d.Revision++
	d.Name = d.Bundle.Meta.Dataset
	if _, err = tx.ExecContext(ctx, "UPDATE datasets SET name=?,revision=?,meta=?,settings=? WHERE id=?", d.Name, d.Revision, string(meta), string(settings), id); err != nil {
		return d, err
	}
	return d, tx.Commit()
}

func (s *Store) Delete(ctx context.Context, id string, revision int) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM datasets WHERE id=? AND revision=?", id, revision)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}

// EmployeeIDs is an indexed SQL filter, reused by HR instead of filtering names in Go.
func (s *Store) EmployeeIDs(ctx context.Context, id, department, role string) ([]string, error) {
	query := "SELECT id FROM employees WHERE dataset_id=?"
	args := []any{id}
	if department != "" {
		query += " AND department=?"
		args = append(args, department)
	}
	if role != "" {
		query += " AND role=?"
		args = append(args, role)
	}
	query += " ORDER BY ordinal"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var value string
		if err = rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("read employee ID: %w", err)
		}
		ids = append(ids, value)
	}
	return ids, rows.Err()
}

// Select loads only matching profiles and history, within one SQLite snapshot.
func (s *Store) Select(ctx context.Context, id, employee, department, role string) (Dataset, error) {
	filter := ""
	args := []any{}
	for _, pair := range [][2]string{{"id", employee}, {"department", department}, {"role", role}} {
		if pair[1] != "" {
			filter += " AND " + pair[0] + "=?"
			args = append(args, pair[1])
		}
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Dataset{}, err
	}
	defer tx.Rollback()
	d, err := getFiltered(ctx, tx, id, filter, args)
	if err != nil {
		return d, err
	}
	return d, tx.Commit()
}

// Capture immutable encoded rows before a domain function changes nested maps.
func snapshot(b domain.Bundle) (map[string]map[string]entity, error) {
	out := map[string]map[string]entity{}
	var err error
	if out["skills"], err = rowsOf(b.Skills, func(v domain.Skill) string { return v.ID }); err != nil {
		return nil, err
	}
	if out["role_profiles"], err = rowsOf(b.RoleProfiles, func(v domain.RoleProfile) string {
		raw, _ := json.Marshal([]string{v.Role, v.Grade})
		return string(raw)
	}); err != nil {
		return nil, err
	}
	if out["events"], err = rowsOf(b.Events, func(v domain.Event) string { return v.ID }); err != nil {
		return nil, err
	}
	if out["employees"], err = rowsOf(b.Employees, func(v domain.Employee) string { return v.ID }); err != nil {
		return nil, err
	}
	if out["activity_history"], err = rowsOf(b.History, func(v domain.Activity) string { return v.ID }); err != nil {
		return nil, err
	}
	return out, nil
}
