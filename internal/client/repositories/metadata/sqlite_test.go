package metadata

import (
	"context"
	"database/sql"
	"testing"

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
CREATE TABLE metadata (
  key   TEXT PRIMARY KEY,
  value BLOB NOT NULL
);`)
	require.NoError(t, err)
	return db
}

func TestSetAndGet_InsertThenGet(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, r.Set(ctx, "k1", []byte{0x01, 0x02}))

	v, err := r.Get(ctx, "k1")
	require.NoError(t, err)
	require.Equal(t, []byte{0x01, 0x02}, v)
}

func TestGet_NotExists_ReturnsNilNil(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	v, err := r.Get(ctx, "absent")
	require.NoError(t, err)
	require.Nil(t, v) // контракт: (nil, nil) если нет строки
}

func TestSet_UpsertOverwritesValue(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, r.Set(ctx, "k", []byte("old")))
	require.NoError(t, r.Set(ctx, "k", []byte("new"))) // upsert

	v, err := r.Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, []byte("new"), v)
}

func TestList_ReturnsAllPairs(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, r.Set(ctx, "a", []byte{0xAA}))
	require.NoError(t, r.Set(ctx, "b", []byte{0xBB, 0xCC}))

	m, err := r.List(ctx)
	require.NoError(t, err)
	assert.Len(t, m, 2)
	assert.Equal(t, []byte{0xAA}, m["a"])
	assert.Equal(t, []byte{0xBB, 0xCC}, m["b"])
}

func TestDelete_RemovesKey_AndIsIdempotent(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, r.Set(ctx, "x", []byte{0x01}))
	require.NoError(t, r.Delete(ctx, "x"))

	// теперь Get вернёт (nil, nil)
	v, err := r.Get(ctx, "x")
	require.NoError(t, err)
	require.Nil(t, v)

	// повторное удаление не должно падать
	require.NoError(t, r.Delete(ctx, "x"))
}

func TestClear_RemovesAllKeys(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, r.Set(ctx, "a", []byte{1}))
	require.NoError(t, r.Set(ctx, "b", []byte{2}))
	require.NoError(t, r.Clear(ctx))

	m, err := r.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, m)
}

// отдельная инициализация: БЕЗ NOT NULL, чтобы можно было вставить NULL и сломать Scan
func setupDBNullable(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`CREATE TABLE metadata (key TEXT PRIMARY KEY, value BLOB);`)
	require.NoError(t, err)
	return db
}

func TestGet_DBErrorWrapped(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	// Закрываем БД, чтобы получить ошибку драйвера
	require.NoError(t, db.Close())

	v, err := r.Get(ctx, "k")
	require.Error(t, err)
	require.Nil(t, v)
	require.Contains(t, err.Error(), "failed to get metadata[k]")
}

func TestSet_DBErrorWrapped(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Close())

	err := r.Set(ctx, "k", []byte("v"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to set metadata[k]")
}

func TestDelete_DBErrorWrapped(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Close())

	err := r.Delete(ctx, "k")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to delete metadata[k]")
}

func TestClear_DBErrorWrapped(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Close())

	err := r.Clear(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to clear metadata")
}

func TestList_DBErrorWrapped(t *testing.T) {
	db := setupDB(t)
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Close())

	_, err := r.List(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list metadata")
}

func TestList_NullValueIsReturnedAsNil(t *testing.T) {
	db := setupDBNullable(t) // таблица без NOT NULL
	r := NewSQLiteRepository(db)
	ctx := context.Background()

	_, err := db.Exec(`INSERT INTO metadata(key, value) VALUES ('bad', NULL);`)
	require.NoError(t, err)

	m, err := r.List(ctx)
	require.NoError(t, err)
	// ожидаем, что значение просто nil
	v, ok := m["bad"]
	require.True(t, ok)
	require.Nil(t, v)
}
