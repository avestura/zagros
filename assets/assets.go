// Package assets embeds binary assets used by the application.
package assets

import _ "embed"

// Icon is the application icon (.ico with 16..256 px images).
//
//go:embed zagros.ico
var Icon []byte
