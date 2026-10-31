// Command jcal displays a calendar highlighting Japanese national holidays.
package main

import (
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/takai/jcal/internal/cli"
	"github.com/takai/jcal/internal/holiday"
	"github.com/takai/jcal/internal/xdg"
)

// version is set at build time with -ldflags "-X main.version=...".
var version string

func main() {
	app := &cli.App{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Getenv:  os.Getenv,
		Now:     time.Now,
		Color:   cli.ColorEnabled(os.Getenv, isTerminal(os.Stdout)),
		Version: buildVersion(),
		Holidays: func() (cli.HolidaySource, error) {
			dir, err := xdg.CacheDir()
			if err != nil {
				return nil, err
			}
			return &holiday.Store{
				Dir:    dir,
				URL:    holiday.OfficialURL,
				Client: http.DefaultClient,
				Now:    time.Now,
			}, nil
		},
	}
	os.Exit(app.Run(os.Args[1:]))
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
