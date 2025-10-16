package filex

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	return func() { _ = os.Chdir(old) }
}

func TestEnsureSubdDir_CreatesDirectoryInCWD(t *testing.T) {
	tmp := t.TempDir()
	defer chdir(t, tmp)()

	got, err := EnsureSubdDir("preupload")
	require.NoError(t, err)

	want := filepath.Join(tmp, "preupload")
	require.Equal(t, want, got)

	fi, err := os.Stat(want)
	require.NoError(t, err)
	require.True(t, fi.IsDir(), "should create a directory")

	if runtime.GOOS != "windows" {
		perm := fi.Mode().Perm()
		require.Equal(t, os.FileMode(0o700), perm&0o700)
	}
}

func TestEnsureSubdDir_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	defer chdir(t, tmp)()

	first, err := EnsureSubdDir("preupload")
	require.NoError(t, err)

	second, err := EnsureSubdDir("preupload")
	require.NoError(t, err)

	require.Equal(t, first, second)
	fi, err := os.Stat(second)
	require.NoError(t, err)
	require.True(t, fi.IsDir())
}

func TestEnsureSubdDir_FailsIfFileWithSameNameExists(t *testing.T) {
	tmp := t.TempDir()
	defer chdir(t, tmp)()

	require.NoError(t, os.WriteFile("preupload", []byte("x"), 0o660))

	_, err := EnsureSubdDir("preupload")
	require.Error(t, err, "should fail when a file exists with the same name")
}
