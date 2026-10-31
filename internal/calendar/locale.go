package calendar

import (
	"strings"
	"time"
)

// Locale holds the localized labels used in calendar headings.
type Locale struct {
	months   [12]string
	weekdays [7]string // indexed by time.Weekday
}

// C is the default English locale.
var C = Locale{
	months: [12]string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	},
	weekdays: [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"},
}

// Japanese is the ja_JP locale.
var Japanese = Locale{
	months: [12]string{
		"1月", "2月", "3月", "4月", "5月", "6月",
		"7月", "8月", "9月", "10月", "11月", "12月",
	},
	weekdays: [7]string{"日", "月", "火", "水", "木", "金", "土"},
}

// LocaleFromEnv selects the locale from LC_ALL, LC_TIME and LANG, in that
// order of precedence. Unsupported locales fall back to C.
func LocaleFromEnv(getenv func(string) string) Locale {
	for _, name := range []string{"LC_ALL", "LC_TIME", "LANG"} {
		v := getenv(name)
		if v == "" {
			continue
		}
		if v == "ja" || strings.HasPrefix(v, "ja_JP") {
			return Japanese
		}
		return C
	}
	return C
}

func (l Locale) monthName(m time.Month) string {
	return l.months[m-1]
}
