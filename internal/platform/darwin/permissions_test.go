//go:build darwin

package darwin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPermissionsCheck(t *testing.T) {
	t.Run("granted when the probe path is readable", func(t *testing.T) {
		probe := filepath.Join(t.TempDir(), "TCC.db")
		if err := os.WriteFile(probe, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		p := &Permissions{ProbePath: probe}
		if got := p.Check(); !got.Granted {
			t.Errorf("Check() = %+v, want Granted", got)
		}
	})

	t.Run("not granted when the probe path is unreadable", func(t *testing.T) {
		probe := filepath.Join(t.TempDir(), "TCC.db")
		if err := os.WriteFile(probe, []byte("x"), 0o000); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		t.Cleanup(func() { os.Chmod(probe, 0o644) })

		p := &Permissions{ProbePath: probe}
		got := p.Check()
		if got.Granted {
			t.Errorf("Check() = %+v, want not Granted", got)
		}
		if got.Title == "" || got.Message == "" || got.SettingsLabel == "" {
			t.Errorf("Check() = %+v, want Title/Message/SettingsLabel populated when not granted", got)
		}
	})

	t.Run("not granted when the probe path doesn't exist", func(t *testing.T) {
		p := &Permissions{ProbePath: filepath.Join(t.TempDir(), "missing")}
		if got := p.Check(); got.Granted {
			t.Errorf("Check() = %+v, want not Granted", got)
		}
	})
}

func TestPermissionsOpenSettings(t *testing.T) {
	var got string
	p := &Permissions{open: func(target string) error {
		got = target
		return nil
	}}

	if err := p.OpenSettings(); err != nil {
		t.Fatalf("OpenSettings: %v", err)
	}
	if got != settingsURL {
		t.Errorf("open target = %q, want %q", got, settingsURL)
	}
}
