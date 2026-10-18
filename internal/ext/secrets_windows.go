//go:build windows

package ext

import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	advapi        = windows.NewLazySystemDLL("advapi32.dll")
	procCredWrite = advapi.NewProc("CredWriteW")
	procCredRead  = advapi.NewProc("CredReadW")
	procCredDel   = advapi.NewProc("CredDeleteW")
	procCredFree  = advapi.NewProc("CredFree")
)

const (
	credTypeGeneric     = 1
	credPersistLocalMac = 2
	errNotFound         = syscall.Errno(1168)
)

type credentialW struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        windows.Filetime
	CredentialBlobSize uint32
	CredentialBlob     *byte
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

type credManager struct{}

func init() { secretBackend = credManager{} }

func (credManager) Set(target, value string) error {
	t, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	u, _ := windows.UTF16PtrFromString("zagros")
	blob := []byte(value)
	c := credentialW{Type: credTypeGeneric, TargetName: t, Persist: credPersistLocalMac, UserName: u, CredentialBlobSize: uint32(len(blob))}
	if len(blob) > 0 {
		c.CredentialBlob = &blob[0]
	}
	r, _, e := procCredWrite.Call(uintptr(unsafe.Pointer(&c)), 0)
	if r == 0 {
		return e
	}
	return nil
}

func (credManager) Get(target string) (string, bool, error) {
	t, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return "", false, err
	}
	var p *credentialW
	r, _, e := procCredRead.Call(uintptr(unsafe.Pointer(t)), credTypeGeneric, 0, uintptr(unsafe.Pointer(&p)))
	if r == 0 {
		if errors.Is(e, errNotFound) {
			return "", false, nil
		}
		return "", false, e
	}
	defer procCredFree.Call(uintptr(unsafe.Pointer(p)))
	if p.CredentialBlobSize == 0 || p.CredentialBlob == nil {
		return "", true, nil
	}
	return string(unsafe.Slice(p.CredentialBlob, p.CredentialBlobSize)), true, nil
}

func (credManager) Delete(target string) error {
	t, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	r, _, e := procCredDel.Call(uintptr(unsafe.Pointer(t)), credTypeGeneric, 0)
	if r == 0 && !errors.Is(e, errNotFound) {
		return e
	}
	return nil
}
