package cli

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddNoteDetails_Success(t *testing.T) {
	input := "hello\nworld\n\n"
	a := &App{
		reader: bufio.NewReader(strings.NewReader(input)),
	}

	got, err := a.addNoteDetails(context.Background())
	require.NoError(t, err)

	note, ok := got.(*models.Note)
	require.True(t, ok, "ожидали *models.Note")
	assert.Equal(t, "hello\nworld", note.Text)
}

func TestAddNoteDetails_EmptyNote(t *testing.T) {
	a := &App{
		reader: bufio.NewReader(strings.NewReader("\n\n")),
	}
	got, err := a.addNoteDetails(context.Background())
	require.NoError(t, err)

	note, ok := got.(*models.Note)
	require.True(t, ok)
	assert.Equal(t, "", note.Text)
}

func newReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

func TestAddCreditCardDetails_ReadsTwoFields(t *testing.T) {
	a := &App{reader: newReader("4111111111111111\n12/25\n")}
	item, err := a.addCreditCardDetails(context.Background())
	require.NoError(t, err)

	cc, ok := item.(*models.CreditCard)
	if !ok {
		if v, vok := any(item).(models.CreditCard); vok {
			assert.Equal(t, "4111111111111111", v.Number)
			assert.Equal(t, "12/25", v.Expiration)
			return
		}
		t.Fatalf("expected models.CreditCard, got %T", item)
	}
	assert.Equal(t, "4111111111111111", cc.Number)
	assert.Equal(t, "12/25", cc.Expiration)
}

func TestAddLoginDetails_ReadsThreeFields(t *testing.T) {
	a := &App{reader: newReader("user\npass\nhttps://ex.com\n")}
	item, err := a.addLoginDetails(context.Background())
	require.NoError(t, err)

	login, ok := item.(*models.Login)
	if !ok {
		if v, vok := any(item).(models.Login); vok {
			assert.Equal(t, "user", v.Username)
			assert.Equal(t, "pass", v.Password)
			assert.Equal(t, "https://ex.com", v.URL)
			return
		}
		t.Fatalf("expected models.Login, got %T", item)
	}
	assert.Equal(t, "user", login.Username)
	assert.Equal(t, "pass", login.Password)
	assert.Equal(t, "https://ex.com", login.URL)
}

func TestAddFileDetails_ReadsPath(t *testing.T) {
	a := &App{reader: newReader("/tmp/file.bin\n")}
	item, err := a.addFileDetails(context.Background())
	require.NoError(t, err)

	bf, ok := item.(*models.BinaryFile)
	if !ok {
		if v, vok := any(item).(models.BinaryFile); vok {
			assert.Equal(t, "/tmp/file.bin", v.Path)
			return
		}
		t.Fatalf("expected models.BinaryFile, got %T", item)
	}
	assert.Equal(t, "/tmp/file.bin", bf.Path)
}

func TestInputEnvelope_ErrorOnEmptyTitle(t *testing.T) {
	a := &App{}
	r := newReader("\n")

	called := false
	rest := func(ctx context.Context) (models.TypedEntry, error) {
		called = true
		return nil, nil
	}

	_, _, err := a.InputEnvelope(context.Background(), r, rest)
	require.Error(t, err)
	assert.EqualError(t, err, "title is required")
	assert.False(t, called, "rest must not be called when title is empty")
}

func TestInputEnvelope_ErrorOnCanceledContextBeforeRest(t *testing.T) {
	a := &App{}
	r := newReader("Some title\n")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	rest := func(ctx context.Context) (models.TypedEntry, error) {
		called = true
		return nil, nil
	}

	_, _, err := a.InputEnvelope(ctx, r, rest)
	require.Error(t, err)
	assert.False(t, called, "rest must not be called when ctx already canceled")
}

func TestInputEnvelope_PropagatesRestError(t *testing.T) {
	a := &App{}
	r := newReader("Title ok\n")

	restErr := errors.New("boom in rest")
	rest := func(ctx context.Context) (models.TypedEntry, error) {
		return nil, restErr
	}

	_, _, err := a.InputEnvelope(context.Background(), r, rest)
	require.Error(t, err)
	assert.ErrorIs(t, err, restErr)
}

func TestInputEnvelope_TitleTrimmed(t *testing.T) {
	a := &App{}
	r := newReader("   My Title   \n")

	rest := func(ctx context.Context) (models.TypedEntry, error) {
		return nil, errors.New("stop here")
	}

	_, _, err := a.InputEnvelope(context.Background(), r, rest)
	require.Error(t, err)
}

func TestInputEnvelope_ReadsFromProvidedReader(t *testing.T) {
	a := &App{reader: newReader("WRONG SOURCE\n")}

	r := newReader("Correct title\n")
	restCalled := false
	rest := func(ctx context.Context) (models.TypedEntry, error) {
		restCalled = true
		return nil, errors.New("stop")
	}

	_, _, err := a.InputEnvelope(context.Background(), r, rest)
	require.Error(t, err)
	assert.True(t, restCalled, "rest should be invoked")
}

func TestInputEnvelope_ContextDeadlineBeforeRest(t *testing.T) {
	a := &App{}
	r := newReader("t\n")

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	time.Sleep(time.Millisecond)

	_, _, err := a.InputEnvelope(ctx, r, func(ctx context.Context) (models.TypedEntry, error) {
		t.Fatal("rest must not be called after context is already timed out")
		return nil, nil
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
