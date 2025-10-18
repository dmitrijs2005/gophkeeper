// Package dbx provides tiny DB abstractions shared by repositories:
// a minimal interface (DBTX) implemented by both *sql.DB and *sql.Tx,
// and a helper to run functions inside a transaction.
package dbx

import (
	"context"
	"database/sql"
)

// DBTX is the subset of database/sql used by our repos.
// Both *sql.DB and *sql.Tx satisfy this interface.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// WithTx begins a transaction, runs fn with a transactional handle, and then
// commits on success or rolls back on error/panic. Panics are rethrown.
//
// Typical use:
//
//	err := dbx.WithTx(ctx, db, nil, func(ctx context.Context, tx dbx.DBTX) error {
//	    // use tx instead of db
//	    _, err := tx.ExecContext(ctx, "UPDATE ...")
//	    return err
//	})
func WithTx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, fn func(ctx context.Context, tx DBTX) error) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = fn(ctx, tx)
	return err
}
