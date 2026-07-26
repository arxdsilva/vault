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
	t.Run("lists real mount directories, skipping the boot volume symlink and dotfiles", func(t *testing.T) {
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
		if len(vols) != 1 {
			t.Fatalf("List returned %d volumes, want 1: %+v", len(vols), vols)
		}
		if vols[0].Name != "USB Drive" {
			t.Errorf("Name = %q, want %q", vols[0].Name, "USB Drive")
		}
		if vols[0].Path != filepath.Join(root, "USB Drive") {
			t.Errorf("Path = %q, want %q", vols[0].Path, filepath.Join(root, "USB Drive"))
		}
		if !vols[0].Ejectable {
			t.Error("Ejectable = false, want true")
		}
		if vols[0].TotalBytes == 0 {
			t.Error("TotalBytes = 0, want a real filesystem size")
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
