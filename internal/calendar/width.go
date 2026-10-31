package calendar

import (
	"strings"
	"unicode/utf8"
)

// displayWidth returns the number of terminal columns s occupies, counting
// East Asian wide characters as two columns.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isWide(r rune) bool {
	return r >= 0x1100 &&
		(r <= 0x115f || // Hangul Jamo
			r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) || // CJK, Kana
			(r >= 0xac00 && r <= 0xd7a3) || // Hangul Syllables
			(r >= 0xf900 && r <= 0xfaff) || // CJK Compatibility Ideographs
			(r >= 0xfe30 && r <= 0xfe6f) || // CJK Compatibility Forms
			(r >= 0xff00 && r <= 0xff60) || // Fullwidth Forms
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x3fffd))
}

// center pads s to width columns. The left padding is based on the
// character count, matching the headings of cal(1) in both C and ja_JP
// locales; the right padding fills the remaining columns exactly.
func center(s string, width int) string {
	left := max((width-utf8.RuneCountInString(s))/2, 0)
	right := max(width-left-displayWidth(s), 0)
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
