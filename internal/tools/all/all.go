// Package all links every tools group so their init functions register.
package all

import (
	_ "github.com/avestura/zagros/internal/tools/cipherx"
	_ "github.com/avestura/zagros/internal/tools/dt"
	_ "github.com/avestura/zagros/internal/tools/encode"
	_ "github.com/avestura/zagros/internal/tools/extract"
	_ "github.com/avestura/zagros/internal/tools/hashx"
	_ "github.com/avestura/zagros/internal/tools/lines"
	_ "github.com/avestura/zagros/internal/tools/mathx"
	_ "github.com/avestura/zagros/internal/tools/netx"
	_ "github.com/avestura/zagros/internal/tools/num"
	_ "github.com/avestura/zagros/internal/tools/randx"
	_ "github.com/avestura/zagros/internal/tools/regexx"
	_ "github.com/avestura/zagros/internal/tools/text"
	_ "github.com/avestura/zagros/internal/tools/uuid"
)

// Doc exists so other packages can import all by name.
const Doc = "all"
