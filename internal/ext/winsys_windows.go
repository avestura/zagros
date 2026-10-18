//go:build windows

package ext

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

var (
	user32            = windows.NewLazySystemDLL("user32.dll")
	dwmapi            = windows.NewLazySystemDLL("dwmapi.dll")
	procGetTitleLen   = user32.NewProc("GetWindowTextLengthW")
	procGetTitle      = user32.NewProc("GetWindowTextW")
	procGetWindowLong = user32.NewProc("GetWindowLongW")
	procDwmAttr       = dwmapi.NewProc("DwmGetWindowAttribute")
)

const (
	gwlExStyle     = ^uintptr(19) // -20
	wsExToolWindow = 0x80
	dwmwaCloaked   = 14
)

func osProcesses() ([]sysProc, error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snap)
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	var out []sysProc
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		out = append(out, sysProc{Name: windows.UTF16ToString(e.ExeFile[:]), PID: int(e.ProcessID)})
	}
	return out, nil
}

func osWindows() ([]sysWindow, error) {
	procs, _ := osProcesses()
	names := map[int]string{}
	for _, p := range procs {
		names[p.PID] = p.Name
	}
	var out []sysWindow
	cb := syscall.NewCallback(func(h uintptr, _ uintptr) uintptr {
		hw := windows.HWND(h)
		if !windows.IsWindowVisible(hw) {
			return 1
		}
		if st, _, _ := procGetWindowLong.Call(h, gwlExStyle); st&wsExToolWindow != 0 {
			return 1
		}
		var cloaked uint32
		if r, _, _ := procDwmAttr.Call(h, dwmwaCloaked, uintptr(unsafe.Pointer(&cloaked)), 4); r == 0 && cloaked != 0 {
			return 1
		}
		n, _, _ := procGetTitleLen.Call(h)
		if n == 0 {
			return 1
		}
		buf := make([]uint16, n+1)
		procGetTitle.Call(h, uintptr(unsafe.Pointer(&buf[0])), n+1)
		title := windows.UTF16ToString(buf)
		if title == "" {
			return 1
		}
		var pid uint32
		windows.GetWindowThreadProcessId(hw, &pid)
		out = append(out, sysWindow{Title: title, Process: names[int(pid)], PID: int(pid)})
		return 1
	})
	if err := windows.EnumWindows(cb, nil); err != nil {
		return nil, err
	}
	return out, nil
}

var stateNames = map[svc.State]string{
	svc.Stopped: "stopped", svc.StartPending: "start_pending", svc.StopPending: "stop_pending",
	svc.Running: "running", svc.ContinuePending: "continue_pending", svc.PausePending: "pause_pending", svc.Paused: "paused",
}

var startNames = map[uint32]string{
	mgr.StartAutomatic: "automatic", mgr.StartManual: "manual", mgr.StartDisabled: "disabled", 0: "boot", 1: "system",
}

func openSCM() (*mgr.Mgr, error) {
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, err
	}
	return &mgr.Mgr{Handle: h}, nil
}

func queryService(m *mgr.Mgr, name string) (sysService, error) {
	n, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return sysService{}, err
	}
	h, err := windows.OpenService(m.Handle, n, windows.SERVICE_QUERY_STATUS|windows.SERVICE_QUERY_CONFIG)
	if err != nil {
		if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
			return sysService{}, fmt.Errorf("service %q not found", name)
		}
		return sysService{}, err
	}
	s := &mgr.Service{Name: name, Handle: h}
	defer s.Close()
	st, err := s.Query()
	if err != nil {
		return sysService{}, err
	}
	out := sysService{Name: name, State: stateNames[st.State], PID: int(st.ProcessId)}
	if cfg, err := s.Config(); err == nil {
		out.Display = cfg.DisplayName
		out.StartType = startNames[cfg.StartType]
	}
	return out, nil
}

func osServices() ([]sysService, error) {
	m, err := openSCM()
	if err != nil {
		return nil, err
	}
	defer m.Disconnect()
	names, err := m.ListServices()
	if err != nil {
		return nil, err
	}
	var out []sysService
	for _, n := range names {
		if s, err := queryService(m, n); err == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

func osServiceStatus(name string) (sysService, error) {
	m, err := openSCM()
	if err != nil {
		return sysService{}, err
	}
	defer m.Disconnect()
	return queryService(m, name)
}

var knownFolders = map[string]*windows.KNOWNFOLDERID{
	"documents": windows.FOLDERID_Documents, "downloads": windows.FOLDERID_Downloads, "desktop": windows.FOLDERID_Desktop,
	"pictures": windows.FOLDERID_Pictures, "music": windows.FOLDERID_Music, "videos": windows.FOLDERID_Videos,
	"appdata": windows.FOLDERID_RoamingAppData, "localappdata": windows.FOLDERID_LocalAppData,
	"programdata": windows.FOLDERID_ProgramData, "profile": windows.FOLDERID_Profile,
}

func osKnownFolder(name string) (string, error) {
	id, ok := knownFolders[strings.ToLower(name)]
	if !ok {
		return "", fmt.Errorf("unknown folder %q", name)
	}
	return windows.KnownFolderPath(id, 0)
}

func hive(root string) registry.Key {
	if root == "HKLM" {
		return registry.LOCAL_MACHINE
	}
	return registry.CURRENT_USER
}

func openKey(root, path string) (registry.Key, error) {
	return registry.OpenKey(hive(root), path, registry.QUERY_VALUE|registry.ENUMERATE_SUB_KEYS)
}

func readValue(k registry.Key, name string) (any, error) {
	_, typ, err := k.GetValue(name, nil)
	if err != nil {
		return nil, err
	}
	switch typ {
	case registry.SZ, registry.EXPAND_SZ:
		s, _, err := k.GetStringValue(name)
		return s, err
	case registry.DWORD, registry.QWORD:
		v, _, err := k.GetIntegerValue(name)
		return int64(v), err
	case registry.MULTI_SZ:
		s, _, err := k.GetStringsValue(name)
		return s, err
	}
	b, _, err := k.GetBinaryValue(name)
	return b, err
}

func osRegistryGet(root, path, name string) (any, error) {
	k, err := openKey(root, path)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	v, err := readValue(k, name)
	if errors.Is(err, registry.ErrNotExist) {
		return nil, nil
	}
	return v, err
}

func osRegistryKeys(root, path string) ([]string, error) {
	k, err := openKey(root, path)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	return k.ReadSubKeyNames(-1)
}

func osRegistryValues(root, path string) (map[string]any, error) {
	k, err := openKey(root, path)
	if err != nil {
		return nil, err
	}
	defer k.Close()
	names, err := k.ReadValueNames(-1)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	for _, n := range names {
		if v, err := readValue(k, n); err == nil {
			out[n] = v
		}
	}
	return out, nil
}
