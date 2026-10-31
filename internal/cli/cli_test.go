package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/takai/jcal/internal/holiday"
)

type fakeSource struct {
	set       holiday.Set
	err       error
	loaded    bool
	refreshed bool
}

func (f *fakeSource) Load(context.Context) (holiday.Set, error) {
	f.loaded = true
	return f.set, f.err
}

func (f *fakeSource) Refresh(context.Context) error {
	f.refreshed = true
	return f.err
}

type result struct {
	code   int
	stdout string
	stderr string
	source *fakeSource
}

func run(t *testing.T, color bool, env map[string]string, args ...string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	src := &fakeSource{}
	app := &App{
		Stdout:  &stdout,
		Stderr:  &stderr,
		Getenv:  func(k string) string { return env[k] },
		Now:     func() time.Time { return time.Date(2026, 11, 1, 12, 0, 0, 0, time.Local) },
		Color:   color,
		Version: "1.2.3",
		Holidays: func() (HolidaySource, error) {
			return src, nil
		},
	}
	code := app.Run(args)
	return result{code, stdout.String(), stderr.String(), src}
}

func TestRunCurrentMonth(t *testing.T) {
	r := run(t, false, nil)
	if r.code != 0 {
		t.Fatalf("exit %d: %s", r.code, r.stderr)
	}
	if !strings.HasPrefix(r.stdout, "   November 2026\nSu Mo") {
		t.Errorf("unexpected output:\n%s", r.stdout)
	}
	if !r.source.loaded || !r.source.refreshed {
		t.Error("holiday data should be loaded and refreshed")
	}
}

func TestRunArguments(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"2026"}, "                              2026\n      January"},
		{[]string{"11", "2026"}, "   November 2026\n"},
		{[]string{"3", "11", "2026"}, "   November 2026\n"},
		{[]string{"--three"}, "    October 2026         November 2026         December 2026\n"},
		{[]string{"-3", "1", "2027"}, "   December 2026          January 2027         February 2027\n"},
		{[]string{"-m", "11", "2026"}, "   November 2026\nMo Tu"},
		{[]string{"-3m"}, "    October 2026         November 2026         December 2026\nMo Tu"},
		{[]string{"--", "11", "2026"}, "   November 2026\n"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			r := run(t, false, nil, tt.args...)
			if r.code != 0 {
				t.Fatalf("exit %d: %s", r.code, r.stderr)
			}
			if !strings.HasPrefix(r.stdout, tt.want) {
				t.Errorf("output does not start with %q:\n%s", tt.want, r.stdout)
			}
		})
	}
}

func TestRunHighlightsToday(t *testing.T) {
	r := run(t, true, nil)
	if !strings.Contains(r.stdout, "\x1b[30;47m 1\x1b[0m") {
		t.Errorf("today not highlighted:\n%q", r.stdout)
	}
	r = run(t, true, nil, "2026")
	if strings.Count(r.stdout, "\x1b[30;47m") != 1 {
		t.Errorf("today should be highlighted once in the year view:\n%q", r.stdout)
	}
	r = run(t, true, nil, "11", "2025")
	if strings.Contains(r.stdout, "\x1b[30;47m") {
		t.Errorf("no date should be highlighted:\n%q", r.stdout)
	}
}

func TestRunHighlightsExplicitDay(t *testing.T) {
	r := run(t, true, nil, "3", "11", "2026")
	if !strings.Contains(r.stdout, "\x1b[30;47m 3\x1b[0m") {
		t.Errorf("selected day not highlighted:\n%q", r.stdout)
	}
	if strings.Count(r.stdout, "\x1b[30;47m") != 1 {
		t.Errorf("only the selected day should be highlighted:\n%q", r.stdout)
	}
}

func TestRunWithoutColor(t *testing.T) {
	r := run(t, false, nil, "3", "11", "2026")
	if strings.Contains(r.stdout, "\x1b") {
		t.Errorf("unexpected escape sequence:\n%q", r.stdout)
	}
}

func TestRunJapaneseLocale(t *testing.T) {
	r := run(t, false, map[string]string{"LANG": "ja_JP.UTF-8"}, "11", "2026")
	if !strings.HasPrefix(r.stdout, "      11月 2026\n日 月 火 水 木 金 土\n") {
		t.Errorf("unexpected output:\n%s", r.stdout)
	}
}

