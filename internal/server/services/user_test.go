package services

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"github.com/dmitrijs2005/gophkeeper/internal/dbx"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/files"
	refreshtokensrepo "github.com/dmitrijs2005/gophkeeper/internal/server/repositories/refreshtokens"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
	usersrepo "github.com/dmitrijs2005/gophkeeper/internal/server/repositories/users"
)

// --- helpers ---

func newSQLMockDB1(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	return db, mock
}

func newUserService(t *testing.T, db *sql.DB, rm repomanager.RepositoryManager) *UserService {
	t.Helper()
	cfg := &config.Config{
		SecretKey:                    "k",           // для JWT
		AccessTokenValidityDuration:  time.Hour,     // не критично
		RefreshTokenValidityDuration: 2 * time.Hour, // не критично
	}
	return NewUserService(db, rm, cfg)
}

type fakeUsersRepo1 struct {
	createOut *models.User
	createErr error

	getOut *models.User
	getErr error
}

func (f *fakeUsersRepo1) Create(ctx context.Context, u *models.User) (*models.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createOut, nil
}
func (f *fakeUsersRepo1) GetUserByLogin(ctx context.Context, userName string) (*models.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.getOut, nil
}

func (f *fakeUsersRepo1) IncrementCurrentVersion(context.Context, string) (int64, error) {
	return 0, nil
}

type fakeRefreshRepo struct {
	findOut *models.RefreshToken
	findErr error

	delErr error

	createErr error
}

func (f *fakeRefreshRepo) Create(ctx context.Context, userID string, token string, validity time.Duration) error {
	return f.createErr
}
func (f *fakeRefreshRepo) Find(ctx context.Context, token string) (*models.RefreshToken, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.findOut, nil
}
func (f *fakeRefreshRepo) Delete(ctx context.Context, token string) error {
	return f.delErr
}

type fakeRepoManager1 struct {
	u *fakeUsersRepo1
	r *fakeRefreshRepo
}

func (m *fakeRepoManager1) RunMigrations(context.Context, *sql.DB) error           { return nil }
func (m *fakeRepoManager1) Users(db dbx.DBTX) usersrepo.Repository                 { return m.u }
func (m *fakeRepoManager1) RefreshTokens(db dbx.DBTX) refreshtokensrepo.Repository { return m.r }

func (m *fakeRepoManager1) Entries(db dbx.DBTX) entries.Repository { return nil }
func (m *fakeRepoManager1) Files(db dbx.DBTX) files.Repository     { return nil }

