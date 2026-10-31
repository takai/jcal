// Package cli implements the jcal command line.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/takai/jcal/internal/calendar"
	"github.com/takai/jcal/internal/holiday"
)

const usage = `Usage: jcal [options] [[[day] month] year]

Display a calendar highlighting Japanese national holidays.

Options:
  -3, --three    display the previous, current and next month
  -s, --sunday   start weeks on Sunday (default)
  -m, --monday   start weeks on Monday
  -h, --help     display this help and exit
  -V, --version  display version information and exit
`

// holidayTimeout bounds each attempt to download holiday data so that an
// unreachable network never stalls the calendar for long.
const holidayTimeout = 3 * time.Second

// HolidaySource provides holiday data, typically a *holiday.Store.
type HolidaySource interface {
	Load(ctx context.Context) (holiday.Set, error)
	Refresh(ctx context.Context) error
}

// App holds the environment the command runs in.
type App struct {
	Stdout io.Writer
	Stderr io.Writer
	Getenv func(string) string
	Now    func() time.Time
	// Color enables colored output; see ColorEnabled.
	Color   bool
	Version string
	// Holidays opens the holiday data source. It is called only when a
	// calendar is displayed.
	Holidays func() (HolidaySource, error)
}

// ColorEnabled reports whether colored output should be used, given
// whether standard output is a terminal.
func ColorEnabled(getenv func(string) string, terminal bool) bool {
	return terminal && getenv("NO_COLOR") == "" && getenv("TERM") != "dumb"
}

type request struct {
	three     bool
	weekStart time.Weekday
	help      bool
	version   bool
	args      []string
}

// Run executes the command and returns the exit status.
func (a *App) Run(args []string) int {
	req, err := parseArgs(args)
	if err != nil {
		return a.fail(err)
	}
	if req.help {
		return a.write(usage)
	}
	if req.version {
		return a.write("jcal " + a.Version + "\n")
	}

	view, err := resolveView(req, a.Now())
	if err != nil {
		return a.fail(err)
	}

	opts := calendar.Options{
		Locale:    calendar.LocaleFromEnv(a.Getenv),
		WeekStart: req.weekStart,
		Highlight: view.highlight,
		Color:     a.Color,
	}
	// Holiday data is best effort: failures leave holidays unhighlighted
	// and are not reported.
	src, srcErr := a.Holidays()
	if srcErr == nil {
		ctx, cancel := context.WithTimeout(context.Background(), holidayTimeout)
		set, _ := src.Load(ctx)
		cancel()
		opts.Holidays = set
	}

	var out string
	switch {
	case view.month == 0:
		out = calendar.Year(view.year, opts)
	case req.three:
		out = calendar.ThreeMonths(view.year, view.month, opts)
	default:
		out = calendar.Month(view.year, view.month, opts)
	}
	if code := a.write(out); code != 0 {
		return code
	}

	// Check for updates only after the calendar has been displayed.
	if srcErr == nil {
		ctx, cancel := context.WithTimeout(context.Background(), holidayTimeout)
		_ = src.Refresh(ctx)
		cancel()
	}
	return 0
}

// write prints s to standard output and returns the exit status.
func (a *App) write(s string) int {
	if _, err := io.WriteString(a.Stdout, s); err != nil {
		return 1
	}
	return 0
}

func (a *App) fail(err error) int {
	// There is nowhere left to report a failure to write to stderr.
	_, _ = fmt.Fprintf(a.Stderr, "jcal: %v\nTry 'jcal --help' for more information.\n", err)
	return 1
}

func parseArgs(args []string) (request, error) {
	req := request{weekStart: time.Sunday}
	var sunday, monday bool
	setOption := func(name string) error {
		switch name {
		case "3", "-three":
			req.three = true
		case "s", "-sunday":
			sunday = true
		case "m", "-monday":
			monday = true
		case "h", "-help":
			req.help = true
		case "V", "-version":
			req.version = true
		default:
			return fmt.Errorf("unknown option: -%s", name)
		}
		return nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			req.args = append(req.args, args[i+1:]...)
			i = len(args)
		case len(arg) > 2 && arg[:2] == "--":
			if err := setOption(arg[1:]); err != nil {
				return req, err
			}
		case len(arg) > 1 && arg[0] == '-':
			for _, c := range arg[1:] {
				if err := setOption(string(c)); err != nil {
					return req, err
				}
			}
		default:
			req.args = append(req.args, arg)
		}
	}

	if sunday && monday {
		return req, errors.New("options -s and -m cannot be used together")
	}
	if monday {
		req.weekStart = time.Monday
	}
	return req, nil
}

type view struct {
	year  int
	month time.Month // zero for a whole year
	// highlight is today or the explicitly selected day.
	highlight calendar.Date
}

func resolveView(req request, now time.Time) (view, error) {
	today := calendar.Date{Year: now.Year(), Month: now.Month(), Day: now.Day()}
	switch len(req.args) {
	case 0:
		return view{today.Year, today.Month, today}, nil
	case 1:
		if req.three {
			return view{}, errors.New("--three cannot be used when only a year is specified")
		}
		year, err := parseYear(req.args[0])
		if err != nil {
			return view{}, err
		}
		return view{year, 0, today}, nil
	case 2:
		year, err := parseYear(req.args[1])
		if err != nil {
			return view{}, err
		}
		month, err := parseMonth(req.args[0])
		if err != nil {
			return view{}, err
		}
		return view{year, month, today}, nil
	case 3:
		year, err := parseYear(req.args[2])
		if err != nil {
			return view{}, err
		}
		month, err := parseMonth(req.args[1])
		if err != nil {
			return view{}, err
		}
		day, ok := parseNumber(req.args[0])
		if !ok || day < 1 || day > daysIn(year, month) {
			return view{}, fmt.Errorf("invalid date: %s %s, %d", month, req.args[0], year)
		}
		return view{year, month, calendar.Date{Year: year, Month: month, Day: day}}, nil
	default:
		return view{}, errors.New("too many arguments")
	}
}

func parseYear(s string) (int, error) {
	n, ok := parseNumber(s)
	if !ok || n < 1 || n > 9999 {
		return 0, fmt.Errorf("invalid year: %s (must be 1-9999)", s)
	}
	return n, nil
}

func parseMonth(s string) (time.Month, error) {
	n, ok := parseNumber(s)
	if !ok || n < 1 || n > 12 {
		return 0, fmt.Errorf("invalid month: %s (must be 1-12)", s)
	}
	return time.Month(n), nil
}

// parseNumber accepts only decimal digits, rejecting signs and spaces.
func parseNumber(s string) (int, bool) {
	if s == "" || len(s) > 9 {
		return 0, false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

func daysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
