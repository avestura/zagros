//go:build !windows

package ext

import "errors"

var errWindowsOnly = errors.New("the windows module is only available on Windows")

func osProcesses() ([]sysProc, error)                         { return nil, errWindowsOnly }
func osWindows() ([]sysWindow, error)                         { return nil, errWindowsOnly }
func osServices() ([]sysService, error)                       { return nil, errWindowsOnly }
func osServiceStatus(string) (sysService, error)              { return sysService{}, errWindowsOnly }
func osKnownFolder(string) (string, error)                    { return "", errWindowsOnly }
func osRegistryGet(string, string, string) (any, error)       { return nil, errWindowsOnly }
func osRegistryKeys(string, string) ([]string, error)         { return nil, errWindowsOnly }
func osRegistryValues(string, string) (map[string]any, error) { return nil, errWindowsOnly }
