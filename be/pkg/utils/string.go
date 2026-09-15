package utils

import (
	"strings"
	"unicode"
)

func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s) + 4)

	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; i++ {
		r := runes[i]

		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < n && unicode.IsLower(runes[i+1]))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

func ToNormalCase(s string) string {
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s) + 4)

	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; i++ {
		r := runes[i]

		if unicode.IsUpper(r) {
			if i > 0 && runes[i-1] != ' ' && (unicode.IsLower(runes[i-1]) || (i+1 < n && unicode.IsLower(runes[i+1]))) {
				b.WriteByte(' ')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}
