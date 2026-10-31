// Package xdg resolves jcal's directories according to the XDG Base
// Directory Specification.
package xdg

import (
	"os"
	"path/filepath"
)

const appName = "jcal"

// CacheDir returns the directory for downloaded holiday data and its
// metadata: $XDG_CACHE_HOME/jcal, or ~/.cache/jcal.
func CacheDir() (string, error) {
	return appDir("XDG_CACHE_HOME", ".cache")
}

// ConfigDir returns the directory for user configuration:
// $XDG_CONFIG_HOME/jcal, or ~/.config/jcal. The directory is optional.
func ConfigDir() (string, error) {
	return appDir("XDG_CONFIG_HOME", ".config")
}

func appDir(envName, fallback string) (string, error) {
	// The specification requires absolute paths; relative ones are ignored.
	if dir := os.Getenv(envName); filepath.IsAbs(dir) {
		return filepath.Join(dir, appName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, fallback, appName), nil
}
