package services

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func setupDBEntry(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:entrysvc?mode=memory&cache=shared")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS entries (
  id TEXT PRIMARY KEY,
  overview BLOB NOT NULL,
  nonce_overview BLOB NOT NULL,
  details BLOB NOT NULL,
  nonce_details BLOB NOT NULL,
  deleted INTEGER NOT NULL DEFAULT 0,
  pending INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS files (
  entry_id TEXT PRIMARY KEY,
  encrypted_file_key BLOB NOT NULL,
  nonce BLOB NOT NULL,
  local_path TEXT NOT NULL,
  upload_status TEXT NOT NULL,
  deleted INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS metadata (
  key   TEXT PRIMARY KEY,
  value BLOB NOT NULL
);
`)
	require.NoError(t, err)
	return db
}

type fakeClientEntry struct {
	client.Client

	// presets
	SyncProcessed   []*models.Entry
	SyncNewEntries  []*models.Entry
	SyncNewFiles    []*models.File
	SyncUploadTasks []*models.FileUploadTask
	SyncMaxVersion  int64
	SyncErr         error

	GetURL string
	URLerr error

	MarkUploadedIDs []string
}

func (f *fakeClientEntry) Sync(ctx context.Context, entries []*models.Entry, files []*models.File, maxVersion int64) ([]*models.Entry, []*models.Entry, []*models.File, []*models.FileUploadTask, int64, error) {
	return f.SyncProcessed, f.SyncNewEntries, f.SyncNewFiles, f.SyncUploadTasks, f.SyncMaxVersion, f.SyncErr
}
func (f *fakeClientEntry) GetPresignedGetURL(ctx context.Context, entryID string) (string, error) {
	return f.GetURL, f.URLerr
}
func (f *fakeClientEntry) MarkUploaded(ctx context.Context, entryID string) error {
	f.MarkUploadedIDs = append(f.MarkUploadedIDs, entryID)
	return nil
}
func (f *fakeClientEntry) Ping(context.Context) error                                { return nil }
func (f *fakeClientEntry) Close() error                                              { return nil }
func (f *fakeClientEntry) Register(ctx context.Context, u string, s, k []byte) error { return nil }
func (f *fakeClientEntry) GetSalt(ctx context.Context, u string) ([]byte, error)     { return nil, nil }
func (f *fakeClientEntry) Login(ctx context.Context, u string, k []byte) error       { return nil }

func oneRow[T any](t *testing.T, db *sql.DB, q string, args ...any) T {
	t.Helper()
	var out T
	require.NoError(t, db.QueryRow(q, args...).Scan(&out))
	return out
}

func TestAdd_Note_InsertsEntry(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{}
	svc := NewEntryService(fc, db)

	key := make([]byte, 32)
	for i := range key {
		key[i] = 1
	}

	env, err := models.Wrap(models.EntryTypeNote, "My Note", nil, models.Note{Text: "hello"})
	require.NoError(t, err)

	require.NoError(t, svc.Add(context.Background(), env, nil, key))

	var cnt int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&cnt))
	require.Equal(t, 1, cnt)
}

func TestAdd_WithFile_InsertsFilePending(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{}
	svc := NewEntryService(fc, db)

	key := make([]byte, 32)
	for i := range key {
		key[i] = 2
	}

	env, err := models.Wrap(models.EntryTypeBinaryFile, "Doc", nil, models.BinaryFile{Path: "/ignored"})
	require.NoError(t, err)

	file := &models.File{
		EncryptedFileKey: []byte("k"),
		Nonce:            []byte("n"),
		LocalPath:        "/tmp/pre-encrypted.bin",
	}

	require.NoError(t, svc.Add(context.Background(), env, file, key))

	var entryID string
	require.NoError(t, db.QueryRow(`SELECT id FROM entries LIMIT 1`).Scan(&entryID))

	var status string
	var deleted int
	require.NoError(t, db.QueryRow(`SELECT upload_status, deleted FROM files WHERE entry_id=?`, entryID).Scan(&status, &deleted))
	require.Equal(t, "pending", status)
	require.Equal(t, 0, deleted)
}

func TestList_DecryptsOverview(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{}
	svc := NewEntryService(fc, db)

	key := make([]byte, 32)
	env, err := models.Wrap(models.EntryTypeLogin, "GitHub", nil, models.Login{Username: "u", Password: "p", URL: "https://gh"})
	require.NoError(t, err)
	require.NoError(t, svc.Add(context.Background(), env, nil, key))

	items, err := svc.List(context.Background(), key)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "GitHub", items[0].Title)
	require.Equal(t, string(models.EntryTypeLogin), items[0].Type)
}

func TestDeleteByID_SetsDeleted(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{}
	svc := NewEntryService(fc, db)

	key := make([]byte, 32)
	env, _ := models.Wrap(models.EntryTypeNote, "T", nil, models.Note{Text: "x"})
	require.NoError(t, svc.Add(context.Background(), env, nil, key))

	var id string
	require.NoError(t, db.QueryRow(`SELECT id FROM entries LIMIT 1`).Scan(&id))

	require.NoError(t, svc.DeleteByID(context.Background(), id))

	var del int
	require.NoError(t, db.QueryRow(`SELECT deleted FROM entries WHERE id=?`, id).Scan(&del))
	require.Equal(t, 1, del)
}

func TestGet_ReturnsDecryptedEnvelope(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{}
	svc := NewEntryService(fc, db)

	key := make([]byte, 32)
	envIn, _ := models.Wrap(models.EntryTypeCreditCard, "Visa", nil, models.CreditCard{Number: "4111", Expiration: "12/25", CVV: "123", Holder: "John"})
	require.NoError(t, svc.Add(context.Background(), envIn, nil, key))

	var id string
	require.NoError(t, db.QueryRow(`SELECT id FROM entries LIMIT 1`).Scan(&id))

	envOut, err := svc.Get(context.Background(), id, key)
	require.NoError(t, err)
	require.Equal(t, "Visa", envOut.Title)
	require.Equal(t, models.EntryTypeCreditCard, envOut.Type)

	x, err := envOut.Unwrap()
	require.NoError(t, err)
	cc, ok := x.(models.CreditCard)
	require.True(t, ok)
	require.Equal(t, "4111", cc.Number)
}

func TestSync_UpsertsAndUpdatesVersion_NoUploads(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClientEntry{
		SyncProcessed: []*models.Entry{
			{Id: "p1", Overview: []byte("ov"), NonceOverview: []byte("no"), Details: []byte("d"), NonceDetails: []byte("nd")},
		},
		SyncNewEntries: []*models.Entry{
			{Id: "n1", Overview: []byte("ovN"), NonceOverview: []byte("noN"), Details: []byte("dN"), NonceDetails: []byte("ndN")},
		},
		SyncNewFiles: []*models.File{
			{EntryID: "n1", EncryptedFileKey: []byte("fk"), Nonce: []byte("fn"), LocalPath: "/tmp/a", UploadStatus: "completed"},
		},
		SyncUploadTasks: nil,
		SyncMaxVersion:  7,
	}
	svc := NewEntryService(fc, db)

	require.NoError(t, svc.Sync(context.Background()))

	cv := oneRow[string](t, db, `SELECT value FROM metadata WHERE key='current_version'`)
	require.Equal(t, "7", cv)

	var cnt int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM entries WHERE id IN ('p1','n1')`).Scan(&cnt))
	require.Equal(t, 2, cnt)

	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM files WHERE entry_id='n1'`).Scan(&cnt))
	require.Equal(t, 1, cnt)
}

func TestGetPresignedGetUrl_DelegatesToClient(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClientEntry{GetURL: "https://dl"}
	svc := NewEntryService(fc, db)

	url, err := svc.GetPresignedGetUrl(context.Background(), "e1")
	require.NoError(t, err)
	require.Equal(t, "https://dl", url)
}

func TestGetFile_ByEntryID(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{}
	svc := NewEntryService(fc, db)

	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
	                   VALUES ('e42', x'01', x'02', '/tmp/x', 'pending', 0)`)
	require.NoError(t, err)

	f, err := svc.GetFile(context.Background(), "e42")
	require.NoError(t, err)
	require.Equal(t, "e42", f.EntryID)
	require.Equal(t, "/tmp/x", f.LocalPath)
}

func makeTempFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(contents), 0o600))
	return p
}

// ---- TESTS ----

func TestSync_UploadsPendingFiles_Success(t *testing.T) {
	db := setupDBEntry(t)

	tmp := t.TempDir()
	local := makeTempFile(t, tmp, "data.bin", "HELLO")
	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
	                   VALUES ('e1', x'AA', x'BB', ?, 'pending', 0)`, local)
	require.NoError(t, err)

	var received []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := os.ReadFile(local)
		_ = b
		body, _ := io.ReadAll(r.Body)
		received = append(received, body...)
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	fc := &fakeClientEntry{
		SyncProcessed:   nil,
		SyncNewEntries:  nil,
		SyncNewFiles:    nil,
		SyncUploadTasks: []*models.FileUploadTask{{EntryID: "e1", URL: srv.URL}},
		SyncMaxVersion:  1,
	}
	svc := NewEntryService(fc, db)

	require.NoError(t, svc.Sync(context.Background()))

	var status string
	require.NoError(t, db.QueryRow(`SELECT upload_status FROM files WHERE entry_id='e1'`).Scan(&status))
	require.Equal(t, "completed", status)

	_, err = os.Stat(local)
	require.Error(t, err)
	require.True(t, errors.Is(err, fs.ErrNotExist))

	require.Equal(t, []string{"e1"}, fc.MarkUploadedIDs)

	require.NotEmpty(t, received)
}

func TestSync_UploadsPendingFiles_ErrorFromServer(t *testing.T) {
	db := setupDBEntry(t)

	tmp := t.TempDir()
	local := makeTempFile(t, tmp, "data.bin", "HELLO")
	_, err := db.Exec(`INSERT INTO files(entry_id, encrypted_file_key, nonce, local_path, upload_status, deleted)
	                   VALUES ('e2', x'AA', x'BB', ?, 'pending', 0)`, local)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", 500)
	}))
	t.Cleanup(srv.Close)

	fc := &fakeClientEntry{
		SyncUploadTasks: []*models.FileUploadTask{{EntryID: "e2", URL: srv.URL}},
		SyncMaxVersion:  2,
	}
	svc := NewEntryService(fc, db)

	err = svc.Sync(context.Background())
	require.Error(t, err)
}

func TestSync_ParseCurrentVersionError(t *testing.T) {
	db := setupDBEntry(t)
	_, err := db.Exec(`INSERT INTO metadata(key,value) VALUES ('current_version','oops')`)
	require.NoError(t, err)

	svc := NewEntryService(&fakeClient{}, db)
	err = svc.Sync(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse current_version")
}

func TestSync_ClientErrorPropagates(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClient{SyncErr: errors.New("server-down")}
	svc := NewEntryService(fc, db)

	err := svc.Sync(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "error client sync")
}

func TestList_DecryptErrorDoesNotFail(t *testing.T) {
	db := setupDBEntry(t)
	svc := NewEntryService(&fakeClient{}, db)

	key1 := bytes.Repeat([]byte{1}, 32)
	env, err := models.Wrap(models.EntryTypeNote, "ShouldNotDecrypt", nil, models.Note{Text: "secret"})
	require.NoError(t, err)
	require.NoError(t, svc.Add(context.Background(), env, nil, key1))

	key2 := bytes.Repeat([]byte{2}, 32)
	outs, err := svc.List(context.Background(), key2)
	require.NoError(t, err)
	require.Len(t, outs, 1)

	require.Empty(t, outs[0].Title)
	require.Empty(t, outs[0].Type)
}
func TestGet_ErrorDecrypting(t *testing.T) {
	db := setupDBEntry(t)
	svc := NewEntryService(&fakeClient{}, db)

	key1 := bytes.Repeat([]byte{1}, 32)
	env, _ := models.Wrap(models.EntryTypeNote, "S", nil, models.Note{Text: "x"})
	require.NoError(t, svc.Add(context.Background(), env, nil, key1))

	var id string
	require.NoError(t, db.QueryRow(`SELECT id FROM entries LIMIT 1`).Scan(&id))

	key2 := bytes.Repeat([]byte{2}, 32)
	_, err := svc.Get(context.Background(), id, key2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error decrypting entry")
}

func TestGetPresignedGetUrl_Error(t *testing.T) {
	db := setupDBEntry(t)
	fc := &fakeClientEntry{URLerr: errors.New("no")}
	svc := NewEntryService(fc, db)

	_, err := svc.GetPresignedGetUrl(context.Background(), "id")
	require.Error(t, err)
	require.Contains(t, err.Error(), "error get presigned url")
}

func TestDeleteByID_NotFoundError(t *testing.T) {
	db := setupDBEntry(t)
	svc := NewEntryService(&fakeClient{}, db)

	err := svc.DeleteByID(context.Background(), "absent")
	require.Error(t, err)
}
