package otp

import (
	"bytes"
	"testing"
)

func TestGenerateScratchCodesDeterministic(t *testing.T) {
	// Prepare deterministic entropy: sequence of uint32 values.
	var entropy bytes.Buffer
	// Values chosen to produce valid eight-digit codes.
	values := []uint32{
		0x7FFFFFFF, // large number
		1234567890,
		42424242,
	}
	for _, v := range values {
		var buf [4]byte
		buf[0] = byte(v >> 24)
		buf[1] = byte(v >> 16)
		buf[2] = byte(v >> 8)
		buf[3] = byte(v)
		entropy.Write(buf[:])
	}

	codes, err := GenerateScratchCodes(3, &entropy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(codes) != 3 {
		t.Fatalf("expected 3 codes, got %d", len(codes))
	}
	for _, code := range codes {
		if code < 10000000 || code >= scratchModulo {
			t.Fatalf("code not 8 digits: %d", code)
		}
	}
}

func TestGenerateScratchCodesBounds(t *testing.T) {
	entropy := bytes.Repeat([]byte{0xFF}, 80)
	codes, err := GenerateScratchCodes(15, bytes.NewReader(entropy))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(codes) != 10 {
		t.Fatalf("should clamp to 10 codes, got %d", len(codes))
	}
	codes, err = GenerateScratchCodes(-2, bytes.NewReader(entropy))
	if err != nil {
		t.Fatalf("unexpected error for negative: %v", err)
	}
	if len(codes) != 0 {
		t.Fatalf("negative count should yield zero codes, got %d", len(codes))
	}
}
