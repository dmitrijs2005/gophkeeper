// Package files provides the client-side persistence layer for file metadata
// associated with vault entries.
//
// # Overview
//
// The package defines a Repository interface for creating, updating, querying,
// and marking upload state of File records (per-file key, nonce, local path,
// upload status, soft-delete). A SQLite-backed implementation (SQLiteRepository)
// persists data via a dbx.DBTX (*sql.DB or *sql.Tx).
//
// Key Types
//
//   - type Repository        — contract used by higher-level services
//   - type SQLiteRepository  — SQLite implementation over dbx.DBTX
//
// Typical Usage
//
//	repo := files.NewSQLiteRepository(db)
//	_ = repo.CreateOrUpdate(ctx, file)
//	f, _ := repo.GetByEntryID(ctx, entryID)
//	pend, _ := repo.GetAllPendingUpload(ctx)
//	_ = repo.MarkUploaded(ctx, entryID)
//
// See also: internal/client/models.File for field semantics.
package files
