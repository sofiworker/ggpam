package util

import (
	"os"
	"path/filepath"
	"strings"
)

func MkDirWithPerm(path string, mode os.FileMode) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, mode)
}

func FileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func ExpandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		trimmed := strings.TrimPrefix(p, "~")
		trimmed = strings.TrimPrefix(trimmed, string(os.PathSeparator))
		return filepath.Join(home, trimmed), nil
	}
	return p, nil
}
