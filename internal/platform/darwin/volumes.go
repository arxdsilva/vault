//go:build darwin

package darwin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arxdsilva/vault/internal/core"
	"golang.org/x/sys/unix"
)

// Volumes implements core.VolumeLister and core.VolumeEjector for macOS by
// reading /Volumes, the same directory Finder's sidebar watches — every
// mounted disk (internal, external/USB, mounted disk image, network share)
// appears there, either as a real directory or, for the boot volume, a
// symlink (e.g. "Macintosh HD" -> /).
type Volumes struct {
	// Root overrides /Volumes (used by tests).
	Root string
	// run executes an external command and returns its combined output;
	// overridable in tests so Eject never has to shell out for real.
	run func(name string, args ...string) ([]byte, error)
}

// NewVolumes returns a Volumes ready for use against the real /Volumes.
func NewVolumes() *Volumes {
	return &Volumes{}
}

func (v *Volumes) root() string {
	if v.Root != "" {
		return v.Root
	}
	return "/Volumes"
}

func (v *Volumes) runner() func(string, ...string) ([]byte, error) {
	if v.run != nil {
		return v.run
	}
	return func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
}

// List returns every visible disk under /Volumes, including the boot
// volume (shown but not ejectable, since it can't be unmounted while
// running from it).
func (v *Volumes) List() ([]core.Volume, error) {
	root := v.root()
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	volumes := make([]core.Volume, 0, len(entries))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(root, e.Name())

		// The boot volume shows up under /Volumes as a symlink to "/"; any
		// other symlink is unexpected and skipped rather than guessed at.
		ejectable := true
		info, err := os.Lstat(path)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil || target != "/" {
				continue
			}
			ejectable = false
		}

		var total, free uint64
		var isNetwork bool
		var stat unix.Statfs_t
		if err := unix.Statfs(path, &stat); err == nil {
			total = uint64(stat.Bsize) * stat.Blocks
			free = uint64(stat.Bsize) * stat.Bavail
			isNetwork = isNetworkFSType(fstypeString(stat.Fstypename))
		}

		volumes = append(volumes, core.Volume{
			Path:       path,
			Name:       e.Name(),
			Ejectable:  ejectable,
			TotalBytes: total,
			FreeBytes:  free,
			IsNetwork:  isNetwork,
		})
	}

	sort.Slice(volumes, func(i, j int) bool { return volumes[i].Name < volumes[j].Name })
	return volumes, nil
}

// Eject unmounts path via diskutil, the same mechanism Finder's sidebar
// eject button uses.
func (v *Volumes) Eject(path string) error {
	out, err := v.runner()("diskutil", "eject", path)
	if err != nil {
		return fmt.Errorf("diskutil eject %s: %w: %s", path, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// fstypeString converts Statfs_t's fixed-size, NUL-terminated Fstypename
// buffer into a plain Go string.
func fstypeString(raw [16]byte) string {
	n := bytes.IndexByte(raw[:], 0)
	if n < 0 {
		n = len(raw)
	}
	return string(raw[:n])
}

// isNetworkFSType reports whether fstype (as macOS's statfs reports it,
// e.g. "smbfs", "nfs", "afpfs", "webdav") names a network filesystem — an
// already-mounted share, not live network-neighborhood discovery.
func isNetworkFSType(fstype string) bool {
	switch fstype {
	case "smbfs", "afpfs", "nfs", "webdav", "ftp":
		return true
	default:
		return false
	}
}
