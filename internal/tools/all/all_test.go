package all

import (
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

// TestEveryGroupIsRegisteredAndWellFormed checks the whole registry. Groups
// that have not been filled in yet are reported by their own packages.
func TestEveryGroupIsWellFormed(t *testing.T) { toolstest.Check(t, "") }
