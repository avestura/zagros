package builtin

import (
	"context"
	"net"
	"os/exec"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

type system struct {
	base
	cmds []sysCmd
}

type sysCmd struct {
	key, title, sub string
	act             core.Action
	confirm         bool
}

func runCmd(name string, args ...string) core.Action {
	return core.Action{Kind: "func", Func: func() error {
		c := exec.Command(name, args...)
		hideCmd(c)
		return c.Run()
	}}
}

func newSystem() *system {
	s := &system{base: base{"system"}}
	s.cmds = []sysCmd{
		{"lock", "Lock", "Lock this PC", runCmd("rundll32.exe", "user32.dll,LockWorkStation"), false},
		{"sleep", "Sleep", "Suspend to RAM", runCmd("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0"), false},
		{"hibernate", "Hibernate", "Suspend to disk", runCmd("shutdown", "/h"), true},
		{"shutdown", "Shut down", "Power off this PC", runCmd("shutdown", "/s", "/t", "0"), true},
		{"restart reboot", "Restart", "Reboot this PC", runCmd("shutdown", "/r", "/t", "0"), true},
		{"sign out log off logoff", "Sign out", "End the current session", runCmd("shutdown", "/l"), true},
		{"empty recycle bin trash", "Empty Recycle Bin", "Permanently delete trashed files",
			runCmd("powershell.exe", "-NoProfile", "-Command", "Clear-RecycleBin -Force -ErrorAction SilentlyContinue"), true},
		{"task manager taskmgr", "Task Manager", "Open Task Manager", core.Action{Kind: "run", Target: "taskmgr.exe"}, false},
		{"settings", "Settings", "Open Windows Settings", core.Action{Kind: "open", Target: "ms-settings:"}, false},
		{"control panel", "Control Panel", "Open Control Panel", core.Action{Kind: "run", Target: "control.exe"}, false},
		{"services", "Services", "Open the Services console", core.Action{Kind: "open", Target: "services.msc"}, false},
		{"device manager devices", "Device Manager", "Open Device Manager", core.Action{Kind: "open", Target: "devmgmt.msc"}, false},
		{"dark mode light mode theme toggle", "Toggle dark mode", "Switch Windows between light and dark",
			core.Action{Kind: "func", Func: toggleDarkMode}, false},
		{"flush dns", "Flush DNS cache", "ipconfig /flushdns", runCmd("ipconfig", "/flushdns"), false},
	}
	return s
}

func (s *system) Ranked() bool       { return false }
func (s *system) Keywords() []string { return []string{"ip"} }

func (s *system) Query(ctx context.Context, q string) []core.Item {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	if _, _, ok := trigger(q, "ip"); ok {
		return localIPItems()
	}
	if len(q) < 2 {
		return nil
	}
	var out []core.Item
	for _, c := range s.cmds {
		if !subseq(q, c.title) && !subseq(q, c.key) {
			continue
		}
		it := core.Item{
			ID: "system:" + strings.Fields(c.key)[0], Title: c.title, Subtitle: c.sub, Category: "System", Action: c.act,
		}
		if c.confirm {
			a := it.Action
			a.Extra = map[string]string{"confirm": "1"}
			it.Action = a
		}
		out = append(out, it)
	}
	return out
}

func localIPItems() []core.Item {
	var out []core.Item
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifc := range ifs {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || ipn.IP.IsLinkLocalUnicast() {
				continue
			}
			ip := ipn.IP.String()
			out = append(out, core.Item{
				ID: "system:ip:" + ifc.Name + ":" + ip, Title: ip, Subtitle: ifc.Name, Category: "Network",
				Action: copyAction(ip),
			})
		}
	}
	if len(out) == 0 {
		return []core.Item{hint("system", "ip", "No active network address", "")}
	}
	return out
}
