package util

import (
	cryptoRand "crypto/rand"
	"encoding/base32"
)

func RandomSecret(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := cryptoRand.Read(buf); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}
