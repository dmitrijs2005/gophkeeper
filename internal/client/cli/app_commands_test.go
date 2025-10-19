package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/client/services"
	"github.com/stretchr/testify/require"
)

// ------------ helpers ------------

func readerFromLines(lines ...string) *bufio.Reader {
	if len(lines) == 0 || lines[len(lines)-1] != "" {
		lines = append(lines, "")
	}
	return bufio.NewReader(strings.NewReader(strings.Join(lines, "\n")))
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	return b
}

func newTestApp(es services.EntryService, sc *bufio.Reader, mk []byte) *App {
	return &App{
		entryService: es,
		reader:       sc,
		masterKey:    mk,
	}
}

type fakeES struct {
	// Add
	addCount int
	addEnv   models.Envelope
	addFile  *models.File
	addMK    []byte
	addErr   error

	// List
	listMK  []byte
	listOut []models.ViewOverview
	listErr error

	// Get
	getID  string
	getMK  []byte
	getOut *models.Envelope
	getErr error

	// Delete
	delID  string
	delErr error

	// Sync
	syncCalled bool
	syncErr    error

	// Presigned GET
	getURLID  string
	getURL    string
	getURLErr error

	// GetFile
	getFileID  string
	getFile    *models.File
	getFileErr error
}

func (f *fakeES) Sync(ctx context.Context) error { f.syncCalled = true; return f.syncErr }
func (f *fakeES) List(ctx context.Context, masterKey []byte) ([]models.ViewOverview, error) {
	f.listMK = masterKey
	return f.listOut, f.listErr
}
func (f *fakeES) Add(ctx context.Context, env models.Envelope, file *models.File, masterKey []byte) error {
	f.addCount++
	f.addEnv = env
	f.addFile = file
	f.addMK = masterKey
	return f.addErr
}
func (f *fakeES) DeleteByID(ctx context.Context, id string) error { f.delID = id; return f.delErr }
func (f *fakeES) Get(ctx context.Context, id string, masterKey []byte) (*models.Envelope, error) {
	f.getID = id
	f.getMK = masterKey
	return f.getOut, f.getErr
}
func (f *fakeES) GetPresignedGetUrl(ctx context.Context, id string) (string, error) {
	f.getURLID = id
	if f.getURLErr != nil {
		return "", f.getURLErr
	}
	return f.getURL, nil
}
func (f *fakeES) GetFile(ctx context.Context, id string) (*models.File, error) {
	f.getFileID = id
	return f.getFile, f.getFileErr
}

// ------------ tests ------------

func TestAddNote_EnvelopeIsPassed(t *testing.T) {
	es := &fakeES{}
	r := readerFromLines(
		"My title",  // Title
		"Note body", // Text/Body
		"",
	)
	app := newTestApp(es, r, []byte("mk"))
	if err := app.AddNote(context.Background()); err != nil {
		t.Fatalf("AddNote err: %v", err)
	}

	if es.addCount != 1 {
		t.Fatalf("Add not called exactly once, got %d", es.addCount)
	}
	if len(es.addMK) == 0 {
		t.Fatalf("masterKey not propagated")
	}
	if es.addEnv.Type != models.EntryTypeNote {
		t.Fatalf("Envelope.Type: want TypeNote, got %v", es.addEnv.Type)
	}
	if es.addEnv.Title == "" || len(es.addEnv.Details) == 0 {
		t.Fatalf("Envelope must have Title and Details, got: %+v", es.addEnv)
	}
}

func TestAddLogin_EnvelopeIsPassed(t *testing.T) {
	es := &fakeES{}
	r := readerFromLines(
		"My login",            // Title
		"alice",               // Username
		"p@ss",                // Password
		"https://example.org", // URL
		"",
	)
	app := newTestApp(es, r, []byte("mk"))
	if err := app.AddLogin(context.Background()); err != nil {
		t.Fatalf("AddLogin err: %v", err)
	}

	if es.addCount != 1 || es.addEnv.Type != models.EntryTypeLogin {
		t.Fatalf("wrong Add call: count=%d type=%v", es.addCount, es.addEnv.Type)
	}
	if es.addEnv.Title == "" || len(es.addEnv.Details) == 0 {
		t.Fatalf("empty Envelope fields: %+v", es.addEnv)
	}
}

