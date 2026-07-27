//go:build linux

// Package linux implements the Linux platform adapter: trash semantics,
// mount enumeration, and any other OS-specific behavior that must stay out
// of internal/core.
package linux

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arxdsilva/vault/internal/core"
	"golang.org/x/sys/unix"
)

// pseudoFSTypes are virtual/kernel filesystems that never represent a real,
// user-relevant disk or share — skipped so the rail isn't cluttered with
// dozens of kernel bookkeeping mounts every Linux system has.
var pseudoFSTypes = map[string]bool{
	"proc": true, "sysfs": true, "devtmpfs": true, "devpts": true,
	"tmpfs": true, "cgroup": true, "cgroup2": true, "mqueue": true,
	"debugfs": true, "tracefs": true, "securityfs": true, "pstore": true,
	"bpf": true, "autofs": true, "hugetlbfs": true, "configfs": true,
	"fusectl": true, "rpc_pipefs": true, "binfmt_misc": true, "squashfs": true,
	"overlay": true, "efivarfs": true, "ramfs": true,
}

// networkFSTypes are filesystem types /proc/mounts reports for an
// already-mounted network share — not live network-neighborhood discovery.
var networkFSTypes = map[string]bool{
	"nfs": true, "nfs4": true, "nfs3": true, "cifs": true, "smbfs": true,
	"smb3": true, "davfs": true, "fuse.sshfs": true, "fuse.davfs": true,
	"fuse.davfs2": true,
}

// Volumes implements core.VolumeLister and core.VolumeEjector for Linux by
// reading /proc/mounts — the root filesystem (shown but not ejectable, same
// as the boot volume on macOS/Windows) plus anything mounted under /media
// or /mnt (the locations udisks2/desktop file managers auto-mount
// removable media to) or any network-filesystem mount regardless of where
// it's mounted.
type Volumes struct {
	// MountsPath overrides /proc/mounts (used by tests).
	MountsPath string
	// run executes an external command and returns its combined output;
	// overridable in tests so Eject never has to shell out for real.
	run func(name string, args ...string) ([]byte, error)
}

// NewVolumes returns a Volumes ready for use against the real /proc/mounts.
func NewVolumes() *Volumes {
	return &Volumes{}
}

func (v *Volumes) mountsPath() string {
	if v.MountsPath != "" {
		return v.MountsPath
	}
	return "/proc/mounts"
}

func (v *Volumes) runner() func(string, ...string) ([]byte, error) {
	if v.run != nil {
		return v.run
	}
	return func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
}

type mountEntry struct {
	device, mountpoint, fstype string
}

// parseMounts reads /proc/mounts-format lines (device mountpoint fstype
// ...), decoding the octal escapes the kernel uses for special characters
// (spaces, tabs, ...) in device/mountpoint fields.
func parseMounts(r io.Reader) ([]mountEntry, error) {
	var out []mountEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		out = append(out, mountEntry{
			device:     unescapeMount(fields[0]),
			mountpoint: unescapeMount(fields[1]),
			fstype:     fields[2],
		})
	}
	return out, scanner.Err()
}

func unescapeMount(s string) string {
	replacer := strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)
	return replacer.Replace(s)
}

func isNetworkFSType(fstype string) bool {
	return networkFSTypes[fstype] || strings.HasPrefix(fstype, "fuse.sshfs")
}

// isRelevantMount reports whether a mount is worth surfacing: the root
// filesystem, anything under a removable-media auto-mount root, or any
// network share wherever it happens to be mounted.
func isRelevantMount(m mountEntry) bool {
	if m.mountpoint == "/" {
		return true
	}
	if pseudoFSTypes[m.fstype] {
		return false
	}
	if isNetworkFSType(m.fstype) {
		return true
	}
	return strings.HasPrefix(m.mountpoint, "/media/") || strings.HasPrefix(m.mountpoint, "/mnt/")
}

func volumeName(mountpoint string) string {
	if mountpoint == "/" {
		if h, err := os.Hostname(); err == nil && h != "" {
			return h
		}
		return "Filesystem"
	}
	return filepath.Base(mountpoint)
}

// List returns the root filesystem plus every mounted removable/network
// share. /proc/mounts can list the same mountpoint more than once (bind
// mounts, remounts) — the last entry for a given mountpoint reflects
// current state, since the kernel appends in mount order.
func (v *Volumes) List() ([]core.Volume, error) {
	f, err := os.Open(v.mountsPath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries, err := parseMounts(f)
	if err != nil {
		return nil, err
	}

	byMountpoint := make(map[string]mountEntry)
	var order []string
	for _, e := range entries {
		if !isRelevantMount(e) {
			continue
		}
		if _, seen := byMountpoint[e.mountpoint]; !seen {
			order = append(order, e.mountpoint)
		}
		byMountpoint[e.mountpoint] = e
	}

	var root *core.Volume
	rest := make([]core.Volume, 0, len(order))
	for _, mountpoint := range order {
		e := byMountpoint[mountpoint]

		var total, free uint64
		var stat unix.Statfs_t
		if err := unix.Statfs(mountpoint, &stat); err == nil {
			total = uint64(stat.Bsize) * stat.Blocks
			free = uint64(stat.Bsize) * stat.Bavail
		}

		vol := core.Volume{
			Path:       mountpoint,
			Name:       volumeName(mountpoint),
			Ejectable:  mountpoint != "/",
			TotalBytes: total,
			FreeBytes:  free,
			IsNetwork:  isNetworkFSType(e.fstype),
		}
		if mountpoint == "/" {
			root = &vol
		} else {
			rest = append(rest, vol)
		}
	}

	sort.Slice(rest, func(i, j int) bool { return rest[i].Name < rest[j].Name })

	volumes := make([]core.Volume, 0, len(rest)+1)
	if root != nil {
		volumes = append(volumes, *root)
	}
	volumes = append(volumes, rest...)
	return volumes, nil
}

// Eject unmounts path via umount, the same underlying mechanism desktop
// environments' own "Eject"/"Unmount" actions use for user-mounted media.
func (v *Volumes) Eject(path string) error {
	out, err := v.runner()("umount", path)
	if err != nil {
		return fmt.Errorf("umount %s: %w: %s", path, err, strings.TrimSpace(string(out)))
	}
	return nil
}
