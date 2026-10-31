# jcal — Functional Specification

## 1. Overview

`jcal` is a command-line calendar program that displays Japanese national holidays.

It behaves like the standard Unix `cal` command, with additional color highlighting for Sundays, Saturdays, Japanese national holidays, and the current date.

The program is intended for everyday terminal use. It requires no initial configuration and automatically obtains holiday information from the Cabinet Office of Japan.

## 2. Command-Line Interface

The command syntax is:

```text
jcal [options] [[[day] month] year]
```

### 2.1 Arguments

| Command | Behavior |
|---|---|
| `jcal` | Display the current month |
| `jcal 2026` | Display the entire year 2026 |
| `jcal 11 2026` | Display November 2026 |
| `jcal 3 11 2026` | Display November 2026, highlighting November 3 |
| `jcal --three` | Display the previous, current, and next months |
| `jcal --three 11 2026` | Display October, November, and December 2026 |

The year must be specified in full. Month and day arguments are numeric.

When no arguments are provided, the current date is determined using the system's local time.

### 2.2 Options

| Option | Description |
|---|---|
| `-3`, `--three` | Display three consecutive months centered on the selected month |
| `-s`, `--sunday` | Start weeks on Sunday |
| `-m`, `--monday` | Start weeks on Monday |
| `-h`, `--help` | Display usage information |
| `-V`, `--version` | Display version information |

Sunday is the default first day of the week.

Options may be combined with date arguments where applicable.

The three-month display must work correctly across year boundaries.

When only a year is specified, `--three` is not applicable.

Other options provided by `cal` are outside the scope of this specification.

## 3. Calendar Display

### 3.1 General Behavior

Calendar output should resemble the standard `cal` command as closely as practical.

- Display dates in a conventional calendar grid.
- Align day numbers and weekday headers consistently.
- Use the Gregorian calendar exclusively.
- Display leap years correctly.
- Do not display holiday names in the calendar grid.
- Do not display additional legends, messages, or explanatory text during normal operation.
- Keep the calendar compact and suitable for terminal use.

### 3.2 Locale Support

The program must support the `C` and `ja_JP` locales, including their common encoding suffixes such as `ja_JP.UTF-8`.

The locale is selected through the standard locale environment variables, including `LANG` and `LC_TIME`.

The output must follow the selected locale for month headings and weekday labels.

**C locale:**

```text
   November 2026
Su Mo Tu We Th Fr Sa
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30
```

**Japanese locale:**

```text
      11月 2026
日 月 火 水 木 金 土
 1  2  3  4  5  6  7
 8  9 10 11 12 13 14
15 16 17 18 19 20 21
22 23 24 25 26 27 28
29 30
```

Localized output must preserve correct visual alignment, including Japanese characters.

The locale affects presentation, not the dates recognized as Japanese holidays.

For unsupported locales, the program may use the `C` locale as a fallback.

### 3.3 Three-Month Display

With `--three`, display the previous month, selected month, and following month side by side, following the layout of `cal --three`.

All three months use the same locale, week-start convention, and coloring rules.

### 3.4 Year Display

When only a year is specified, display all twelve months in a compact, conventional year-calendar layout.

The same holiday highlighting and localization rules apply.

## 4. Colors and Highlighting

The default terminal colors are:

| Date category | Foreground | Background |
|---|---|---|
| Today or explicitly selected date | Black | White |
| Japanese national holiday or statutory holiday | Red | Default |
| Sunday | Red | Default |
| Saturday | Blue | Default |
| Ordinary weekday | Default | Default |

### 4.1 Current Date

Today's date must be highlighted with a white background and black text, regardless of whether it falls on a weekday, weekend, or national holiday.

Today's date is highlighted whenever it appears in the displayed calendar, including three-month and year views.

When an explicit day is supplied, that day is highlighted using the same white-background, black-text style instead of today's date.

Only one date is highlighted as the selected date.

When displaying a month or year that does not contain today's date, no date is highlighted unless an explicit day is supplied.

### 4.2 Weekends and Holidays

Sundays and Japanese national holidays are displayed in red.

Saturdays are displayed in blue.

If a holiday falls on Saturday, red takes precedence over blue.

The current-date or explicitly selected-date highlighting takes precedence over weekend and holiday colors.

Colors must not affect spacing or alignment.

### 4.3 Color Availability

- Enable color by default when output is connected to a compatible terminal.
- Do not emit color escape sequences when output is redirected to a file or pipe.
- Respect the `NO_COLOR` environment variable.
- Preserve readable output on terminals without color support.

## 5. Japanese Holiday Data

### 5.1 Official Data Source

Japanese national holidays must be obtained from the official CSV published by the Cabinet Office of Japan:

