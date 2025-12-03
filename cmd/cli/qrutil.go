package main

import (
	"strings"
)

func qrcodeToUTF8(bitmap [][]bool, inverse bool) string {
	var sb strings.Builder
	full := []rune("██")
	empty := []rune("  ")
	if inverse {
		full, empty = empty, full
	}
	for _, row := range bitmap {
		for _, dot := range row {
			if dot {
				sb.WriteString(string(full))
			} else {
				sb.WriteString(string(empty))
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}
