package utils

import (
	"math"
	"strings"
)

func mask(s string, mask string, start int, end int) string {
	if len(s) <= 0 {
		return s
	}

	if start < 0 {
		start = 0
	}

	if end > len(s) {
		end = len(s)
	}

	if start >= end {
		return s
	}

	return s[:start] + mask + s[end:]
}

// MaskToken Token 脱敏.
func MaskToken(token string) string {
	quarter := int(math.Ceil(float64(len(token)) / 4))
	return mask(token, strings.Repeat("*", quarter*2), quarter, quarter*3)
}