https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv

This is the authoritative source for holiday dates.

The program must recognize all holidays and statutory non-working days included in the published data, including:

- National holidays (国民の祝日)
- Substitute holidays (振替休日)
- Holidays between national holidays (国民の休日)
- Special holidays included in the official dataset

The program must not independently calculate Japanese national holidays or infer unlisted future holidays.

### 5.2 Automatic Download

The program automatically downloads the official holiday dataset.

No user action or setup command should be required.

Expected behavior:

1. When no local holiday data exists, attempt to download it automatically.
2. Reuse locally stored data during subsequent executions.
3. Periodically check for updates so that changes published by the Cabinet Office become available without manual intervention.
4. Avoid unnecessary downloads on every invocation.
5. Store successfully downloaded data for offline use.

Holiday updates must not interrupt normal calendar display.

### 5.3 Network Failures

The program must remain usable without network access.

When an update fails:

- Continue using previously downloaded holiday data, if available.
- If no holiday data is available, display the calendar without holiday-specific highlighting.
- Continue highlighting Saturdays and Sundays normally.
- Do not display routine network errors in the calendar output.
- Allow a later invocation to retry the download.

A failure to retrieve holiday data must not prevent calendar display.

### 5.4 Data Coverage

Only dates included in the official dataset are treated as holidays.

Dates outside the available dataset are displayed as ordinary calendar dates, except for Saturday and Sunday highlighting.

The absence of a date from the dataset must not be interpreted as a definitive statement that it is not a holiday.

The program must tolerate updates to the official dataset without requiring a new release.

## 6. Configuration and Local Storage

The program must follow the XDG Base Directory Specification.

Use the following locations:

| Purpose | Location |
|---|---|
| User configuration | `$XDG_CONFIG_HOME/jcal/` |
| Downloaded holiday data | `$XDG_CACHE_HOME/jcal/` |

When the corresponding environment variables are not defined, use the standard XDG defaults:

- `~/.config/jcal/`
- `~/.cache/jcal/`

Downloaded data and update-related metadata belong in the cache directory.

The program must work without a configuration file. Running `jcal` must not require users to create or edit any settings.

The absence of a configuration directory is not an error.

The program must not write configuration or cached data into the current working directory.

## 7. Input Validation and Errors

Invalid input must produce a clear error message on standard error and a nonzero exit status.

Invalid input includes:

- An invalid month number.
- A nonexistent date.
- An invalid or unsupported year.
- Too many positional arguments.
- An unknown command-line option.
- Conflicting options such as `-s` and `-m`.

Successful calendar output is written to standard output.

Help and version requests must not require network access or holiday data.

## 8. Compatibility and Scope

The program aims to preserve the familiar appearance and basic behavior of Unix `cal`, but is not intended to be a complete replacement.

The following are explicitly outside scope:

- Julian calendar support.
- Historical Gregorian calendar reform behavior.
- Easter calculations.
- Support for holidays from countries other than Japan.
- User-defined holidays or events.
- Holiday names displayed inside calendar cells.
- Interactive terminal interfaces.
- Advanced date expressions or natural-language date parsing.
- Full compatibility with every `cal` option.

All calendar dates use the Gregorian calendar, including dates before its historical adoption.

## 9. Acceptance Criteria

The program is considered functionally complete when:

1. `jcal` displays the current month in a layout comparable to `cal`.
2. `jcal 2026` displays all twelve months of 2026.
3. `jcal 11 2026` displays November 2026.
4. `jcal 3 11 2026` highlights November 3 with black text on a white background.
5. `jcal --three` displays three consecutive months.
6. `-s` and `-m` correctly change the first weekday.
7. `LANG=C` produces English calendar headings.
8. `LANG=ja_JP.UTF-8` produces Japanese calendar headings.
9. Sundays and official holidays appear in red, and Saturdays appear in blue.
10. Today's date appears with black text on a white background, overriding other date colors.
11. Holiday coloring correctly recognizes substitute holidays and holidays between national holidays.
12. Holiday data is downloaded automatically and reused on subsequent executions.
13. Calendar display remains available when the network is unavailable.
14. Redirected output contains no terminal color escape sequences.
15. Configuration and cached files follow the XDG Base Directory Specification.
16. Invalid arguments produce clear errors without displaying a misleading calendar.

## 10. References

- [cal(1) — Linux manual page](https://man7.org/linux/man-pages/man1/cal.1.html)
- [Cabinet Office of Japan — National Holidays](https://www8.cao.go.jp/chosei/shukujitsu/gaiyou.html)
- [Cabinet Office — Official Holiday CSV](https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv)
- [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/)