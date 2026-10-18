//go:build windows

// Command zagros is the Zagros launcher. Without arguments it runs the
// launcher (tray icon, hotkey, window); with arguments it acts as the CLI.
package main

import (
	"fmt"
	"os"

	"github.com/avestura/zagros/internal/core"
	_ "github.com/avestura/zagros/internal/tools/all" // registers the tools library
	"github.com/avestura/zagros/internal/ui"
)

// version is set with -ldflags "-X main.version=...".
var version = "dev"

const usage = `Zagros - keyboard launcher for Windows

Usage:
  zagros                              run the launcher (tray icon + hotkey)
  zagros --version                    print the version
  zagros --config                     print the config file path and open it
  zagros --bench-input                type into the real launcher and report key->screen latency
  zagros --selftest                   render frames offscreen, run a typing-burst latency check and exit
  zagros --trace                      run with timing log in the data dir (trace.log); also ZAGROS_TRACE=1
  zagros mcp [--http]                 MCP server so coding agents can search the extension API
  zagros mcp prompt                   the prompt to give your agent
  zagros update [--json]              check GitHub for a newer release (never installs)
  zagros ext install <ref>            install from a .zagros file, folder, github:owner/repo[@tag], URL or registry id
  zagros ext list                     list installed extensions
  zagros ext remove <id> [--keep-data]
  zagros ext pack <dir> [-o file]     build a .zagros archive
  zagros ext init <dir>               new extension project with API stubs (.pyi) and a coding-agent guide
  zagros ext stubs <dir>              refresh the stubs, AGENTS.md and API_INDEX.md of a project
  zagros ext validate <dir>           check an extension folder
  zagros theme list                   list built-in, user and installed themes
  zagros theme install <id|https url to .toml>
  zagros theme remove <name>          delete a user or installed theme
  zagros market list                  list everything in the registry
  zagros market search <query>
  zagros ai providers|key set <name>|key delete <name>|test [name]   language-model providers for extensions
`

func main() {
	core.AppVersion = version
	args := stripTraceFlag(os.Args[1:])
	if len(args) == 0 {
		os.Exit(runApp())
	}
	attachConsole()
	switch args[0] {
	case "--version", "-v", "version":
		fmt.Println("zagros", version)
	case "--config", "config":
		p := core.ConfigPath()
		if _, err := core.LoadConfig(p); err != nil { // creates the default file if missing
			fmt.Fprintln(os.Stderr, "warning:", err)
		}
		fmt.Println(p)
		if err := ui.ShellOpen("open", p, "", ""); err != nil {
			_ = ui.ShellOpen("open", "notepad.exe", `"`+p+`"`, "")
		}
	case "--bench-input":
		os.Exit(runBench())
	case "--selftest":
		os.Exit(selfTest())
	case "mcp":
		os.Exit(cmdMCP(args[1:]))
	case "update":
		os.Exit(cmdUpdate(args[1:]))
	case "ext":
		os.Exit(cmdExt(args[1:]))
	case "theme":
		os.Exit(cmdTheme(args[1:]))
	case "market":
		os.Exit(cmdMarket(args[1:]))
	case "ai":
		os.Exit(cmdAI(args[1:]))
	case "--help", "-h", "help":
		fmt.Print(usage)
	default:
		fmt.Fprint(os.Stderr, "unknown command: ", args[0], "\n\n", usage)
		os.Exit(2)
	}
}

// stripTraceFlag removes --trace from args and switches tracing on (also when
// ZAGROS_TRACE is set). Events go to DataDir/trace.log; with tracing off the
// instrumentation costs one atomic load per event.
func stripTraceFlag(args []string) []string {
	var out []string
	want := core.TraceRequested()
	for _, a := range args {
		if a == "--trace" {
			want = true
			continue
		}
		out = append(out, a)
	}
	if want {
		if err := core.Trace.EnableFile(); err != nil {
			fmt.Fprintln(os.Stderr, "zagros: cannot open trace log:", err)
		}
	}
	return out
}
