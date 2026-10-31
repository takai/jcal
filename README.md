# jcal

A simple command-line calendar with Japanese national holidays.

`jcal` works like the Unix `cal` command, highlighting Japanese national holidays and weekends. Holiday data is automatically downloaded from the Cabinet Office of Japan.

## Features

- Familiar `cal`-style calendar display
- Japanese national holidays and Sundays in red, Saturdays in blue
- Today's date highlighted in black on white
- English (`LANG=C`) and Japanese (`LANG=ja_JP.UTF-8`) output
- Automatic holiday data updates with offline support
- XDG Base Directory compliant

## Usage

```sh
jcal [options] [[[day] month] year]
```

Examples:

```sh
jcal                 # Current month
jcal 2026            # Entire year
jcal 11 2026         # November 2026
jcal 3 11 2026       # Highlight November 3, 2026
jcal --three         # Previous, current, and next months
jcal -m              # Start weeks on Monday
LANG=ja_JP.UTF-8 jcal # Japanese output
```

## Options

| Option | Description |
|---|---|
| `-3`, `--three` | Show three months |
| `-s`, `--sunday` | Start weeks on Sunday (default) |
| `-m`, `--monday` | Start weeks on Monday |
| `-h`, `--help` | Show help |
| `-V`, `--version` | Show version |

## Holiday Data

Holiday information is provided by the [Cabinet Office of Japan](https://www8.cao.go.jp/chosei/shukujitsu/gaiyou.html).

Data is downloaded and cached automatically in `$XDG_CACHE_HOME/jcal/` (default: `~/.cache/jcal/`).

The calendar continues to work offline using cached holiday data.

## Releasing

Releases are built with [GoReleaser](https://goreleaser.com/). Push a version tag and GitHub Actions publishes the archives:

```sh
git tag v0.1.0
git push origin v0.1.0
```

To try a local build without publishing, run `goreleaser release --snapshot --clean`.
