package calendar

import (
	"strings"
	"testing"
	"time"
)

type holidays map[Date]bool

func (h holidays) IsHoliday(year int, month time.Month, day int) bool {
	return h[Date{year, month, day}]
}

func TestMonthC(t *testing.T) {
	got := Month(2026, time.November, Options{Locale: C})
	want := `   November 2026
Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30
`
	assertEqual(t, got, want)
}

func TestMonthJapanese(t *testing.T) {
	got := Month(2026, time.November, Options{Locale: Japanese})
	want := `      11月 2026
日 月 火 水 木 金 土
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30
`
	assertEqual(t, got, want)
}

func TestMonthStartingMonday(t *testing.T) {
	got := Month(2026, time.November, Options{Locale: C, WeekStart: time.Monday})
	want := `   November 2026
Mo Tu We Th Fr Sa Su
                   1
 2  3  4  5  6  7  8
 9 10 11 12 13 14 15
16 17 18 19 20 21 22
23 24 25 26 27 28 29
30
`
	assertEqual(t, got, want)
}

func TestMonthLeapYear(t *testing.T) {
	got := Month(2028, time.February, Options{Locale: C})
	if !strings.HasSuffix(got, "27 28 29\n") {
		t.Errorf("February 2028 should end with 29:\n%s", got)
	}
	got = Month(2100, time.February, Options{Locale: C})
	if !strings.HasSuffix(got, "28\n") || strings.Contains(got, "29") {
		t.Errorf("February 2100 should end with 28:\n%s", got)
	}
}

func TestThreeMonthsAcrossYearBoundary(t *testing.T) {
	got := ThreeMonths(2026, time.December, Options{Locale: C})
	want := `   November 2026         December 2026          January 2027
Su Mo Tu We Th Fr Sa  Su Mo Tu We Th Fr Sa  Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7         1  2  3  4  5                  1  2
 8  9 10 11 12 13 14   6  7  8  9 10 11 12   3  4  5  6  7  8  9
15 16 17 18 19 20 21  13 14 15 16 17 18 19  10 11 12 13 14 15 16
22 23 24 25 26 27 28  20 21 22 23 24 25 26  17 18 19 20 21 22 23
29 30                 27 28 29 30 31        24 25 26 27 28 29 30
                                            31
`
	assertEqual(t, got, want)
}

func TestThreeMonthsJapaneseAlignment(t *testing.T) {
	got := ThreeMonths(2026, time.November, Options{Locale: Japanese})
	lines := strings.Split(got, "\n")
	want := "      10月 2026             11月 2026             12月 2026"
	if lines[0] != want {
		t.Errorf("header:\n got %q\nwant %q", lines[0], want)
	}
	if lines[1] != "日 月 火 水 木 金 土  日 月 火 水 木 金 土  日 月 火 水 木 金 土" {
		t.Errorf("weekdays: %q", lines[1])
	}
}

func TestYear(t *testing.T) {
	got := Year(2026, Options{Locale: C})
	lines := strings.Split(got, "\n")
	wantHead := []string{
		"                              2026",
		"      January               February               March",
		"Su Mo Tu We Th Fr Sa  Su Mo Tu We Th Fr Sa  Su Mo Tu We Th Fr Sa",
		"             1  2  3   1  2  3  4  5  6  7   1  2  3  4  5  6  7",
	}
	for i, w := range wantHead {
		if lines[i] != w {
			t.Errorf("line %d:\n got %q\nwant %q", i, lines[i], w)
		}
	}
	for _, name := range []string{"April", "July", "October", "December"} {
		if !strings.Contains(got, name) {
			t.Errorf("missing %s", name)
		}
	}
	// Four rows of months separated by blank lines.
	if n := strings.Count(got, "\n\n"); n != 3 {
		t.Errorf("blank separators = %d, want 3\n%s", n, got)
	}
}

func TestColors(t *testing.T) {
	opts := Options{
		Locale:    C,
		Color:     true,
		Holidays:  holidays{{2026, time.November, 3}: true, {2026, time.November, 7}: true},
		Highlight: Date{2026, time.November, 23},
	}
	got := Month(2026, time.November, opts)

	for _, want := range []string{
		"\x1b[31m 1\x1b[0m",    // Sunday
		"\x1b[31m 3\x1b[0m",    // holiday
		"\x1b[31m 7\x1b[0m",    // holiday on Saturday: red wins
		"\x1b[34m14\x1b[0m",    // Saturday
		"\x1b[30;47m23\x1b[0m", // highlighted date
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output does not contain %q:\n%q", want, got)
		}
	}
	if strings.Contains(got, "\x1b[31m23") {
		t.Error("highlight must take precedence over holiday colors")
	}
	if strings.Contains(got, "\x1b[34m 7") {
		t.Error("holiday on Saturday must not be blue")
	}
	if !strings.Contains(got, "\x1b[0m  2 \x1b[31m 3") {
		t.Error("ordinary weekdays must not be colored")
	}

	plain := stripANSI(got)
	opts.Color = false
	assertEqual(t, plain, Month(2026, time.November, opts))
}

func TestNoColorEscapesWhenDisabled(t *testing.T) {
	opts := Options{
		Locale:    C,
		Holidays:  holidays{{2026, time.November, 3}: true},
		Highlight: Date{2026, time.November, 3},
	}
	if got := Year(2026, opts); strings.Contains(got, "\x1b") {
		t.Errorf("unexpected escape sequence:\n%q", got)
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := map[string]int{"": 0, "Su": 2, "日": 2, "11月 2026": 9}
	for s, want := range tests {
		if got := displayWidth(s); got != want {
			t.Errorf("displayWidth(%q) = %d, want %d", s, got, want)
		}
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("output mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