func TestRefreshToken_Success(t *testing.T) {
	db, mock := newSQLMockDB1(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()

	rm := &fakeRepoManager1{
		r: &fakeRefreshRepo{
			findOut: &models.RefreshToken{UserID: "u1", Expires: time.Now().Add(10 * time.Minute)},
		},
	}
	s := newUserService(t, db, rm)

	pair, err := s.RefreshToken(context.Background(), "refresh-xyz")
	if err != nil {
		t.Fatalf("RefreshToken error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("empty tokens: %+v", pair)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRefreshToken_Expired(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	rm := &fakeRepoManager1{
		r: &fakeRefreshRepo{
			findOut: &models.RefreshToken{UserID: "u1", Expires: time.Now().Add(-1 * time.Minute)},
		},
	}
	s := newUserService(t, db, rm)

	_, err := s.RefreshToken(context.Background(), "r")
	if !errors.Is(err, common.ErrRefreshTokenExpired) {
		t.Fatalf("want ErrRefreshTokenExpired, got %v", err)
	}
}

func TestRefreshToken_FindErr(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	rm := &fakeRepoManager1{r: &fakeRefreshRepo{findErr: errBoom{}}}
	s := newUserService(t, db, rm)

	_, err := s.RefreshToken(context.Background(), "r")
	if err == nil || !regexp.MustCompile(`error searching refresh token: .*boom`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped find error, got %v", err)
	}
}

func TestRefreshToken_DeleteErr(t *testing.T) {
	db, mock := newSQLMockDB(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()

	rm := &fakeRepoManager1{
		r: &fakeRefreshRepo{
			findOut: &models.RefreshToken{UserID: "u1", Expires: time.Now().Add(10 * time.Minute)},
			delErr:  errBoom{},
		},
	}
	s := newUserService(t, db, rm)

	_, err := s.RefreshToken(context.Background(), "r")
	if err == nil || !regexp.MustCompile(`error deleting refresh token: .*boom`).MatchString(err.Error()) {
		t.Fatalf("expected wrapped delete error, got %v", err)
	}
}

func TestRefreshToken_GeneratePair_CreateErr(t *testing.T) {
	db, mock := newSQLMockDB(t)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectRollback()

	rm := &fakeRepoManager1{
		r: &fakeRefreshRepo{
			findOut:   &models.RefreshToken{UserID: "u1", Expires: time.Now().Add(10 * time.Minute)},
			createErr: errBoom{},
		},
	}
	s := newUserService(t, db, rm)

	_, err := s.RefreshToken(context.Background(), "r")

	if !errors.Is(err, common.ErrorInternal) && err != nil && !regexp.MustCompile(`error generating token pair:`).MatchString(err.Error()) {
		t.Fatalf("expected internal/generate error, got %v", err)
	}
}

func TestRegister_SuccessAndError(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	rmOK := &fakeRepoManager1{
		u: &fakeUsersRepo1{createOut: &models.User{ID: "42", UserName: "alice"}},
		r: &fakeRefreshRepo{},
	}
	sOK := newUserService(t, db, rmOK)
	u, err := sOK.Register(context.Background(), "alice", []byte("s"), []byte("v"))
	if err != nil || u.ID != "42" {
		t.Fatalf("Register ok: got (%v, %v)", u, err)
	}

	rmErr := &fakeRepoManager1{
		u: &fakeUsersRepo1{createErr: errBoom{}},
		r: &fakeRefreshRepo{},
	}
	sErr := newUserService(t, db, rmErr)
	_, err = sErr.Register(context.Background(), "bob", []byte("s"), []byte("v"))
	if err == nil || !regexp.MustCompile(`error creating user: .*boom`).MatchString(err.Error()) {
		t.Fatalf("Register expected wrapped error, got %v", err)
	}
}

func TestGetSalt_Found_NotFound_Internal(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	rmFound := &fakeRepoManager1{
		u: &fakeUsersRepo1{getOut: &models.User{Salt: []byte("SALT")}},
		r: &fakeRefreshRepo{},
	}
	s := newUserService(t, db, rmFound)
	salt, err := s.GetSalt(context.Background(), "alice")
	if err != nil || string(salt) != "SALT" {
		t.Fatalf("GetSalt found: got (%q, %v)", string(salt), err)
	}

	rmNF := &fakeRepoManager1{
		u: &fakeUsersRepo1{getErr: common.ErrorNotFound},
		r: &fakeRefreshRepo{},
	}
	s2 := newUserService(t, db, rmNF)
	salt2, err := s2.GetSalt(context.Background(), "ghost")
	if err != nil || len(salt2) != 32 {
		t.Fatalf("GetSalt not found: len=%d err=%v", len(salt2), err)
	}

	rmErr := &fakeRepoManager1{
		u: &fakeUsersRepo1{getErr: errBoom{}},
		r: &fakeRefreshRepo{},
	}
	s3 := newUserService(t, db, rmErr)
	_, err = s3.GetSalt(context.Background(), "xx")
	if !errors.Is(err, common.ErrorInternal) {
		t.Fatalf("GetSalt internal: want ErrorInternal, got %v", err)
	}
}

func TestLogin_Flows(t *testing.T) {
	db, _ := newSQLMockDB(t)
	defer db.Close()

	// not found → unauthorized
	rmNF := &fakeRepoManager1{
		u: &fakeUsersRepo1{getErr: common.ErrorNotFound},
		r: &fakeRefreshRepo{},
	}
	sNF := newUserService(t, db, rmNF)
	if _, err := sNF.Login(context.Background(), "ghost", []byte("x")); !errors.Is(err, common.ErrorUnauthorized) {
		t.Fatalf("notfound → unauthorized, got %v", err)
	}

	// internal error
	rmIE := &fakeRepoManager1{
		u: &fakeUsersRepo1{getErr: errBoom{}},
		r: &fakeRefreshRepo{},
	}
	sIE := newUserService(t, db, rmIE)
	if _, err := sIE.Login(context.Background(), "u", []byte("x")); !errors.Is(err, common.ErrorInternal) {
		t.Fatalf("internal → ErrorInternal, got %v", err)
	}

	// wrong verifier → unauthorized
	rmWV := &fakeRepoManager1{
		u: &fakeUsersRepo1{getOut: &models.User{ID: "u1", Verifier: []byte("right")}},
		r: &fakeRefreshRepo{},
	}
	sWV := newUserService(t, db, rmWV)
	if _, err := sWV.Login(context.Background(), "u", []byte("wrong")); !errors.Is(err, common.ErrorUnauthorized) {
		t.Fatalf("wrong verifier → unauthorized, got %v", err)
	}

	rmOK := &fakeRepoManager1{
		u: &fakeUsersRepo1{getOut: &models.User{ID: "u1", Verifier: []byte("right")}},
		r: &fakeRefreshRepo{},
	}
	sOK := newUserService(t, db, rmOK)
	pair, err := sOK.Login(context.Background(), "u", []byte("right"))
	if err != nil || pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("Login success: pair=%+v err=%v", pair, err)
	}
}
