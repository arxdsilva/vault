//go:build windows

package windows

import (
	"errors"
	"strings"
	"testing"
)

func TestVolumesEject(t *testing.T) {
	t.Run("invokes the Shell.Application eject verb via PowerShell", func(t *testing.T) {
		var gotName string
		var gotArgs []string
		v := &Volumes{run: func(name string, args ...string) ([]byte, error) {
			gotName = name
			gotArgs = args
			return nil, nil
		}}

		if err := v.Eject(`D:\`); err != nil {
			t.Fatalf("Eject: %v", err)
		}
		if gotName != "powershell" {
			t.Errorf("command = %q, want %q", gotName, "powershell")
		}
		script := gotArgs[len(gotArgs)-1]
		if !strings.Contains(script, "ParseName('D:')") {
			t.Errorf("script = %q, want it to reference the drive letter", script)
		}
	})

	t.Run("wraps a PowerShell failure with its output", func(t *testing.T) {
		v := &Volumes{run: func(name string, args ...string) ([]byte, error) {
			return []byte("Access is denied"), errors.New("exit status 1")
		}}

		err := v.Eject(`D:\`)
		if err == nil {
			t.Fatal("Eject: want an error, got nil")
		}
		if !strings.Contains(err.Error(), "Access is denied") {
			t.Errorf("error = %q, want it to include the command's output", err.Error())
		}
	})
}

func TestVolumesSystemDrive(t *testing.T) {
	v := &Volumes{SystemDrive: "d:"}
	if got := v.systemDrive(); got != "D:" {
		t.Errorf("systemDrive() = %q, want %q", got, "D:")
	}
}
