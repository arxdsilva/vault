//go:build windows

package windows

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"unsafe"

	"github.com/asilva/vault/internal/core"
)

// Win32 GetDriveType return values (fileapi.h).
const (
	driveRemovable = 2
	driveFixed     = 3
	driveRemote    = 4
	driveCDROM     = 5
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives     = kernel32.NewProc("GetLogicalDrives")
	procGetDriveType         = kernel32.NewProc("GetDriveTypeW")
	procGetVolumeInformation = kernel32.NewProc("GetVolumeInformationW")
	procGetDiskFreeSpaceEx   = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// Volumes implements core.VolumeLister and core.VolumeEjector for Windows
// by enumerating drive letters via the Win32 API — the same set Explorer's
// "This PC" shows.
type Volumes struct {
	// SystemDrive overrides boot-drive detection (used by tests); defaults
	// to the SystemDrive environment variable (e.g. "C:").
	SystemDrive string
	// run executes an external command and returns its combined output;
	// overridable in tests so Eject never has to shell out for real.
	run func(name string, args ...string) ([]byte, error)
}

// NewVolumes returns a Volumes ready for use against the real drive list.
func NewVolumes() *Volumes {
	return &Volumes{}
}

func (v *Volumes) systemDrive() string {
	if v.SystemDrive != "" {
		return strings.ToUpper(v.SystemDrive)
	}
	if sd := os.Getenv("SystemDrive"); sd != "" {
		return strings.ToUpper(sd)
	}
	return "C:"
}

func (v *Volumes) runner() func(string, ...string) ([]byte, error) {
	if v.run != nil {
		return v.run
	}
	return func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
}

// List returns every visible drive letter, including the system drive
// (shown but not ejectable, since it can't be unmounted while running
// from it).
func (v *Volumes) List() ([]core.Volume, error) {
	mask, _, _ := procGetLogicalDrives.Call()
	if mask == 0 {
		return nil, fmt.Errorf("GetLogicalDrives: no drives found")
	}

	sysDrive := v.systemDrive()
	volumes := make([]core.Volume, 0, 8)
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		drive := string(rune('A'+i)) + ":"
		root := drive + `\`

		driveType, _, _ := procGetDriveType.Call(strPtr(root))
		switch driveType {
		case driveRemovable, driveFixed, driveRemote, driveCDROM:
		default:
			continue // DRIVE_UNKNOWN/DRIVE_NO_ROOT_DIR: empty bay or nothing mounted
		}

		name := volumeLabel(root)
		if name == "" {
			name = drive
		}

		var free, total, totalFree uint64
		procGetDiskFreeSpaceEx.Call(
			strPtr(root),
			uintptr(unsafe.Pointer(&free)),
			uintptr(unsafe.Pointer(&total)),
			uintptr(unsafe.Pointer(&totalFree)),
		)

		volumes = append(volumes, core.Volume{
			Path:       root,
			Name:       name,
			Ejectable:  !strings.EqualFold(drive, sysDrive),
			TotalBytes: total,
			FreeBytes:  free,
		})
	}

	sort.Slice(volumes, func(i, j int) bool { return volumes[i].Name < volumes[j].Name })
	return volumes, nil
}

func volumeLabel(root string) string {
	buf := make([]uint16, 261)
	ok, _, _ := procGetVolumeInformation.Call(
		strPtr(root),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
		0, 0, 0, 0, 0,
	)
	if ok == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

func strPtr(s string) uintptr {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return 0
	}
	return uintptr(unsafe.Pointer(p))
}

// Eject ejects the drive at path (e.g. "D:\") via the Shell.Application COM
// eject verb, the same mechanism Explorer's own "Eject" context-menu item
// uses.
func (v *Volumes) Eject(path string) error {
	drive := strings.TrimSuffix(path, `\`)
	script := fmt.Sprintf(
		`(New-Object -ComObject Shell.Application).Namespace(17).ParseName('%s').InvokeVerb('Eject')`,
		drive,
	)
	out, err := v.runner()("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if err != nil {
		return fmt.Errorf("eject %s: %w: %s", path, err, strings.TrimSpace(string(out)))
	}
	return nil
}
