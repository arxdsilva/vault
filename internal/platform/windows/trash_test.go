//go:build windows

package windows

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrashLocation(t *testing.T) {
	t.Run("resolves and creates $Recycle.Bin under the given drive and SID", func(t *testing.T) {
		drive := t.TempDir() // stands in for a drive root in this test
		tr := &Trash{SystemDrive: drive, SID: "S-1-5-21-TEST"}

		got, err := tr.Location()
		if err != nil {
			t.Fatalf("Location: %v", err)
		}
		want := filepath.Join(drive, "$Recycle.Bin", "S-1-5-21-TEST")
		if got != want {
			t.Errorf("Location() = %q, want %q", got, want)
		}
		if info, err := os.Stat(got); err != nil || !info.IsDir() {
			t.Errorf("expected Location to create %q as a directory", got)
		}
	})

	t.Run("defaults SystemDrive from the environment", func(t *testing.T) {
		t.Setenv("SystemDrive", filepath.Join(t.TempDir()))
		tr := &Trash{SID: "S-1-5-21-TEST"}
		if got := tr.systemDrive(); got != os.Getenv("SystemDrive") {
			t.Errorf("systemDrive() = %q, want %q", got, os.Getenv("SystemDrive"))
		}
	})
}
