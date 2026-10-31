package calendar

import (
	"testing"
)

func TestLocaleFromEnv(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Locale
	}{
		{"default", nil, C},
		{"C", map[string]string{"LANG": "C"}, C},
		{"ja_JP", map[string]string{"LANG": "ja_JP"}, Japanese},
		{"ja_JP.UTF-8", map[string]string{"LANG": "ja_JP.UTF-8"}, Japanese},
		{"ja_JP.eucJP", map[string]string{"LANG": "ja_JP.eucJP"}, Japanese},
		{"unsupported falls back to C", map[string]string{"LANG": "fr_FR.UTF-8"}, C},
		{"LC_TIME overrides LANG", map[string]string{"LANG": "C", "LC_TIME": "ja_JP.UTF-8"}, Japanese},
		{"LC_ALL overrides LC_TIME", map[string]string{"LC_ALL": "C", "LC_TIME": "ja_JP.UTF-8"}, C},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LocaleFromEnv(func(k string) string { return tt.env[k] })
			if got != tt.want {
				t.Errorf("got %v, want %v", got.months[0], tt.want.months[0])
			}
		})
	}
}
