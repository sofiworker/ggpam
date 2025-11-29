package authenticator

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"gpam/pkg/config"
	"gpam/pkg/otp"
)

func TestVerifyTOTP(t *testing.T) {
	cfg := &config.Config{
		Secret: "JBSWY3DPEHPK3PXP",
		Options: config.Options{
			TOTPAuth:   true,
			StepSize:   30,
			WindowSize: 3,
			Additional: map[string]string{},
		},
	}
	now := time.Unix(1_600_000_000, 0)
	auth := &Authenticator{
		Now: func() time.Time { return now },
	}
	secret, err := cfg.SecretBytes()
	if err != nil {
		t.Fatalf("secret decode failed: %v", err)
	}
	counter := uint64(now.Unix() / int64(cfg.Step()))
	code := otp.Compute(secret, counter)
	token := fmt.Sprintf("%06d", code)
	res, err := auth.VerifyCode(cfg, token, VerifyOptions{})
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if res.Type != ResultTOTP {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestScratchCode(t *testing.T) {
	cfg := &config.Config{
		Secret: "JBSWY3DPEHPK3PXP",
		ScratchCodes: []int{
			12345678,
		},
		Options: config.Options{
			TOTPAuth:   true,
			Additional: map[string]string{},
		},
	}
	auth := &Authenticator{}
	res, err := auth.VerifyCode(cfg, "12345678", VerifyOptions{})
	if err != nil {
		t.Fatalf("scratch verify failed: %v", err)
	}
	if res.Type != ResultScratch {
		t.Fatalf("unexpected result: %+v", res)
	}
	if len(cfg.ScratchCodes) != 0 {
		t.Fatal("scratch code not removed")
	}
}

func TestTOTPTimeSkewRecalibration(t *testing.T) {
	cfg := &config.Config{
		Secret: "JBSWY3DPEHPK3PXP",
		Options: config.Options{
			TOTPAuth:   true,
			StepSize:   30,
			WindowSize: 3,
			Additional: map[string]string{},
		},
	}
	secretBytes, err := cfg.SecretBytes()
	if err != nil {
		t.Fatalf("secret decode failed: %v", err)
	}
	skewSteps := int64(4)
	current := int64(1_700_000_000)
	auth := &Authenticator{
		Now: func() time.Time { return time.Unix(current, 0) },
	}
	makeCode := func() string {
		counter := current/int64(cfg.Step()) + skewSteps
		return fmt.Sprintf("%06d", otp.Compute(secretBytes, uint64(counter)))
	}

	for i := 0; i < 2; i++ {
		if _, err := auth.VerifyCode(cfg, makeCode(), VerifyOptions{}); !errors.Is(err, ErrInvalidCode) {
			t.Fatalf("expected invalid code on attempt %d, got %v", i+1, err)
		}
		current += int64(cfg.Step())
	}

	res, err := auth.VerifyCode(cfg, makeCode(), VerifyOptions{})
	if err != nil {
		t.Fatalf("expected success after skew reset: %v", err)
	}
	if res.Type != ResultTOTP {
		t.Fatalf("unexpected result type: %+v", res)
	}
	if cfg.Options.TimeSkew != int(skewSteps) {
		t.Fatalf("time skew not updated, got %d", cfg.Options.TimeSkew)
	}
	if len(cfg.Options.ResettingTimeSkew) != 0 {
		t.Fatalf("resetting skew samples not cleared: %+v", cfg.Options.ResettingTimeSkew)
	}
	if !cfg.Dirty {
		t.Fatal("config should be marked dirty after skew update")
	}
}
