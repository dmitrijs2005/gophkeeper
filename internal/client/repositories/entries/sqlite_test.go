package entries

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
CREATE TABLE entries (
  id TEXT PRIMARY KEY,
  overview BLOB NOT NULL,
  nonce_overview BLOB NOT NULL,
  details BLOB NOT NULL,
  nonce_details BLOB NOT NULL,
  deleted INTEGER NOT NULL DEFAULT 0,
  pending INTEGER NOT NULL DEFAULT 0
);
`)
	require.NoError(t, err)

	return db
}

func TestCreateOrUpdate_InsertAndUpdate(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	// insert
	e1 := &models.Entry{
		Id:            "id1",
		Overview:      []byte("ov1"),
		NonceOverview: []byte("no1"),
		Details:       []byte("d1"),
		NonceDetails:  []byte("nd1"),
		Deleted:       false,
	}
	require.NoError(t, r.CreateOrUpdate(ctx, e1))

	var ov, no, d, nd []byte
	var del int
	err := db.QueryRow(`SELECT overview, nonce_overview, details, nonce_details, deleted FROM entries WHERE id=?`, "id1").
		Scan(&ov, &no, &d, &nd, &del)
	require.NoError(t, err)
	assert.Equal(t, []byte("ov1"), ov)
	assert.Equal(t, []byte("no1"), no)
	assert.Equal(t, []byte("d1"), d)
	assert.Equal(t, []byte("nd1"), nd)
	assert.Equal(t, 0, del)

	// update по тому же id
	e1b := &models.Entry{
		Id:            "id1",
		Overview:      []byte("ov2"),
		NonceOverview: []byte("no2"),
		Details:       []byte("d2"),
		NonceDetails:  []byte("nd2"),
		Deleted:       true,
	}
	require.NoError(t, r.CreateOrUpdate(ctx, e1b))

	err = db.QueryRow(`SELECT overview, nonce_overview, details, nonce_details, deleted FROM entries WHERE id=?`, "id1").
		Scan(&ov, &no, &d, &nd, &del)
	require.NoError(t, err)
	assert.Equal(t, []byte("ov2"), ov)
	assert.Equal(t, []byte("no2"), no)
	assert.Equal(t, []byte("d2"), d)
	assert.Equal(t, []byte("nd2"), nd)
	assert.Equal(t, 1, del) // updated
}

func TestGetAll_OnlyNotDeleted(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	// seed: два активных, один удалённый
	_, err := db.Exec(`INSERT INTO entries(id, overview, nonce_overview, details, nonce_details, deleted) VALUES
	  ('a', x'01', x'02', x'03', x'04', 0),
	  ('b', x'05', x'06', x'07', x'08', 0),
	  ('c', x'09', x'0A', x'0B', x'0C', 1)
	`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)
	got, err := r.GetAll(ctx)
	require.NoError(t, err)

	ids := make(map[string]struct{})
	for _, e := range got {
		ids[e.Id] = struct{}{}
		// метод выбирает только overview/nonce_overview — этого достаточно
		require.NotNil(t, e.Overview)
		require.NotNil(t, e.NonceOverview)
	}
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, ids)
}

func TestDeleteByID_SuccessAndNotFound(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO entries(id, overview, nonce_overview, details, nonce_details, deleted) 
	                   VALUES ('x', x'01', x'01', x'01', x'01', 0)`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)

	require.NoError(t, r.DeleteByID(ctx, "x"))

	err = r.DeleteByID(ctx, "x")
	require.Error(t, err)
}

func TestGetByID_SuccessAndNotFound(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO entries(id, overview, nonce_overview, details, nonce_details, deleted) 
	                   VALUES ('ok', x'01', x'02', x'aa', x'bb', 0)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO entries(id, overview, nonce_overview, details, nonce_details, deleted) 
	                   VALUES ('del', x'01', x'02', x'aa', x'bb', 1)`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)

	e, err := r.GetByID(ctx, "ok")
	require.NoError(t, err)
	require.Equal(t, []byte{0xaa}, e.Details)
	require.Equal(t, []byte{0xbb}, e.NonceDetails)

	_, err = r.GetByID(ctx, "del")
	require.Error(t, err)

	_, err = r.GetByID(ctx, "nope")
	require.Error(t, err)
}

func TestGetAllPending_ReturnsOnlyPending(t *testing.T) {
	db := setupDB(t)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO entries(id, overview, nonce_overview, details, nonce_details, deleted, pending) VALUES
	  ('p1', x'01', x'02', x'03', x'04', 0, 1),
	  ('p2', x'05', x'06', x'07', x'08', 0, 1),
	  ('n1', x'09', x'0A', x'0B', x'0C', 0, 0)
	`)
	require.NoError(t, err)

	r := NewSQLiteRepository(db)
	got, err := r.GetAllPending(ctx)
	require.NoError(t, err)

	ids := make(map[string]struct{})
	for _, e := range got {
		ids[e.Id] = struct{}{}
		require.NotNil(t, e.Overview)
		require.NotNil(t, e.NonceOverview)
		require.NotNil(t, e.Details)
		require.NotNil(t, e.NonceDetails)
	}
	assert.Equal(t, map[string]struct{}{"p1": {}, "p2": {}}, ids)
}
