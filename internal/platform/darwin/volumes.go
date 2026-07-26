//go:build darwin

package darwin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/asilva/vault/internal/core"
	"golang.org/x/sys/unix"
)

// Volumes implements core.VolumeLister and core.VolumeEjector for macOS by
// reading /Volumes, the same directory Finder's sidebar watches — every
// mounted disk (external/USB, mounted disk image, network share) appears
// there as a real directory, and the boot volume appears only as a symlink
// (e.g. "Macintosh HD" -> /).
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

// List returns every mounted volume under /Volumes except the boot volume,
// which is already reachable via the user's home directory and would
// otherwise duplicate it in the rail.
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

		// The boot volume shows up under /Volumes as a symlink to "/";
		// every other entry is a real mount point.
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		var total, free uint64
		var stat unix.Statfs_t
		if err := unix.Statfs(path, &stat); err == nil {
			total = uint64(stat.Bsize) * stat.Blocks
			free = uint64(stat.Bsize) * stat.Bavail
		}

		volumes = append(volumes, core.Volume{
			Path:       path,
			Name:       e.Name(),
			Ejectable:  true,
			TotalBytes: total,
			FreeBytes:  free,
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
