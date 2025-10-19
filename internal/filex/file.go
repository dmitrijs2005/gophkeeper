// Package filex provides small filesystem utilities used across the project.
package filex

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureSubdDir ensures a subdirectory with the given name exists in the
// current working directory. If it doesn't exist, it is created with 0770
// permissions. It returns the absolute path to the directory.
func EnsureSubdDir(dirName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	dir := filepath.Join(cwd, dirName)

	if err := os.MkdirAll(dir, 0o770); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}

	return dir, nil
}
