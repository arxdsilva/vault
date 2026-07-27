//go:build windows

// Package windows implements the Windows platform adapter: Recycle Bin
// semantics and any other OS-specific behavior that must stay out of
// internal/core.
package windows

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Win32 SHFileOperationW constants (shellapi.h). Not exposed by
// golang.org/x/sys/windows, so declared directly here.
const (
	foDelete          = 0x0003
	fofAllowUndo      = 0x0040
	fofNoConfirmation = 0x0010
	fofSilent         = 0x0004
)

type shFileOpStruct struct {
	hwnd                  uintptr
	wFunc                 uint32
	pFrom                 *uint16
	pTo                   *uint16
	fFlags                uint16
	fAnyOperationsAborted int32
	hNameMappings         uintptr
	lpszProgressTitle     *uint16
}

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	procSHFileOperation = shell32.NewProc("SHFileOperationW")
)

// Trash implements core.Trasher for Windows via SHFileOperationW with
// FOF_ALLOWUNDO, the same call Explorer makes for a Recycle Bin delete.
type Trash struct {
	// SystemDrive overrides which drive's $Recycle.Bin Location resolves
	// (used by tests); defaults to the SystemDrive environment variable.
	SystemDrive string
	// SID overrides the current user's SID subfolder (used by tests); the
	// zero value resolves it from the OS lazily.
	SID string
}

// NewTrash returns a Trash ready for use.
func NewTrash() *Trash {
	return &Trash{}
}

func (t *Trash) systemDrive() string {
	if t.SystemDrive != "" {
		return t.SystemDrive
	}
	if sd := os.Getenv("SystemDrive"); sd != "" {
		return sd
	}
	return "C:"
}

func (t *Trash) sid() (string, error) {
	if t.SID != "" {
		return t.SID, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Uid, nil
}

// Location returns the current user's $Recycle.Bin subfolder on the system
// drive, creating it if it doesn't exist yet. A real, browsable directory,
// but (unlike Finder's ~/.Trash or Linux's XDG trash) its entries are
// opaque, shell-obfuscated names ("$Rxxxxxx.ext"/"$Ixxxxxx.ext" pairs) —
// Explorer's own Recycle Bin view resolves those to original names via
// Shell API, which this project doesn't implement yet.
func (t *Trash) Location() (string, error) {
	sid, err := t.sid()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(t.systemDrive()+`\`, "$Recycle.Bin", sid)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// Trash moves path to the Recycle Bin.
func (t *Trash) Trash(path string) error {
	// pFrom must be a list of null-terminated strings ending in an extra
	// null (double-null-terminated), per the SHFileOperationW contract.
	from, err := syscall.UTF16FromString(path)
	if err != nil {
		return err
	}
	from = append(from, 0)

	op := shFileOpStruct{
		wFunc:  foDelete,
		pFrom:  &from[0],
		fFlags: fofAllowUndo | fofNoConfirmation | fofSilent,
	}
	ret, _, _ := procSHFileOperation.Call(uintptr(unsafe.Pointer(&op)))
	if ret != 0 {
		return fmt.Errorf("SHFileOperationW: error code %d", ret)
	}
	if op.fAnyOperationsAborted != 0 {
		return fmt.Errorf("SHFileOperationW: operation aborted")
	}
	return nil
}
