package cmd

import (
	cryptoRand "crypto/rand"
	"encoding/base32"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func defaultSecretPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.google_authenticator"
	}
	return filepath.Join(home, ".google_authenticator")
}

func expandPath(p string) (string, error) {
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

func randomSecret(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := cryptoRand.Read(buf); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

func ensureParent(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o700)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func writeInfo(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}
