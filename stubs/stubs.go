// Package stubs embeds the extension API stubs (.pyi) and the coding-agent
// guide that `zagros ext init` and `zagros ext stubs` write next to an
// extension, so people and coding agents can search the API.
package stubs

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed zagros/*.pyi AGENTS.md
var files embed.FS

// FS exposes the embedded files: zagros/*.pyi and AGENTS.md.
func FS() fs.FS { return files }

// Names lists the stub files (zagros/*.pyi).
func Names() []string {
	ents, _ := fs.ReadDir(files, "zagros")
	var out []string
	for _, e := range ents {
		out = append(out, e.Name())
	}
	return out
}

// WriteStubs writes the .pyi files into dir/zagros/.
func WriteStubs(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "zagros"), 0o755); err != nil {
		return err
	}
	for _, n := range Names() {
		b, err := files.ReadFile("zagros/" + n)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "zagros", n), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// AgentGuide returns the AGENTS.md text.
func AgentGuide() []byte {
	b, _ := files.ReadFile("AGENTS.md")
	return b
}
