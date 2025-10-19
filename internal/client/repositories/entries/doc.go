// Package entries provides the client-side persistence layer for vault entries.
//
// # Overview
//
// The package defines a Repository interface for CRUD and query operations on
// Entry models (see internal/client/models). A SQLite-backed implementation
// (SQLiteRepository) persists data using a dbx.DBTX (either *sql.DB or *sql.Tx).
//
// # Data Model
//
// Each entry stores encrypted fields (overview/details + nonces), a soft-delete
// flag (deleted), and may be marked pending for synchronization. Implementations
// typically return only overview fields for listings and full details for
// single-item reads.
//
// # Concurrency
//
// Implementations are expected to be safe for concurrent use when backed by a
// properly configured *sql.DB. When using *sql.Tx (DBTX), follow normal
// transaction scoping rules.
//
// Key Types
//
//   - type Repository        — interface used by higher-level services
//   - type SQLiteRepository  — SQLite implementation over dbx.DBTX
//
// Typical Usage
//
//	repo := entries.NewSQLiteRepository(db)
//	_ = repo.CreateOrUpdate(ctx, entry)
//	list, _ := repo.GetAll(ctx)
//	one, _ := repo.GetByID(ctx, id)
//	_ = repo.DeleteByID(ctx, id)
//	pend, _ := repo.GetAllPending(ctx)
//
// See also: internal/client/models for the Entry structure and encryption fields.
package entries
