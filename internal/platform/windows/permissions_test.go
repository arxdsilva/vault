//go:build windows

package windows

import (
	"errors"
	"strings"
	"testing"
)

func fakeMpPreference(mode string, allowed string) func(name string, args ...string) ([]byte, error) {
	return func(name string, args ...string) ([]byte, error) {
		script := args[len(args)-1]
		if strings.Contains(script, "EnableControlledFolderAccess") {
			return []byte(mode), nil
		}
		if strings.Contains(script, "ControlledFolderAccessAllowedApplications") {
			return []byte(allowed), nil
		}
		return nil, errors.New("unexpected script: " + script)
	}
}

func TestPermissionsCheck(t *testing.T) {
	t.Run("granted when Controlled Folder Access is off", func(t *testing.T) {
		p := &Permissions{run: fakeMpPreference("0", "")}
		if got := p.Check(); !got.Granted {
			t.Errorf("Check() = %+v, want Granted", got)
		}
	})

	t.Run("granted when Controlled Folder Access is in audit mode", func(t *testing.T) {
		p := &Permissions{run: fakeMpPreference("2", "")}
		if got := p.Check(); !got.Granted {
			t.Errorf("Check() = %+v, want Granted", got)
		}
	})

	t.Run("granted when blocking is on but Vault is already allow-listed", func(t *testing.T) {
		p := &Permissions{
			ExecutablePath: `C:\Program Files\Vault\vault.exe`,
			run:            fakeMpPreference("1", `C:\Program Files\Vault\vault.exe`+"\n"),
		}
		if got := p.Check(); !got.Granted {
			t.Errorf("Check() = %+v, want Granted", got)
		}
	})

	t.Run("not granted when blocking is on and Vault isn't allow-listed", func(t *testing.T) {
		p := &Permissions{
			ExecutablePath: `C:\Program Files\Vault\vault.exe`,
			run:            fakeMpPreference("1", `C:\Program Files\OtherApp\other.exe`+"\n"),
		}
		got := p.Check()
		if got.Granted {
			t.Errorf("Check() = %+v, want not Granted", got)
		}
		if got.Title == "" || got.Message == "" || got.SettingsLabel == "" {
			t.Errorf("Check() = %+v, want Title/Message/SettingsLabel populated when not granted", got)
		}
	})

	t.Run("granted when PowerShell is unavailable", func(t *testing.T) {
		p := &Permissions{run: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("powershell: command not found")
		}}
		if got := p.Check(); !got.Granted {
			t.Errorf("Check() = %+v, want Granted (fail open)", got)
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
	if got != controlledFolderAccessURL {
		t.Errorf("open target = %q, want %q", got, controlledFolderAccessURL)
	}
}
