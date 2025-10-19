package models

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetadataFromString_OK(t *testing.T) {
	in := []string{"a=1", "b=two", "name = value"}
	md, err := MetadataFromString(in)
	require.NoError(t, err)
	require.Len(t, md, 3)
	require.Equal(t, "a", md[0].Name)
	require.Equal(t, "1", md[0].Value)
	require.Equal(t, "b", md[1].Name)
	require.Equal(t, "two", md[1].Value)
	require.Equal(t, "name ", md[2].Name)
	require.Equal(t, " value", md[2].Value)
}

func TestMetadataFromString_ErrorOnMalformed(t *testing.T) {
	_, err := MetadataFromString([]string{"justname", "x=y", "a=b=c"})
	require.ErrorIs(t, err, ErrIncorrectMetadata)
}

func TestWrapUnwrap_Login(t *testing.T) {
	src := Login{Username: "u", Password: "p", URL: "https://ex"}
	env, err := Wrap(EntryTypeLogin, "title", []Metadata{{Name: "k", Value: "v"}}, src)
	require.NoError(t, err)

	out, err := env.Unwrap()
	require.NoError(t, err)

	got, ok := out.(Login)
	require.True(t, ok)
	require.Equal(t, src, got)
	require.Equal(t, Overview{Type: EntryTypeLogin, Title: "title"}, env.Overview())
}

func TestWrapUnwrap_Note(t *testing.T) {
	src := Note{Text: "hello"}
	env, err := Wrap(EntryTypeNote, "t", nil, src)
	require.NoError(t, err)

	out, err := env.Unwrap()
	require.NoError(t, err)
	got, ok := out.(Note)
	require.True(t, ok)
	require.Equal(t, src, got)
}

func TestWrapUnwrap_CreditCard(t *testing.T) {
	src := CreditCard{Number: "4111", Expiration: "12/25", CVV: "123", Holder: "John"}
	env, err := Wrap(EntryTypeCreditCard, "cc", nil, src)
	require.NoError(t, err)

	out, err := env.Unwrap()
	require.NoError(t, err)
	got, ok := out.(CreditCard)
	require.True(t, ok)
	require.Equal(t, src, got)
}

func TestWrapUnwrap_BinaryFile(t *testing.T) {
	src := BinaryFile{Path: "/tmp/file.bin"}
	env, err := Wrap(EntryTypeBinaryFile, "file", nil, src)
	require.NoError(t, err)

	out, err := env.Unwrap()
	require.NoError(t, err)
	got, ok := out.(BinaryFile)
	require.True(t, ok)
	require.Equal(t, src, got)
}

func TestUnwrap_UnknownType_ReturnsGenericMap(t *testing.T) {
	env := Envelope{
		Type:     EntryType("unknown"),
		Title:    "x",
		Metadata: nil,
		Details:  []byte(`{"a":1}`),
	}
	out, err := env.Unwrap()
	require.NoError(t, err)
	_, ok := out.(map[string]any)
	require.True(t, ok)
}

func TestBinaryFile_Materialize_CreatesEncryptedTempFile(t *testing.T) {
	wd := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(old) })
	require.NoError(t, os.Chdir(wd))

	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "plain.txt")
	require.NoError(t, os.WriteFile(srcPath, []byte("secret-data"), 0o600))

	bf := BinaryFile{Path: srcPath}
	f, err := bf.Materialize(context.Background())
	require.NoError(t, err)

	require.NotNil(t, f.EncryptedFileKey)
	require.NotNil(t, f.Nonce)
	require.NotEmpty(t, f.LocalPath)

	st, err := os.Stat(f.LocalPath)
	require.NoError(t, err)
	require.False(t, st.IsDir())
	require.Greater(t, st.Size(), int64(0))

	require.Contains(t, f.LocalPath, string(filepath.Separator)+"preupload"+string(filepath.Separator))
}
