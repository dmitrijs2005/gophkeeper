package client

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&n)
	if err != nil {
		t.Fatalf("tableExists query failed: %v", err)
	}
	return n > 0
}

func TestInitDatabase_CreatesDBAndGooseVersionTable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dsn := filepath.Join(t.TempDir(), "app.db")

	db, err := InitDatabase(ctx, dsn)
	if err != nil {
		t.Fatalf("InitDatabase error: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext failed: %v", err)
	}

	if !tableExists(t, db, "goose_db_version") {
		t.Fatalf("expected goose_db_version table to exist after migrations")
	}
}

func TestRunMigrations_IsIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dsn := filepath.Join(t.TempDir(), "app.db")

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("sql.Open error: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(ctx, db); err != nil {
		t.Fatalf("RunMigrations (first) error: %v", err)
	}

	if err := RunMigrations(ctx, db); err != nil {
		t.Fatalf("RunMigrations (second) should be idempotent, got error: %v", err)
	}

	if !tableExists(t, db, "goose_db_version") {
		t.Fatalf("expected goose_db_version table to exist after repeated migrations")
	}
}

func TestInitDatabase_UsesSqliteDriver(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dsn := filepath.Join(t.TempDir(), "app.db")

	db, err := InitDatabase(ctx, dsn)
	if err != nil {
		t.Fatalf("InitDatabase error: %v", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS __probe (id INTEGER PRIMARY KEY, v TEXT)`)
	if err != nil {
		t.Fatalf("create probe table failed: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO __probe(v) VALUES ('ok')`)
	if err != nil {
		t.Fatalf("insert probe failed: %v", err)
	}
	var got string
	if err := db.QueryRowContext(ctx, `SELECT v FROM __probe LIMIT 1`).Scan(&got); err != nil {
		t.Fatalf("select probe failed: %v", err)
	}
	if got != "ok" {
		t.Fatalf("unexpected probe value: %q", got)
	}
}
