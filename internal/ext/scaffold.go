package ext

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/avestura/zagros/stubs"
)

const scaffoldManifest = `id = "%s"
name = "%s"
version = "0.1.0"
description = "Describe what %s does."
author = "Your name"
license = "MIT"
entry = "main.star"
keywords = ["%s"]

# Declare what the extension needs; the user approves it. See zagros-stubs/zagros/permissions.pyi.
# permissions = ["network:api.example.com", "clipboard.write"]
permissions = ["clipboard.write"]

[[permission_detail]]
permission = "clipboard.write"
reason = "Copies the result when you press Enter."
`

const scaffoldMain = `# %s: type "%s <text>" in Zagros.
# API: zagros-stubs/zagros/*.pyi. Guide for coding agents: AGENTS.md.

def on_query(ctx, query):
    text = query or "world"
    return [
        item(
            id = "greet",
            title = "Hello, " + text,
            subtitle = "Enter copies it. Tab shows other styles.",
            action = action_copy("Hello, " + text),
            # Tab enters a sub-context: the input shows a "Style" chip and
            # on_scope below supplies the list.
            scope = scope(id = "style", label = "Style", placeholder = "Filter styles", query = query),
        ),
    ]

def on_scope(ctx, scope_id, query, data):
    text = query or "world"
    styles = ["Hello, " + text, "HELLO, " + text.upper(), "hi " + text.lower()]
    return [item(id = "s%%d" %% i, title = s, action = action_copy(s)) for i, s in enumerate(styles)]
`

var scaffoldIDRe = regexp.MustCompile(`[^a-z0-9._-]+`)

// Scaffold creates a new extension project in dir: manifest, a working
// main.star that uses a scope, the .pyi API stubs and the coding-agent guide
// (AGENTS.md, with CLAUDE.md pointing at it). It refuses to touch a folder that
// already has a manifest.
func Scaffold(dir, id, name string) error {
	if _, err := os.Stat(filepath.Join(dir, ManifestFile)); err == nil {
		return fmt.Errorf("%s already contains a %s", dir, ManifestFile)
	}
	base := filepath.Base(filepath.Clean(dir))
	if name == "" {
		name = strings.Title(strings.NewReplacer("-", " ", "_", " ", ".", " ").Replace(base)) //nolint:staticcheck
	}
	if id == "" {
		id = "local." + strings.Trim(scaffoldIDRe.ReplaceAllString(strings.ToLower(base), "-"), "-.")
	}
	kw := strings.Trim(scaffoldIDRe.ReplaceAllString(strings.ToLower(strings.Fields(name + " x")[0]), ""), "-")
	if kw == "" {
		kw = "hello"
	}
	manifest := fmt.Sprintf(scaffoldManifest, id, name, name, kw)
	if _, err := ParseManifest([]byte(manifest)); err != nil {
		return fmt.Errorf("generated manifest is invalid (pass --id): %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	write := func(rel, text string) error {
		return os.WriteFile(filepath.Join(dir, rel), []byte(text), 0o644)
	}
	if err := write(ManifestFile, manifest); err != nil {
		return err
	}
	if err := write("main.star", fmt.Sprintf(scaffoldMain, name, kw)); err != nil {
		return err
	}
	if err := write("AGENTS.md", string(stubs.AgentGuide())); err != nil {
		return err
	}
	if err := write("CLAUDE.md", "@AGENTS.md\n"); err != nil {
		return err
	}
	return stubs.WriteStubs(filepath.Join(dir, StubsDir))
}

// WriteStubs refreshes the .pyi stubs (and AGENTS.md) of an existing project.
func WriteStubs(dir string) error {
	if err := stubs.WriteStubs(filepath.Join(dir, StubsDir)); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "AGENTS.md"), stubs.AgentGuide(), 0o644)
}
