package main

import (
	cryptoRand "crypto/rand"
	"encoding/base32"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ggpam/pkg/logging"
)

const (
	DefaultSecretFilename = ".google_authenticator"
	DefaultSecretDirPerm  = 0o700
	DefaultSecretFilePerm = 0o600
)

var (
	EnvSecretPath = "GPAM_SECRET_PATH"
)

func defaultSecretPath() string {
	if p := os.Getenv(EnvSecretPath); strings.TrimSpace(p) != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.google_authenticator"
	}
	return filepath.Join(home, DefaultSecretFilename)
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
	return os.MkdirAll(dir, DefaultSecretDirPerm)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func writeInfo(format string, args ...any) {
	logging.Infof(format, args...)
	fmt.Printf(format+"\n", args...)
}
