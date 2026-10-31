package xdg

import (
	"path/filepath"
	"testing"
)

func TestCacheDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name string
		env  string
		want string
	}{
		{"uses XDG_CACHE_HOME", "/var/cache/me", "/var/cache/me/jcal"},
		{"falls back when unset", "", filepath.Join(home, ".cache", "jcal")},
		{"ignores relative path", "relative/cache", filepath.Join(home, ".cache", "jcal")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CACHE_HOME", tt.env)
			got, err := CacheDir()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("CacheDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	t.Setenv("XDG_CONFIG_HOME", "/etc/me")
	if got, _ := ConfigDir(); got != "/etc/me/jcal" {
		t.Errorf("ConfigDir() = %q", got)
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	if got, _ := ConfigDir(); got != filepath.Join(home, ".config", "jcal") {
		t.Errorf("ConfigDir() = %q", got)
	}
}
