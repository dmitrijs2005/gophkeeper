package filex

import (
	"fmt"
	"os"
	"path/filepath"
)

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
