package core

import (
	"os"
	"path/filepath"
)

// root returns the base directory override: ZAGROS_HOME, or the exe directory
// when a file named "portable" sits next to the executable. ok is false when
// neither applies.
func root() (dir string, ok bool) {
	if h := os.Getenv("ZAGROS_HOME"); h != "" {
		return h, true
	}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(d, "portable")); err == nil {
			return d, true
		}
	}
	return "", false
}

// ConfigDir is %APPDATA%\Zagros (or <home>\config in portable/test mode).
func ConfigDir() string {
	if r, ok := root(); ok {
		return filepath.Join(r, "config")
	}
	return filepath.Join(envOr("APPDATA", userSub("AppData", "Roaming")), "Zagros")
}

// DataDir is %LOCALAPPDATA%\Zagros (or <home>\data in portable/test mode).
func DataDir() string {
	if r, ok := root(); ok {
		return filepath.Join(r, "data")
	}
	return filepath.Join(envOr("LOCALAPPDATA", userSub("AppData", "Local")), "Zagros")
}

// ExtDir holds installed extensions.
func ExtDir() string { return filepath.Join(ConfigDir(), "extensions") }

// ThemeDir holds user themes.
func ThemeDir() string { return filepath.Join(ConfigDir(), "themes") }

// ConfigPath is the full path of zagros.toml.
func ConfigPath() string { return filepath.Join(ConfigDir(), "zagros.toml") }

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

func userSub(parts ...string) string {
	h, err := os.UserHomeDir()
	if err != nil {
		h = "."
	}
	return filepath.Join(append([]string{h}, parts...)...)
}
