package otp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
)

const Modulo = 1000000

func Compute(secret []byte, counter uint64) int {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, secret)
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0F
	var value uint32
	for i := 0; i < 4; i++ {
		value <<= 8
		value |= uint32(sum[int(offset)+i])
	}
	value &= 0x7FFFFFFF
	return int(value % Modulo)
}
