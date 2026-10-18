// Package all links every tools group so their init functions register.
package all

import (
	_ "github.com/avestura/zagros/internal/tools/cipherx"
	_ "github.com/avestura/zagros/internal/tools/csvx"
	_ "github.com/avestura/zagros/internal/tools/dt"
	_ "github.com/avestura/zagros/internal/tools/encode"
	_ "github.com/avestura/zagros/internal/tools/extract"
	_ "github.com/avestura/zagros/internal/tools/hashx"
	_ "github.com/avestura/zagros/internal/tools/htmlx"
	_ "github.com/avestura/zagros/internal/tools/jwtx"
	_ "github.com/avestura/zagros/internal/tools/lines"
	_ "github.com/avestura/zagros/internal/tools/mathx"
	_ "github.com/avestura/zagros/internal/tools/netx"
	_ "github.com/avestura/zagros/internal/tools/num"
	_ "github.com/avestura/zagros/internal/tools/pathx"
	_ "github.com/avestura/zagros/internal/tools/randx"
	_ "github.com/avestura/zagros/internal/tools/regexx"
	_ "github.com/avestura/zagros/internal/tools/text"
	_ "github.com/avestura/zagros/internal/tools/tmplx"
	_ "github.com/avestura/zagros/internal/tools/tomlx"
	_ "github.com/avestura/zagros/internal/tools/urlx"
	_ "github.com/avestura/zagros/internal/tools/uuid"
	_ "github.com/avestura/zagros/internal/tools/xmlx"
	_ "github.com/avestura/zagros/internal/tools/yamlx"
)

// Doc exists so other packages can import all by name.
const Doc = "all"
