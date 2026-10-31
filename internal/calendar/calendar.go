// Package calendar renders cal(1)-style calendars with Japanese holiday
// highlighting.
package calendar

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	monthWidth = 20
	gutter     = "  "

	sgrReset     = "\x1b[0m"
	sgrHighlight = "\x1b[30;47m" // black on white
	sgrRed       = "\x1b[31m"
	sgrBlue      = "\x1b[34m"
)

// Date is a calendar date without time or location.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// HolidayChecker reports whether a date is a holiday.
type HolidayChecker interface {
	IsHoliday(year int, month time.Month, day int) bool
}

// Options controls how calendars are rendered.
type Options struct {
	Locale Locale
	// WeekStart is the first day of the week: time.Sunday or time.Monday.
	WeekStart time.Weekday
	// Holidays may be nil, in which case no holidays are highlighted.
	Holidays HolidayChecker
	// Highlight is the date shown in black on white; the zero Date
	// highlights nothing.
	Highlight Date
	// Color enables ANSI color escape sequences.
	Color bool
}

// Month renders a single month.
func Month(year int, month time.Month, o Options) string {
	return join([][]string{o.monthBlock(year, month, true)})
}

// ThreeMonths renders the months before and after the given month beside it.
func ThreeMonths(year int, month time.Month, o Options) string {
	blocks := make([][]string, 0, 3)
	for i := -1; i <= 1; i++ {
		t := time.Date(year, month+time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		blocks = append(blocks, o.monthBlock(t.Year(), t.Month(), true))
	}
	return join(blocks)
}

// Year renders all twelve months, three per row.
func Year(year int, o Options) string {
	var b strings.Builder
	header := center(strconv.Itoa(year), 3*monthWidth+2*len(gutter))
	b.WriteString(strings.TrimRight(header, " ") + "\n")
	for row := range 4 {
		if row > 0 {
			b.WriteString("\n")
		}
		blocks := make([][]string, 0, 3)
		for col := range 3 {
			blocks = append(blocks, o.monthBlock(year, time.Month(row*3+col+1), false))
		}
		b.WriteString(join(blocks))
	}
	return b.String()
}

// monthBlock returns the lines of one month, each exactly monthWidth
// columns wide.
func (o Options) monthBlock(year int, month time.Month, withYear bool) []string {
	title := o.Locale.monthName(month)
	if withYear {
		title += " " + strconv.Itoa(year)
	}

	labels := make([]string, 7)
	for i := range labels {
		labels[i] = o.Locale.weekdays[(int(o.WeekStart)+i)%7]
	}
	lines := []string{center(title, monthWidth), strings.Join(labels, " ")}

	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	days := first.AddDate(0, 1, -1).Day()
	offset := (int(first.Weekday()) - int(o.WeekStart) + 7) % 7

	cells := make([]string, offset, offset+days+6)
	for i := range cells {
		cells[i] = "  "
	}
	for day := 1; day <= days; day++ {
		weekday := time.Weekday((int(first.Weekday()) + day - 1) % 7)
		cells = append(cells, o.dayCell(year, month, day, weekday))
	}
	for len(cells)%7 != 0 {
		cells = append(cells, "  ")
	}
	for i := 0; i < len(cells); i += 7 {
		lines = append(lines, strings.Join(cells[i:i+7], " "))
	}
	return lines
}

func (o Options) dayCell(year int, month time.Month, day int, weekday time.Weekday) string {
	s := fmt.Sprintf("%2d", day)
	if !o.Color {
		return s
	}
	var sgr string
	switch {
	case o.Highlight == Date{year, month, day}:
		sgr = sgrHighlight
	case weekday == time.Sunday || (o.Holidays != nil && o.Holidays.IsHoliday(year, month, day)):
		sgr = sgrRed
	case weekday == time.Saturday:
		sgr = sgrBlue
	default:
		return s
	}
	return sgr + s + sgrReset
}

// join places month blocks side by side, padding shorter months with blank
// lines, and trims trailing spaces from every line.
func join(blocks [][]string) string {
	height := 0
	for _, blk := range blocks {
		height = max(height, len(blk))
	}
	blank := strings.Repeat(" ", monthWidth)
	var b strings.Builder
	for i := range height {
		parts := make([]string, len(blocks))
		for j, blk := range blocks {
			parts[j] = blank
			if i < len(blk) {
				parts[j] = blk[i]
			}
		}
		b.WriteString(strings.TrimRight(strings.Join(parts, gutter), " ") + "\n")
	}
	return b.String()
}