func TestRunUsesHolidays(t *testing.T) {
	var stdout bytes.Buffer
	set, err := holiday.Parse(strings.NewReader("2026/11/3,x\n"))
	if err != nil {
		t.Fatal(err)
	}
	app := &App{
		Stdout:   &stdout,
		Stderr:   &bytes.Buffer{},
		Getenv:   func(string) string { return "" },
		Now:      time.Now,
		Color:    true,
		Holidays: func() (HolidaySource, error) { return &fakeSource{set: set}, nil },
	}
	if code := app.Run([]string{"11", "2026"}); code != 0 {
		t.Fatal(code)
	}
	if !strings.Contains(stdout.String(), "\x1b[31m 3\x1b[0m") {
		t.Errorf("holiday not highlighted:\n%q", stdout.String())
	}
}

func TestRunHolidayFailureIsSilent(t *testing.T) {
	var stdout, stderr bytes.Buffer
	src := &fakeSource{err: errors.New("network unreachable")}
	app := &App{
		Stdout:   &stdout,
		Stderr:   &stderr,
		Getenv:   func(string) string { return "" },
		Now:      func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local) },
		Color:    true,
		Holidays: func() (HolidaySource, error) { return src, nil },
	}
	if code := app.Run([]string{"11", "2026"}); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if stderr.Len() != 0 {
		t.Errorf("unexpected stderr: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "\x1b[31m 1\x1b[0m") || !strings.Contains(stdout.String(), "\x1b[34m 7\x1b[0m") {
		t.Errorf("weekends should still be colored:\n%q", stdout.String())
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}} {
		r := run(t, false, nil, args...)
		if r.code != 0 || !strings.Contains(r.stdout, "Usage: jcal") {
			t.Errorf("%v: exit %d, output %q", args, r.code, r.stdout)
		}
		if r.source.loaded {
			t.Errorf("%v: holiday data should not be loaded", args)
		}
	}
	for _, args := range [][]string{{"-V"}, {"--version"}} {
		r := run(t, false, nil, args...)
		if r.code != 0 || r.stdout != "jcal 1.2.3\n" {
			t.Errorf("%v: exit %d, output %q", args, r.code, r.stdout)
		}
		if r.source.loaded {
			t.Errorf("%v: holiday data should not be loaded", args)
		}
	}
}

func TestRunErrors(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"13", "2026"}, "invalid month: 13"},
		{[]string{"0", "2026"}, "invalid month: 0"},
		{[]string{"30", "2", "2026"}, "invalid date: February 30, 2026"},
		{[]string{"29", "2", "2100"}, "invalid date: February 29, 2100"},
		{[]string{"0"}, "invalid year: 0"},
		{[]string{"10000"}, "invalid year: 10000"},
		{[]string{"abc"}, "invalid year: abc"},
		{[]string{"+5"}, "invalid year: +5"},
		{[]string{"1", "1", "1", "2026"}, "too many arguments"},
		{[]string{"-x"}, "unknown option: -x"},
		{[]string{"--julian"}, "unknown option: --julian"},
		{[]string{"-s", "-m"}, "options -s and -m cannot be used together"},
		{[]string{"--monday", "--sunday"}, "options -s and -m cannot be used together"},
		{[]string{"-3", "2026"}, "--three cannot be used when only a year is specified"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			r := run(t, false, nil, tt.args...)
			if r.code == 0 {
				t.Fatal("expected nonzero exit status")
			}
			if r.stdout != "" {
				t.Errorf("no calendar should be printed:\n%s", r.stdout)
			}
			if !strings.Contains(r.stderr, "jcal: "+tt.want) {
				t.Errorf("stderr = %q, want %q", r.stderr, tt.want)
			}
			if r.source.loaded {
				t.Error("holiday data should not be loaded")
			}
		})
	}
}

func TestColorEnabled(t *testing.T) {
	tests := []struct {
		env      map[string]string
		terminal bool
		want     bool
	}{
		{nil, true, true},
		{nil, false, false},
		{map[string]string{"NO_COLOR": "1"}, true, false},
		{map[string]string{"TERM": "dumb"}, true, false},
	}
	for _, tt := range tests {
		got := ColorEnabled(func(k string) string { return tt.env[k] }, tt.terminal)
		if got != tt.want {
			t.Errorf("ColorEnabled(%v, %v) = %v", tt.env, tt.terminal, got)
		}
	}
}
