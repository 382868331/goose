package goose_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

// TestTaskGOO05UpVsAllowMissing reproduces the inconsistency between `up`
// (without -allow-missing) and `up -allow-missing` when a migration between
// already-applied versions is added later (a "missing"/out-of-order migration).
//
// Scenario: versions 1,2,4,3 are applied to the database in that order, so the
// applied set is {1,2,3,4} but the applied version_id sequence is
// non-monotonic. A subsequent plain `up` must NOT re-apply an already-applied
// migration (e.g. version 4) and must NOT stop with an error; new migrations
// (version 5) must still be applied afterwards.
func TestTaskGOO05UpVsAllowMissing(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 5; i++ {
		f := filepath.Join(dir, fmt.Sprintf("%d_test.sql", i))
		content := fmt.Sprintf("-- +goose Up\nCREATE TABLE t%d (id INTEGER);\n\n-- +goose Down\nDROP TABLE t%d;\n", i, i)
		if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	db, err := goose.OpenDBWithDriver("sqlite", filepath.Join(t.TempDir(), "goose.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Ensure the goose version table exists.
	if _, err := goose.EnsureDBVersion(db); err != nil {
		t.Fatal(err)
	}

	// Apply 1, 2, then 4, then 3 (out-of-order): the applied set is {1,2,3,4}
	// but the last applied version is 3.
	migrations, err := goose.CollectMigrations(dir, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(migrations) != 4 {
		t.Fatalf("expected 4 migrations, got %d", len(migrations))
	}
	for _, m := range migrations[:2] {
		if err := m.Up(db); err != nil {
			t.Fatal(err)
		}
	}
	if err := migrations[3].Up(db); err != nil { // version 4
		t.Fatal(err)
	}
	if err := migrations[2].Up(db); err != nil { // version 3
		t.Fatal(err)
	}

	// Plain `up` must not re-apply migration 4 (already applied) and must not
	// stop with an error.
	if err := goose.UpTo(db, dir, 4); err != nil {
		t.Fatalf("up (without -allow-missing) unexpectedly failed: %v", err)
	}
	// `up -allow-missing` must also succeed.
	if err := goose.UpTo(db, dir, 4, goose.WithAllowMissing()); err != nil {
		t.Fatalf("up -allow-missing unexpectedly failed: %v", err)
	}

	// A brand-new migration (version 5) must still be applied by plain `up`.
	if err := goose.UpTo(db, dir, 5); err != nil {
		t.Fatalf("up (without -allow-missing) failed to apply new migration: %v", err)
	}
	version, err := goose.GetDBVersion(db)
	if err != nil {
		t.Fatal(err)
	}
	if version != 5 {
		t.Fatalf("expected current version 5, got %d", version)
	}
}
