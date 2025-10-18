package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/dmitrijs2005/gophkeeper/internal/client/client"
	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/cryptox"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

// ---- helpers ----

func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:authsvc?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec(`
CREATE TABLE metadata (
  key   TEXT PRIMARY KEY,
  value BLOB NOT NULL
);
`)
	require.NoError(t, err)
	return db
}

func insertMeta(t *testing.T, db *sql.DB, k string, v []byte) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO metadata(key,value) VALUES(?,?)`, k, v)
	require.NoError(t, err)
}

func getMeta(t *testing.T, db *sql.DB, k string) []byte {
	t.Helper()
	var v []byte
	err := db.QueryRow(`SELECT value FROM metadata WHERE key=?`, k).Scan(&v)
	require.NoError(t, err)
	return v
}

// ---- fake client ----

// fakeClient реализует client.Client для юнит-тестов AuthService.
type fakeClient struct {
	// поведение/результаты
	CloseErr    error
	RegisterErr error

	GetSaltRet []byte
	GetSaltErr error

	LoginErr error

	PingErr error

	// Sync/MarkUploaded/GetPresignedGetURL — не используются этими тестами,
	// но должны соответствовать интерфейсу
	SyncErr               error
	MarkUploadedErr       error
	GetPresignedGetURLRet string
	GetPresignedGetURLErr error

	// для проверок аргументов
	LastRegisterUser string
	LastRegisterSalt []byte
	LastRegisterKey  []byte

	LastGetSaltUser string

	LastLoginUser string
	LastLoginKey  []byte
}

func (f *fakeClient) Close() error { return f.CloseErr }

func (f *fakeClient) Register(ctx context.Context, username string, salt []byte, key []byte) error {
	f.LastRegisterUser = username
	f.LastRegisterSalt = append([]byte(nil), salt...)
	f.LastRegisterKey = append([]byte(nil), key...)
	return f.RegisterErr
}

func (f *fakeClient) GetSalt(ctx context.Context, username string) ([]byte, error) {
	f.LastGetSaltUser = username
	return append([]byte(nil), f.GetSaltRet...), f.GetSaltErr
}

func (f *fakeClient) Login(ctx context.Context, username string, key []byte) error {
	f.LastLoginUser = username
	f.LastLoginKey = append([]byte(nil), key...)
	return f.LoginErr
}

func (f *fakeClient) Ping(ctx context.Context) error { return f.PingErr }

func (f *fakeClient) Sync(ctx context.Context, entries []*models.Entry, files []*models.File, maxVersion int64) (
	processed []*models.Entry, newEntries []*models.Entry, newFiles []*models.File, uploadTasks []*models.FileUploadTask, globalMax int64, err error,
) {
	return nil, nil, nil, nil, 0, f.SyncErr
}

func (f *fakeClient) MarkUploaded(ctx context.Context, entryID string) error {
	return f.MarkUploadedErr
}

func (f *fakeClient) GetPresignedGetURL(ctx context.Context, entryID string) (string, error) {
	return f.GetPresignedGetURLRet, f.GetPresignedGetURLErr
}

// ---- TESTS ----

func TestOfflineLogin_NoLocalData_CurrentBehaviorUnauthorized(t *testing.T) {
	db := setupDB(t) // пустая таблица metadata
	fc := &fakeClient{}
	svc := NewAuthService(fc, db)

	_, err := svc.OfflineLogin(context.Background(), "user@example.com", []byte("pass"))
	// По текущему коду: savedUsername == nil => string(nil) == "" => != username => ErrUnauthorized
	require.ErrorIs(t, err, client.ErrUnauthorized)
}

func TestOfflineLogin_UsernameMismatch_Unauthorized(t *testing.T) {
	db := setupDB(t)

	// seed offline data для другого пользователя
	insertMeta(t, db, "username", []byte("other"))
	insertMeta(t, db, "salt", []byte("salt"))
	insertMeta(t, db, "verifier", []byte{1, 2, 3})

	fc := &fakeClient{}
	svc := NewAuthService(fc, db)

	_, err := svc.OfflineLogin(context.Background(), "user", []byte("p"))
	require.ErrorIs(t, err, client.ErrUnauthorized)
}

func TestOfflineLogin_WrongPassword_Unauthorized(t *testing.T) {
	db := setupDB(t)

	// создаём валидные salt/verifier для пароля "correct"
	salt := []byte("salty")
	mk := cryptox.DeriveMasterKey([]byte("correct"), salt)
	ver := cryptox.MakeVerifier(mk)

	insertMeta(t, db, "username", []byte("user"))
	insertMeta(t, db, "salt", salt)
	insertMeta(t, db, "verifier", ver)

	fc := &fakeClient{}
	svc := NewAuthService(fc, db)

	_, err := svc.OfflineLogin(context.Background(), "user", []byte("wrong"))
	require.ErrorIs(t, err, client.ErrUnauthorized)
}

func TestOfflineLogin_Success_ReturnsMasterKey(t *testing.T) {
	db := setupDB(t)

	salt := []byte("salty")
	mk := cryptox.DeriveMasterKey([]byte("pass"), salt)
	ver := cryptox.MakeVerifier(mk)

	insertMeta(t, db, "username", []byte("user"))
	insertMeta(t, db, "salt", salt)
	insertMeta(t, db, "verifier", ver)

	fc := &fakeClient{}
	svc := NewAuthService(fc, db)

	got, err := svc.OfflineLogin(context.Background(), "user", []byte("pass"))
	require.NoError(t, err)
	require.Equal(t, mk, got)
}

func TestOnlineLogin_GetSaltError_Wrapped(t *testing.T) {
	db := setupDB(t)
	fc := &fakeClient{GetSaltErr: errors.New("network down")}
	svc := NewAuthService(fc, db)

	_, err := svc.OnlineLogin(context.Background(), "u", []byte("p"))
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "get salt error:"))
}

func TestOnlineLogin_LoginError_Wrapped(t *testing.T) {
	db := setupDB(t)
	salt := []byte("s")
	fc := &fakeClient{GetSaltRet: salt, LoginErr: errors.New("bad creds")}
	svc := NewAuthService(fc, db)

	_, err := svc.OnlineLogin(context.Background(), "u", []byte("p"))
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "login error:"))
}

func TestOnlineLogin_Success_SavesOfflineDataAndReturnsMasterKey(t *testing.T) {
	db := setupDB(t)
	fc := &fakeClient{GetSaltRet: []byte("salt")}
	svc := NewAuthService(fc, db)

	got, err := svc.OnlineLogin(context.Background(), "user", []byte("pass"))
	require.NoError(t, err)

	// проверяем, что в metadata легли username/salt/verifier
	require.Equal(t, []byte("user"), getMeta(t, db, "username"))
	require.Equal(t, []byte("salt"), getMeta(t, db, "salt"))
	savedVerifier := getMeta(t, db, "verifier")
	require.NotEmpty(t, savedVerifier)

	// и что мастер-ключ соответствует паролю/соли
	expected := cryptox.DeriveMasterKey([]byte("pass"), []byte("salt"))
	require.Equal(t, expected, got)

	// клиент получил логин с верным verifierCandidate
	require.Equal(t, "user", fc.LastLoginUser)
	require.Equal(t, savedVerifier, fc.LastLoginKey)
}

func TestRegister_DelegatesToClient(t *testing.T) {
	db := setupDB(t)
	fc := &fakeClient{}
	svc := NewAuthService(fc, db)

	err := svc.Register(context.Background(), "u", []byte("p"))
	require.NoError(t, err)

	require.Equal(t, "u", fc.LastRegisterUser)
	require.NotEmpty(t, fc.LastRegisterSalt)
	require.NotEmpty(t, fc.LastRegisterKey)
}

func TestPing_Close_ClearOfflineData_Delegations(t *testing.T) {
	db := setupDB(t)
	// seed что-нибудь и проверим Clear
	insertMeta(t, db, "x", []byte("y"))

	fc := &fakeClient{}
	svc := NewAuthService(fc, db)

	require.NoError(t, svc.Ping(context.Background()))

	require.NoError(t, svc.Close(context.Background()))

	require.NoError(t, svc.ClearOfflineData(context.Background()))
	// таблица очищена
	var n int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM metadata`).Scan(&n))
	require.Equal(t, 0, n)
}

func TestRegister_ErrorFromClient(t *testing.T) {
	db := setupDB(t)
	fc := &fakeClient{RegisterErr: errors.New("dup")}
	svc := NewAuthService(fc, db)
	err := svc.Register(context.Background(), "u", []byte("p"))
	require.Error(t, err)
}

func TestPing_ErrorPropagates(t *testing.T) {
	db := setupDB(t)
	fc := &fakeClient{PingErr: errors.New("down")}
	svc := NewAuthService(fc, db)
	err := svc.Ping(context.Background())
	require.Error(t, err)
}

func TestClose_ErrorPropagates(t *testing.T) {
	db := setupDB(t)
	fc := &fakeClient{CloseErr: errors.New("io")}
	svc := NewAuthService(fc, db)
	err := svc.Close(context.Background())
	require.Error(t, err)
}
