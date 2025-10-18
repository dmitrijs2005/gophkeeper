package files

import (
	"context"
	"database/sql"
	"testing"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
CREATE TABLE files (
  entry_id TEXT PRIMARY KEY,
  encrypted_file_key BLOB NOT NULL,
  nonce BLOB NOT NULL,
  local_path TEXT NOT NULL,
  upload_status TEXT NOT NULL,
  deleted INTEGER NOT NULL DEFAULT 0
);
`)
	require.NoError(t, err)
	return db
}

func TestCreateOrUpdate_InsertAndUpdate(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	f := &models.File{
		EntryID:          "e1",
		EncryptedFileKey: []byte("k1"),
		Nonce:            []byte("n1"),
		LocalPath:        "/tmp/one",
		UploadStatus:     "pending",
		Deleted:          false,
	}
	require.NoError(t, r.CreateOrUpdate(ctx, f))

	var key, nonce []byte
	var lp, us string
	var del int
	err := db.QueryRow(`SELECT encrypted_file_key, nonce, local_path, upload_status, deleted FROM files WHERE entry_id=?`, "e1").
		Scan(&key, &nonce, &lp, &us, &del)
	require.NoError(t, err)
	assert.Equal(t, []byte("k1"), key)
	assert.Equal(t, []byte("n1"), nonce)
	assert.Equal(t, "/tmp/one", lp)
	assert.Equal(t, "pending", us)
	assert.Equal(t, 0, del)

	// update той же записи
	f2 := &models.File{
		EntryID:          "e1",
		EncryptedFileKey: []byte("k2"),
		Nonce:            []byte("n2"),
		LocalPath:        "/tmp/two",
		UploadStatus:     "completed",
		Deleted:          true,
	}
	require.NoError(t, r.CreateOrUpdate(ctx, f2))

	err = db.QueryRow(`SELECT encrypted_file_key, nonce, local_path, upload_status, deleted FROM files WHERE entry_id=?`, "e1").
		Scan(&key, &nonce, &lp, &us, &del)
	require.NoError(t, err)
	assert.Equal(t, []byte("k2"), key)
	assert.Equal(t, []byte("n2"), nonce)
	assert.Equal(t, "/tmp/two", lp)
	assert.Equal(t, "completed", us)
	assert.Equal(t, 1, del)
}

func TestDeleteByEntryID_SuccessAndNotFound(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
	                   VALUES ('x', x'01', x'02', '/tmp/x', 'pending', 0)`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)

	require.NoError(t, r.DeleteByEntryID(ctx, "x"))

	err = r.DeleteByEntryID(ctx, "x")
	require.Error(t, err)
}

func TestGetByEntryID_SuccessAndNotFound(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
	                   VALUES ('ok', x'0A', x'0B', '/tmp/ok', 'pending', 0)`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)

	got, err := r.GetByEntryID(ctx, "ok")
	require.NoError(t, err)
	assert.Equal(t, "ok", got.EntryID)
	assert.Equal(t, []byte{0x0A}, got.EncryptedFileKey)
	assert.Equal(t, []byte{0x0B}, got.Nonce)
	assert.Equal(t, "/tmp/ok", got.LocalPath)
	assert.Equal(t, "pending", got.UploadStatus)
	assert.Equal(t, false, got.Deleted)

	_, err = r.GetByEntryID(ctx, "nope")
	require.Error(t, err)
}

func TestGetAllPendingUpload_OnlyPending(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted) VALUES
	  ('p1', x'01', x'11', '/tmp/p1', 'pending', 0),
	  ('p2', x'02', x'12', '/tmp/p2', 'pending', 0),
	  ('c1', x'03', x'13', '/tmp/c1', 'completed', 0),
	  ('d1', x'04', x'14', '/tmp/d1', 'pending', 1)
	`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)
	got, err := r.GetAllPendingUpload(ctx)
	require.NoError(t, err)

	ids := map[string]struct{}{}
	for _, f := range got {
		ids[f.EntryID] = struct{}{}
		require.NotNil(t, f.EncryptedFileKey)
		require.NotNil(t, f.Nonce)
	}
	assert.Equal(t, map[string]struct{}{"p1": {}, "p2": {}, "d1": {}}, ids)
}

func TestMarkUploaded_SuccessAndNotFound(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
	                   VALUES ('m1', x'FF', x'EE', '/tmp/m1', 'pending', 0)`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)

	require.NoError(t, r.MarkUploaded(ctx, "m1"))

	var status string
	err = db.QueryRow(`SELECT upload_status FROM files WHERE entry_id='m1'`).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "completed", status)

	err = r.MarkUploaded(ctx, "absent")
	require.Error(t, err)
}
