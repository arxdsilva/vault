//go:build linux

package linux

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arxdsilva/vault/internal/core"
)

func writeMounts(t *testing.T, lines string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mounts")
	if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestVolumesList(t *testing.T) {
	t.Run("lists the root filesystem and relevant mounts, skipping pseudo filesystems", func(t *testing.T) {
		usbDir := t.TempDir()
		nfsDir := t.TempDir()
		mediaMount := filepath.Join(filepath.Dir(usbDir), "USB_DRIVE")
		if err := os.Rename(usbDir, mediaMount); err != nil {
			t.Fatalf("Rename: %v", err)
		}
		netMount := filepath.Join(filepath.Dir(nfsDir), "share")
		if err := os.Rename(nfsDir, netMount); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		mounts := writeMounts(t, ""+
			"proc /proc proc rw 0 0\n"+
			"tmpfs /run tmpfs rw 0 0\n"+
			"/dev/sda1 / ext4 rw 0 0\n"+
			"/dev/sdb1 "+mediaMount+" vfat rw 0 0\n"+
			"server:/export "+netMount+" nfs4 rw 0 0\n")

		vols, err := (&Volumes{MountsPath: mounts}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(vols) != 3 {
			t.Fatalf("List returned %d volumes, want 3 (root + usb + nfs): %+v", len(vols), vols)
		}

		root := vols[0]
		if root.Path != "/" {
			t.Errorf("first volume Path = %q, want %q (root listed first)", root.Path, "/")
		}
		if root.Ejectable {
			t.Error("root Ejectable = true, want false")
		}

		var usb, nfs *core.Volume
		for i, v := range vols[1:] {
			if v.Path == mediaMount {
				usb = &vols[1:][i]
			}
			if v.Path == netMount {
				nfs = &vols[1:][i]
			}
		}
		if usb == nil {
			t.Fatal("USB mount not found in results")
		}
		if !usb.Ejectable || usb.IsNetwork {
			t.Errorf("usb = %+v, want Ejectable=true, IsNetwork=false", usb)
		}
		if usb.Name != "USB_DRIVE" {
			t.Errorf("usb.Name = %q, want %q", usb.Name, "USB_DRIVE")
		}

		if nfs == nil {
			t.Fatal("NFS mount not found in results")
		}
		if !nfs.Ejectable || !nfs.IsNetwork {
			t.Errorf("nfs = %+v, want Ejectable=true, IsNetwork=true", nfs)
		}
	})

	t.Run("ignores mounts outside /media, /mnt that aren't network shares", func(t *testing.T) {
		mounts := writeMounts(t, ""+
			"/dev/sda1 / ext4 rw 0 0\n"+
			"/dev/sda2 /home ext4 rw 0 0\n")

		vols, err := (&Volumes{MountsPath: mounts}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(vols) != 1 {
			t.Fatalf("List returned %d volumes, want 1 (root only): %+v", len(vols), vols)
		}
	})

	t.Run("last entry for a repeated mountpoint wins", func(t *testing.T) {
		dir := t.TempDir()
		mediaMount := filepath.Join(filepath.Dir(dir), "Drive")
		if err := os.Rename(dir, mediaMount); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		mounts := writeMounts(t, ""+
			"/dev/sda1 / ext4 rw 0 0\n"+
			"/dev/sdb1 "+mediaMount+" vfat rw 0 0\n"+
			"server:/x "+mediaMount+" nfs4 rw 0 0\n") // remounted as nfs

		vols, err := (&Volumes{MountsPath: mounts}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		var found bool
		for _, v := range vols {
			if v.Path == mediaMount {
				found = true
				if !v.IsNetwork {
					t.Errorf("expected the later (nfs4) mount entry to win, got IsNetwork=false")
				}
			}
		}
		if !found {
			t.Fatalf("mount %q not found: %+v", mediaMount, vols)
		}
	})
}

func TestVolumesEject(t *testing.T) {
	t.Run("invokes umount", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		v := &Volumes{run: func(name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = args
			return nil, nil
		}}

		if err := v.Eject("/media/user/USB"); err != nil {
			t.Fatalf("Eject: %v", err)
		}
		if gotName != "umount" {
			t.Errorf("command = %q, want %q", gotName, "umount")
		}
		if len(gotArgs) != 1 || gotArgs[0] != "/media/user/USB" {
			t.Errorf("args = %v, want [/media/user/USB]", gotArgs)
		}
	})

	t.Run("wraps a umount failure with its output", func(t *testing.T) {
		v := &Volumes{run: func(name string, args ...string) ([]byte, error) {
			return []byte("target is busy"), errors.New("exit status 1")
		}}

		err := v.Eject("/media/user/USB")
		if err == nil {
			t.Fatal("Eject: want an error, got nil")
		}
		if !strings.Contains(err.Error(), "target is busy") {
			t.Errorf("error = %q, want it to include umount's output", err.Error())
		}
	})
}

func TestParseMounts(t *testing.T) {
	r := strings.NewReader("/dev/sda1 /mnt/my\\040drive ext4 rw 0 0\n")
	entries, err := parseMounts(r)
	if err != nil {
		t.Fatalf("parseMounts: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("parseMounts returned %d entries, want 1", len(entries))
	}
	if entries[0].mountpoint != "/mnt/my drive" {
		t.Errorf("mountpoint = %q, want %q (octal-escaped space decoded)", entries[0].mountpoint, "/mnt/my drive")
	}
}

func TestIsNetworkFSType(t *testing.T) {
	network := []string{"nfs", "nfs4", "nfs3", "cifs", "smbfs", "smb3", "davfs", "fuse.sshfs"}
	for _, fstype := range network {
		if !isNetworkFSType(fstype) {
			t.Errorf("isNetworkFSType(%q) = false, want true", fstype)
		}
	}

	local := []string{"ext4", "xfs", "btrfs", "vfat", "ntfs", ""}
	for _, fstype := range local {
		if isNetworkFSType(fstype) {
			t.Errorf("isNetworkFSType(%q) = true, want false", fstype)
		}
	}
}
