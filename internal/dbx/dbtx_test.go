package dbx

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:dbx_tests?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS t (id INTEGER PRIMARY KEY, v TEXT);`)
	require.NoError(t, err)
	return db
}

func countRows(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM t`).Scan(&n))
	return n
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	db := setupDB(t)

	err := WithTx(context.Background(), db, nil, func(ctx context.Context, tx DBTX) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO t(v) VALUES ('ok')`)
		return err
	})
	require.NoError(t, err)
	require.Equal(t, 1, countRows(t, db), "must commit on success")
}

func TestWithTx_RollbackOnFnError(t *testing.T) {
	db := setupDB(t)

	err := WithTx(context.Background(), db, nil, func(ctx context.Context, tx DBTX) error {
		_, e := tx.ExecContext(ctx, `INSERT INTO t(v) VALUES ('fail')`)
		require.NoError(t, e)
		return errors.New("boom") // должно привести к ROLLBACK
	})
	require.Error(t, err)

	require.Equal(t, 0, countRows(t, db), "must rollback when fn returns error")
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	db := setupDB(t)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic to propagate")
		}
		require.Equal(t, 0, countRows(t, db), "must rollback on panic")
	}()

	_ = WithTx(context.Background(), db, nil, func(ctx context.Context, tx DBTX) error {
		_, e := tx.ExecContext(ctx, `INSERT INTO t(v) VALUES ('panic')`)
		require.NoError(t, e)
		panic("kaput")
	})
}

func TestWithTx_BeginError(t *testing.T) {
	db := setupDB(t)
	require.NoError(t, db.Close()) // ломаем соединение

	err := WithTx(context.Background(), db, nil, func(ctx context.Context, tx DBTX) error {
		return nil
	})
	require.Error(t, err, "begin should fail when DB is closed")
}
