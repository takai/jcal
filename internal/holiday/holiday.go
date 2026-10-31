// Package holiday reads Japanese national holidays from the Cabinet Office
// CSV and keeps a local copy of it up to date.
package holiday

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

type date struct {
	year  int
	month time.Month
	day   int
}

// Set holds the dates listed in the official dataset. A nil Set contains no
// holidays.
type Set map[date]struct{}

// IsHoliday reports whether the date is listed in the dataset.
func (s Set) IsHoliday(year int, month time.Month, day int) bool {
	_, ok := s[date{year, month, day}]
	return ok
}

// Parse reads the official CSV. Only the date column is used, so the
// Shift_JIS encoded names need no decoding. Lines without a valid date,
// such as the header, are skipped.
func Parse(r io.Reader) (Set, error) {
	set := Set{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		field, _, _ := bytes.Cut(sc.Bytes(), []byte(","))
		if d, ok := parseDate(strings.TrimSpace(string(field))); ok {
			set[d] = struct{}{}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(set) == 0 {
		return nil, errors.New("holiday data contains no dates")
	}
	return set, nil
}

// parseDate parses dates written as "2026/1/1".
func parseDate(s string) (date, bool) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return date{}, false
	}
	var nums [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return date{}, false
		}
		nums[i] = n
	}
	t := time.Date(nums[0], time.Month(nums[1]), nums[2], 0, 0, 0, 0, time.UTC)
	if t.Year() != nums[0] || int(t.Month()) != nums[1] || t.Day() != nums[2] {
		return date{}, false
	}
	return date{nums[0], time.Month(nums[1]), nums[2]}, true
}
