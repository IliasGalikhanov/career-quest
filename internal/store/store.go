// Package store persists versioned dataset aggregates and recommendation evidence.
package store

import (
	"careerquest/internal/domain"
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

var ErrConflict = errors.New("revision conflict")
var ErrNotFound = errors.New("dataset not found")

type Store struct{ db *sql.DB }
type Dataset struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Revision int           `json:"revision"`
	Bundle   domain.Bundle `json:"bundle"`
}
type Summary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Revision int    `json:"revision"`
}

func NewID() string {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(id[:])
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	for _, q := range []string{"PRAGMA foreign_keys=ON", "PRAGMA busy_timeout=5000", "PRAGMA journal_mode=WAL"} {
		if _, err = db.Exec(q); err != nil {
			db.Close()
			return nil, err
		}
	}
	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)"); err != nil {
		db.Close()
		return nil, err
	}
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		db.Close()
		return nil, err
	}
	for _, entry := range entries {
		var exists int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=?", entry.Name()).Scan(&exists)
		if err != nil {
			db.Close()
			return nil, err
		}
		if exists > 0 {
			continue
		}
		data, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			db.Close()
			return nil, err
		}
		tx, err := db.Begin()
		if err != nil {
			db.Close()
			return nil, err
		}
		if _, err = tx.Exec(string(data)); err == nil {
			_, err = tx.Exec("INSERT INTO schema_migrations(version) VALUES(?)", entry.Name())
		}
		if err != nil {
			tx.Rollback()
			db.Close()
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &Store{db: db}, nil
}
func (s *Store) Close() error                   { return s.db.Close() }
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *Store) List(ctx context.Context) ([]Summary, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id,name,revision FROM datasets ORDER BY created_at,id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Summary{}
	for rows.Next() {
		var d Summary
		if err = rows.Scan(&d.ID, &d.Name, &d.Revision); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (s *Store) SaveRun(ctx context.Context, d Dataset, a domain.Analysis) error {
	data, err := json.Marshal(struct {
		Settings domain.Settings `json:"settings"`
		Analysis domain.Analysis `json:"analysis"`
	}{d.Bundle.Settings, a})
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, "INSERT INTO recommendation_runs(dataset_id,revision,employee_id,algorithm,payload) VALUES(?,?,?,?,?)", d.ID, d.Revision, a.Employee.ID, a.Algorithm, string(data))
	if err != nil {
		return err
	}
	// Keep a bounded local audit trail rather than growing forever on refresh.
	_, err = tx.ExecContext(ctx, "DELETE FROM recommendation_runs WHERE dataset_id=? AND id NOT IN (SELECT id FROM recommendation_runs WHERE dataset_id=? ORDER BY id DESC LIMIT 500)", d.ID, d.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) Runs(ctx context.Context, id string) ([]json.RawMessage, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT json_object('revision',revision,'employee_id',employee_id,'algorithm',algorithm,'created_at',created_at,'evidence',json(payload)) FROM recommendation_runs WHERE dataset_id=? ORDER BY id DESC LIMIT 500", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("read evidence: %w", err)
		}
		out = append(out, json.RawMessage(raw))
	}
	return out, rows.Err()
}
