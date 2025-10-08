package cli

import (
	"bufio"
	"context"
	"strings"
	"testing"

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
