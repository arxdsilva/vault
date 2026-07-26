//go:build darwin

package darwin

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumesList(t *testing.T) {
	t.Run("lists real mount directories and the boot volume, skipping dotfiles", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "USB Drive"), 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.Mkdir(filepath.Join(root, ".hidden"), 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.Symlink("/", filepath.Join(root, "Macintosh HD")); err != nil {
			t.Fatalf("Symlink: %v", err)
		}

		vols, err := (&Volumes{Root: root}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(vols) != 2 {
			t.Fatalf("List returned %d volumes, want 2: %+v", len(vols), vols)
		}

		// Sorted by name: "Macintosh HD" before "USB Drive".
		boot, usb := vols[0], vols[1]

		if boot.Name != "Macintosh HD" {
			t.Errorf("Name = %q, want %q", boot.Name, "Macintosh HD")
		}
		if boot.Path != filepath.Join(root, "Macintosh HD") {
			t.Errorf("Path = %q, want %q", boot.Path, filepath.Join(root, "Macintosh HD"))
		}
		if boot.Ejectable {
			t.Error("boot volume Ejectable = true, want false")
		}
		if boot.TotalBytes == 0 {
			t.Error("boot volume TotalBytes = 0, want a real filesystem size")
		}

		if usb.Name != "USB Drive" {
			t.Errorf("Name = %q, want %q", usb.Name, "USB Drive")
		}
		if usb.Path != filepath.Join(root, "USB Drive") {
			t.Errorf("Path = %q, want %q", usb.Path, filepath.Join(root, "USB Drive"))
		}
		if !usb.Ejectable {
			t.Error("USB Drive Ejectable = false, want true")
		}
		if usb.TotalBytes == 0 {
			t.Error("USB Drive TotalBytes = 0, want a real filesystem size")
		}
	})

	t.Run("skips a symlink that isn't the boot volume", func(t *testing.T) {
		root := t.TempDir()
		other := t.TempDir()
		if err := os.Symlink(other, filepath.Join(root, "Not Boot")); err != nil {
			t.Fatalf("Symlink: %v", err)
		}

		vols, err := (&Volumes{Root: root}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(vols) != 0 {
			t.Fatalf("List returned %d volumes, want 0: %+v", len(vols), vols)
		}
	})

	t.Run("returns an empty slice, not an error, when nothing is mounted", func(t *testing.T) {
		root := t.TempDir()

		vols, err := (&Volumes{Root: root}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(vols) != 0 {
			t.Errorf("List returned %d volumes, want 0", len(vols))
		}
	})

	t.Run("sorts volumes by name", func(t *testing.T) {
		root := t.TempDir()
		for _, name := range []string{"Zeta", "Alpha", "Mu"} {
			if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
				t.Fatalf("Mkdir: %v", err)
			}
		}

		vols, err := (&Volumes{Root: root}).List()
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		want := []string{"Alpha", "Mu", "Zeta"}
		for i, name := range want {
			if vols[i].Name != name {
				t.Errorf("vols[%d].Name = %q, want %q", i, vols[i].Name, name)
			}
		}
	})

	t.Run("propagates an error when the root directory doesn't exist", func(t *testing.T) {
		_, err := (&Volumes{Root: filepath.Join(t.TempDir(), "missing")}).List()
		if err == nil {
			t.Fatal("List: want an error for a missing root, got nil")
		}
	})
}

func TestVolumesEject(t *testing.T) {
	t.Run("shells out to diskutil eject with the volume path", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		v := &Volumes{run: func(name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = args
			return []byte("Unmount successful"), nil
		}}

		if err := v.Eject("/Volumes/USB Drive"); err != nil {
			t.Fatalf("Eject: %v", err)
		}
		if gotName != "diskutil" {
			t.Errorf("command = %q, want %q", gotName, "diskutil")
		}
		wantArgs := []string{"eject", "/Volumes/USB Drive"}
		if len(gotArgs) != len(wantArgs) || gotArgs[0] != wantArgs[0] || gotArgs[1] != wantArgs[1] {
			t.Errorf("args = %v, want %v", gotArgs, wantArgs)
		}
	})

	t.Run("wraps a diskutil failure with its output", func(t *testing.T) {
		v := &Volumes{run: func(name string, args ...string) ([]byte, error) {
			return []byte("Resource busy"), errors.New("exit status 1")
		}}

		err := v.Eject("/Volumes/USB Drive")
		if err == nil {
			t.Fatal("Eject: want an error, got nil")
		}
		if got := err.Error(); !strings.Contains(got, "Resource busy") {
			t.Errorf("error = %q, want it to include diskutil's output", got)
		}
	})
}
