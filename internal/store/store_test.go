package store

import (
	"careerquest/internal/demo"
	"careerquest/internal/domain"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLegacyMigrationAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := migrations.ReadFile("migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("CREATE TABLE schema_migrations(version TEXT PRIMARY KEY); INSERT INTO schema_migrations VALUES('001_initial.sql')"); err != nil {
		t.Fatal(err)
	}
	want := demo.Dataset()
	raw, _ := json.Marshal(want)
	if _, err = db.Exec("INSERT INTO datasets(id,name,revision,payload) VALUES('legacy',?,7,?)", want.Meta.Dataset, string(raw)); err != nil {
		t.Fatal(err)
	}
	db.Close()
	for i := 0; i < 2; i++ {
		s, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		d, err := s.Get(context.Background(), "legacy")
		if err != nil {
			t.Fatal(err)
		}
		if d.Revision != 7 || !reflect.DeepEqual(d.Bundle, want) {
			t.Fatal("migration changed domain state")
		}
		s.Close()
	}
}

func TestRollbackForeignKeysAndCascade(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	d, err := s.Create(ctx, demo.Dataset())
	if err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("cancel change")
	_, err = s.Update(ctx, d.ID, 1, func(b *domain.Bundle) error { b.Employees[0].Skills["architecture"] = 5; return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatal(err)
	}
	_, err = s.Update(ctx, d.ID, 1, func(b *domain.Bundle) error { b.History[0].EmployeeID = "missing"; return nil })
	if err == nil {
		t.Fatal("foreign key accepted unknown employee")
	}
	current, err := s.Get(ctx, d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(current, d) {
		t.Fatal("failed transaction altered data")
	}
	if err = s.Delete(ctx, d.ID, 2); !errors.Is(err, ErrConflict) {
		t.Fatal("stale delete accepted")
	}
	if err = s.Delete(ctx, d.ID, 1); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"employees", "events", "skills", "role_profiles", "activity_history", "recommendation_runs"} {
		var count int
		if err = s.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s: count=%d err=%v", table, count, err)
		}
	}
}

func TestIndexedFilterAndUnchangedRows(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	d, err := s.Create(ctx, demo.Dataset())
	if err != nil {
		t.Fatal(err)
	}
	// Audit actual writes: changing a goal must not rewrite catalogs or history.
	_, err = s.db.Exec("CREATE TABLE writes(table_name TEXT); CREATE TRIGGER event_write AFTER UPDATE ON events BEGIN INSERT INTO writes VALUES('events'); END; CREATE TRIGGER history_write AFTER UPDATE ON activity_history BEGIN INSERT INTO writes VALUES('history'); END;")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Update(ctx, d.ID, 1, func(b *domain.Bundle) error { b.Employees[0].Goal.Grade = "Lead"; return nil })
	if err != nil {
		t.Fatal(err)
	}
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM writes").Scan(&count)
	if count != 0 {
		t.Fatal("unchanged entities rewritten")
	}
	ids, err := s.EmployeeIDs(ctx, d.ID, "Platform", "Platform Engineer")
	if err != nil || len(ids) != 4 {
		t.Fatalf("filter: %v %v", ids, err)
	}
}
