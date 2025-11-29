package otp

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"io"
)

const scratchModulo = 100000000

// GenerateScratchCodes returns n scratch codes using the provided entropy source.
// It bounds n to [0,10] and ensures codes are eight digits without leading zeros.
func GenerateScratchCodes(n int, randSrc io.Reader) ([]int, error) {
	if randSrc == nil {
		randSrc = cryptoRand.Reader
	}
	if n < 0 {
		n = 0
	}
	if n > 10 {
		n = 10
	}
	codes := make([]int, 0, n)
	for len(codes) < n {
		var buf [4]byte
		if _, err := io.ReadFull(randSrc, buf[:]); err != nil {
			return nil, err
		}
		val := int(int32(binary.BigEndian.Uint32(buf[:]) & 0x7FFFFFFF))
		code := val % scratchModulo
		if code < scratchModulo/10 {
			continue
		}
		codes = append(codes, code)
	}
	return codes, nil
}

// GenerateScratchCodesDefault generates scratch codes using crypto/rand.
func GenerateScratchCodesDefault(n int) ([]int, error) {
	return GenerateScratchCodes(n, cryptoRand.Reader)
}
