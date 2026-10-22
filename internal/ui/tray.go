//go:build windows

package ui

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/avestura/zagros/internal/core"
	"golang.org/x/sys/windows"
)

const (
	trayID = 1

	cmdShow = iota + 1
	cmdSettings
	cmdReload
	cmdConfigDir
	cmdExtensions
	cmdAutostart
	cmdUpdate
	cmdMCP
	cmdAgentPrompt
	cmdLogs
	cmdExit
)

func (a *App) trayIconHandle() windows.Handle {
	if a.trayIcon != 0 {
		return a.trayIcon
	}
	if len(a.opt.IconICO) > 0 {
		a.trayIcon = iconFromICO(a.opt.IconICO, 16*max(a.dpi, 96)/96)
	}
	if a.trayIcon == 0 {
		h, _, _ := user32.NewProc("LoadIconW").Call(0, 32512) // IDI_APPLICATION
		a.trayIcon = windows.Handle(h)
	}
	return a.trayIcon
}

func (a *App) trayTip() string {
	hk := a.opt.Config.Hotkey
	if hk == "" {
		hk = "Alt+Space"
	}
	s := "Zagros - ready (" + hk + ")"
	switch {
	case a.hotkeyErr != "":
		s = "Zagros - hotkey unavailable"
	case a.visible:
		s = "Zagros - open"
	}
	if a.trayStateText != "" {
		s += "\n" + a.trayStateText
	}
	return s
}

func copyU16(dst []uint16, s string) {
	w, _ := windows.UTF16FromString(s)
	if len(w) > len(dst) {
		w = w[:len(dst)]
		w[len(w)-1] = 0
	}
	copy(dst, w)
}

func (a *App) nid(flags uint32) *notifyIconData {
	n := &notifyIconData{
		Size:            uint32(unsafe.Sizeof(notifyIconData{})),
		Hwnd:            a.hwnd,
		ID:              trayID,
		Flags:           flags,
		CallbackMessage: msgTray,
		Icon:            a.trayIconHandle(),
	}
	copyU16(n.Tip[:], a.trayTip())
	return n
}

// addTray (re)creates the notification-area icon. It is also called when
// Explorer restarts (TaskbarCreated).
func (a *App) addTray() {
	n := a.nid(nifMessage | nifIcon | nifTip)
	r, _, _ := pShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(n)))
	a.trayAdded = r != 0
	if a.hotkeyErr != "" {
		a.balloon("Zagros", a.hotkeyErr)
	}
}

func (a *App) updateTray() {
	if !a.trayAdded {
		return
	}
	n := a.nid(nifMessage | nifIcon | nifTip)
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(n)))
}

func (a *App) removeTray() {
	if a.trayAdded {
		n := a.nid(0)
		pShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(n)))
		a.trayAdded = false
	}
	if a.trayIcon != 0 {
		pDestroyIcon.Call(uintptr(a.trayIcon))
		a.trayIcon = 0
	}
}

func (a *App) balloon(title, text string) {
	if !a.trayAdded {
		return
	}
	n := a.nid(nifInfo)
	copyU16(n.InfoTitle[:], title)
	copyU16(n.Info[:], text)
	n.InfoFlags = niifInfo
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(n)))
}

func (a *App) onTray(mouse uint32) {
	switch mouse {
	case wmLButtonUp:
		a.toggle()
	case wmRButtonUp:
		a.trayMenu()
	}
}

