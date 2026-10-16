package theme

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// DefaultName is the theme used for light mode when nothing is configured.
const (
	DefaultLight = "boxy-light"
	DefaultDark  = "boxy-dark"
)

// Builtin returns the built-in theme with the given name (case-insensitive).
func Builtin(name string) (Theme, bool) {
	for _, mk := range builtins {
		if t := mk(); strings.EqualFold(t.Name, name) {
			return t, true
		}
	}
	return Theme{}, false
}

// List returns the names of the built-in themes.
func List() []string {
	names := make([]string, len(builtins))
	for i, mk := range builtins {
		names[i] = mk().Name
	}
	return names
}

// Default returns boxy-dark or boxy-light.
func Default(dark bool) Theme {
	if dark {
		return boxyDark()
	}
	return boxyLight()
}

// Parse decodes theme TOML. Fields that are absent take their value from the
// boxy theme matching the file's "dark" flag, so partial themes are valid.
// fallbackName is used when the file has no name.
func Parse(data []byte, fallbackName string) (Theme, error) {
	var probe struct {
		Dark bool `toml:"dark"`
	}
	if err := toml.Unmarshal(data, &probe); err != nil {
		return Theme{}, err
	}
	t := Default(probe.Dark)
	t.Name, t.Author = "", ""
	if err := toml.Unmarshal(data, &t); err != nil {
		return Theme{}, err
	}
	if t.Name == "" {
		t.Name = fallbackName
	}
	t.fixup()
	return t, nil
}

// fixup repairs values that would break layout.
func (t *Theme) fixup() {
	d := Default(t.Dark)
	if t.Font == "" {
		t.Font = d.Font
	}
	if t.MonoFont == "" {
		t.MonoFont = d.MonoFont
	}
	if t.FontSize <= 0 {
		t.FontSize = d.FontSize
	}
	if t.Width <= 0 {
		t.Width = d.Width
	}
	if t.InputHeight <= 0 {
		t.InputHeight = d.InputHeight
	}
	if t.RowHeight <= 0 {
		t.RowHeight = d.RowHeight
	}
	if t.Padding < 0 {
		t.Padding = d.Padding
	}
	if t.BorderWidth < 0 {
		t.BorderWidth = 0
	}
	if t.Radius < 0 {
		t.Radius = 0
	}
}

// Load reads one theme file.
func Load(path string) (Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, err
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	t, err := Parse(data, base)
	if err != nil {
		return Theme{}, fmt.Errorf("theme %s: %w", path, err)
	}
	return t, nil
}

// LoadAll returns the built-in themes followed by the *.toml themes in dir.
// A user theme with the same name as an earlier one replaces it in place.
// Unreadable files are skipped and reported in the joined error, which can be
// ignored when only the themes matter. A missing dir is not an error.
func LoadAll(dir string) ([]Theme, error) {
	var all []Theme
	for _, mk := range builtins {
		all = append(all, mk())
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return all, nil
	}
	if err != nil {
		return all, err
	}
	var errs []error
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".toml") {
			continue
		}
		t, err := Load(filepath.Join(dir, e.Name()))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		replaced := false
		for i := range all {
			if strings.EqualFold(all[i].Name, t.Name) {
				all[i], replaced = t, true
				break
			}
		}
		if !replaced {
			all = append(all, t)
		}
	}
	return all, errors.Join(errs...)
}

// Resolve picks the theme called name from dir (user themes override
// built-ins). An empty or unknown name yields the default for dark.
func Resolve(name, dir string, dark bool) Theme {
	all, _ := LoadAll(dir)
	for _, t := range all {
		if name != "" && strings.EqualFold(t.Name, name) {
			return t
		}
	}
	return Default(dark)
}
