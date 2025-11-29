package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const sampleConfig = `JBSWY3DPEHPK3PXP
" TOTP_AUTH
" STEP_SIZE 30
" WINDOW_SIZE 5
" DISALLOW_REUSE 100 200
" RATE_LIMIT 3 30 1000 1010
" TIME_SKEW 1
12345678
87654321
`

func TestParseAndSerialize(t *testing.T) {
	cfg, err := Parse(strings.NewReader(sampleConfig))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if cfg.Secret != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("unexpected secret %s", cfg.Secret)
	}
	if cfg.Options.StepSize != 30 || cfg.Options.WindowSize != 5 {
		t.Fatalf("unexpected options: %+v", cfg.Options)
	}
	if !cfg.Options.TOTPAuth {
		t.Fatal("missing TOTP flag")
	}
	if !cfg.Options.DisallowReuse || len(cfg.Options.DisallowedTimestamps) != 2 {
		t.Fatalf("unexpected disallow list: %+v", cfg.Options.DisallowedTimestamps)
	}
	if cfg.Options.RateLimit == nil || cfg.Options.RateLimit.Attempts != 3 {
		t.Fatalf("missing rate limit: %+v", cfg.Options.RateLimit)
	}
	if len(cfg.ScratchCodes) != 2 {
		t.Fatalf("scratch codes not parsed: %v", cfg.ScratchCodes)
	}
	data, err := cfg.Bytes()
	if err != nil {
		t.Fatalf("Bytes error: %v", err)
	}
	if !strings.Contains(string(data), "\" RATE_LIMIT 3 30") {
		t.Fatalf("serialized data missing rate limit: %s", data)
	}
}

func TestRateLimit(t *testing.T) {
	cfg, err := Parse(strings.NewReader(sampleConfig))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	now := time.Unix(2000, 0)
	if err := cfg.EnforceRateLimit(now); err != nil {
		t.Fatalf("rate limit failed: %v", err)
	}
	cfg.Options.RateLimit.Timestamps = []int64{1990, 1995, 1998}
	if err := cfg.EnforceRateLimit(now); err == nil {
		t.Fatalf("expected rate limit error")
	}
}

func TestGracePeriod(t *testing.T) {
	cfg := &Config{
		Secret: "JBSWY3DPEHPK3PXP",
		Options: Options{
			TOTPAuth:   true,
			Additional: map[string]string{},
			LastLogins: map[int]LoginRecord{},
		},
	}
	now := time.Unix(2_000_000, 0)
	cfg.UpdateLoginRecord("example.com", now.Add(-10*time.Second))
	if !cfg.WithinGracePeriod("example.com", 20*time.Second, now) {
		t.Fatalf("expected within grace period")
	}
	if cfg.WithinGracePeriod("example.com", 5*time.Second, now) {
		t.Fatalf("grace period should have expired")
	}
	if cfg.WithinGracePeriod("", 30*time.Second, now) {
		t.Fatalf("empty host should not match")
	}
}

func TestLoadTooLarge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "google_authenticator")
	oversize := maxFileSize + 1
	if err := os.WriteFile(path, bytes.Repeat([]byte("A"), oversize), 0o600); err != nil {
		t.Fatalf("failed to write oversized file: %v", err)
	}
	if _, err := Load(path); !errors.Is(err, errFileTooLarge) {
		t.Fatalf("expected errFileTooLarge, got %v", err)
	}
}

func TestLoadLargeValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "google_authenticator")
	var sb strings.Builder
	sb.WriteString("JBSWY3DPEHPK3PXP\n")
	for sb.Len() < maxFileSize-16 {
		sb.WriteString("12345678\n")
	}
	content := sb.String()
	if len(content) >= maxFileSize {
		t.Fatalf("test data unexpectedly large: %d", len(content))
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("expected large file to parse, got %v", err)
	}
}