func (a *App) trayMenu() {
	menu, _, _ := pCreatePopupMenu.Call()
	defer pDestroyMenu.Call(menu)
	add := func(flags uint32, id int, label string) {
		pAppendMenuW.Call(menu, uintptr(flags), uintptr(id), uintptr(unsafe.Pointer(u16(label))))
	}
	ver := strings.TrimPrefix(core.AppVersion, "v")
	if ver == "" {
		ver = "dev"
	}
	add(mfString|mfGrayed, 0, "Zagros "+ver)
	add(mfSeparator, 0, "")
	add(mfString, cmdShow, "Show\tAlt+Space")
	add(mfSeparator, 0, "")
	add(mfString, cmdSettings, "Settings")
	add(mfString, cmdReload, "Reload extensions and themes")
	add(mfString, cmdConfigDir, "Open config folder")
	ext := uint32(mfString)
	if a.opt.Ext == nil && a.opt.Market == nil {
		ext |= mfGrayed
	}
	add(ext, cmdExtensions, "Extensions...")
	as := uint32(mfString)
	if AutostartEnabled() {
		as |= mfChecked
	}
	add(as, cmdAutostart, "Start with Windows")
	upd := uint32(mfString)
	if a.opt.CheckUpdate == nil {
		upd |= mfGrayed
	}
	add(upd, cmdUpdate, "Check for updates")
	mcp := uint32(mfString)
	if a.opt.MCPToggle == nil {
		mcp |= mfGrayed
	} else if a.opt.MCPOn != nil && a.opt.MCPOn() {
		mcp |= mfChecked
	}
	add(mcp, cmdMCP, "MCP server for coding agents (HTTP)")
	ap := uint32(mfString)
	if a.opt.AgentPrompt == nil {
		ap |= mfGrayed
	}
	add(ap, cmdAgentPrompt, "Copy coding-agent prompt")
	lg := uint32(mfString)
	if a.opt.LogsToggle == nil {
		lg |= mfGrayed
	} else if a.opt.LogsOn != nil && a.opt.LogsOn() {
		lg |= mfChecked
	}
	add(lg, cmdLogs, "Log console")
	add(mfSeparator, 0, "")
	add(mfString, cmdExit, "Exit")

	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	a.inMenu = true
	pSetForegroundWindow.Call(uintptr(a.hwnd)) // required so the menu closes on outside click
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmReturnCmd|tpmRightButton|tpmBottomAlign,
		uintptr(pt.X), uintptr(pt.Y), 0, uintptr(a.hwnd), 0)
	postMessage(a.hwnd, 0, 0, 0) // WM_NULL
	a.inMenu = false
	a.runTrayCommand(int(cmd))
}

func (a *App) runTrayCommand(cmd int) {
	switch cmd {
	case cmdShow:
		a.showLauncher()
	case cmdSettings:
		if a.opt.ConfigPath != "" {
			path := a.opt.ConfigPath
			go func() { // ShellExecute can block on file associations
				if err := ShellOpen("open", path, "", ""); err != nil {
					// no association for .toml: fall back to Notepad
					if err2 := ShellOpen("open", "notepad.exe", `"`+path+`"`, ""); err2 != nil {
						a.Notify("Zagros", err.Error())
					}
				}
			}()
		}
	case cmdReload:
		if a.opt.Reload == nil {
			return
		}
		go func() {
			err := a.opt.Reload()
			if err != nil {
				a.Notify("Zagros", "Reload failed: "+err.Error())
			} else {
				a.Notify("Zagros", "Extensions and themes reloaded")
			}
		}()
	case cmdConfigDir:
		if a.opt.ConfigDir != "" {
			dir := a.opt.ConfigDir
			go ShellOpen("open", dir, "", "")
		}
	case cmdExtensions:
		a.openManager()
	case cmdAutostart:
		on := !AutostartEnabled()
		if err := SetAutostart(on); err != nil {
			a.balloon("Zagros", fmt.Sprintf("Cannot change autostart: %v", err))
			return
		}
		if a.opt.OnAutostart != nil {
			go a.opt.OnAutostart(on) // writes the config file
		}
	case cmdUpdate:
		if a.opt.CheckUpdate != nil {
			go func() { a.Notify("Zagros", a.opt.CheckUpdate()) }()
		}
	case cmdMCP:
		if a.opt.MCPToggle != nil {
			go func() { a.Notify("Zagros", a.opt.MCPToggle()) }()
		}
	case cmdAgentPrompt:
		if a.opt.AgentPrompt != nil {
			SetClipboard(a.opt.AgentPrompt())
			a.balloon("Zagros", "Agent prompt copied. Paste it into your coding agent.")
		}
	case cmdLogs:
		if a.opt.LogsToggle != nil {
			go a.opt.LogsToggle()
		}
	case cmdExit:
		a.shutdown()
	}
}