func TestAddCreditCard_EnvelopeIsPassed(t *testing.T) {
	es := &fakeES{}
	r := readerFromLines(
		"My card",          // Title
		"4111111111111111", // Card number
		"holder=John Doe",  // metadata
		"expires=10/29",    // metadata
		"",
	)
	app := newTestApp(es, r, []byte("mk"))

	if err := app.AddCreditCard(context.Background()); err != nil {
		t.Fatalf("AddCreditCard err: %v", err)
	}

	if es.addCount != 1 || es.addEnv.Type != models.EntryTypeCreditCard {
		t.Fatalf("wrong Add call: count=%d type=%v", es.addCount, es.addEnv.Type)
	}
	if es.addEnv.Title == "" || len(es.addEnv.Details) == 0 {
		t.Fatalf("empty Envelope fields: %+v", es.addEnv)
	}
}

func TestAddFile_PassesFileAndEnvelope(t *testing.T) {
	es := &fakeES{}

	dir := t.TempDir()
	fp := filepath.Join(dir, "file.bin")
	require.NoError(t, os.WriteFile(fp, []byte{1, 2, 3, 4}, 0o600))

	r := readerFromLines(
		"My file title", // Title
		fp,              // File path
		"",
	)
	app := newTestApp(es, r, []byte("mk"))
	if err := app.AddFile(context.Background()); err != nil {
		t.Fatalf("AddFile err: %v", err)
	}
	if es.addCount != 1 || es.addEnv.Type != models.EntryTypeBinaryFile {
		t.Fatalf("wrong Add call for file: count=%d type=%v", es.addCount, es.addEnv.Type)
	}
	if es.addEnv.Title == "" || len(es.addEnv.Details) == 0 {
		t.Fatalf("envelope must have Title+Details, got: %+v", es.addEnv)
	}
	if es.addFile == nil {
		t.Fatalf("file payload is nil")
	}
}

func TestList_OK(t *testing.T) {
	es := &fakeES{
		listOut: []models.ViewOverview{
			{Id: "1", Title: "A", Type: string(models.EntryTypeNote)},
			{Id: "2", Title: "B", Type: string(models.EntryTypeLogin)},
		},
	}
	app := newTestApp(es, nil, []byte("mk"))
	if err := app.List(context.Background()); err != nil {
		t.Fatalf("List err: %v", err)
	}
	if len(es.listMK) == 0 {
		t.Fatalf("masterKey not passed to List")
	}
}

func TestShow_Note_And_File(t *testing.T) {
	ctx := context.Background()
	es := &fakeES{}

	// --- 1) NOTE ---
	es.getOut = &models.Envelope{
		Type:    models.EntryTypeNote,
		Title:   "Note T",
		Details: mustJSON(t, models.Note{Text: "Body"}),
	}

	app := newTestApp(es, readerFromLines(
		"42",
		"",
	), nil)

	if err := app.Show(ctx); err != nil {
		t.Fatalf("Show(note) err: %v", err)
	}
	if es.getID != "42" {
		t.Fatalf("Get called with wrong id: %q", es.getID)
	}
	if es.getURLID != "" {
		t.Fatalf("unexpected presigned URL for note: %q", es.getURLID)
	}

}

func TestDelete_And_Sync_OK(t *testing.T) {
	es := &fakeES{}
	app := newTestApp(es, readerFromLines("777"), []byte("mk"))

	if err := app.Delete(context.Background()); err != nil {
		t.Fatalf("Delete err: %v", err)
	}
	if es.delID != "777" {
		t.Fatalf("DeleteByID called with wrong id: %q", es.delID)
	}

	if err := app.Sync(context.Background()); err != nil {
		t.Fatalf("Sync err: %v", err)
	}
	if !es.syncCalled {
		t.Fatalf("Sync not called")
	}
}

func TestShow_ErrorPropagates(t *testing.T) {
	es := &fakeES{getErr: errors.New("boom")}
	app := newTestApp(es, readerFromLines("id-err"), []byte("mk"))
	if err := app.Show(context.Background()); err == nil {
		t.Fatalf("want error from Get to propagate")
	}
}
