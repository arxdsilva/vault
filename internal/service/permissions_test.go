package service

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/arxdsilva/vault/internal/core"
)

type fakePermissionAdapter struct {
	status       core.PermissionStatus
	openErr      error
	openedCalled bool
}

func (f *fakePermissionAdapter) Check() core.PermissionStatus { return f.status }

func (f *fakePermissionAdapter) OpenSettings() error {
	f.openedCalled = true
	return f.openErr
}

func TestPermissionsStatus(t *testing.T) {
	adapter := &fakePermissionAdapter{status: core.PermissionStatus{Granted: true}}
	svc := NewPermissions(adapter, filepath.Join(t.TempDir(), "permission-prompt.json"))

	if got := svc.Status(); !got.Granted {
		t.Errorf("Status() = %+v, want Granted", got)
	}
}

func TestPermissionsOpenSettings(t *testing.T) {
	adapter := &fakePermissionAdapter{}
	svc := NewPermissions(adapter, filepath.Join(t.TempDir(), "permission-prompt.json"))

	if err := svc.OpenSettings(); err != nil {
		t.Fatalf("OpenSettings: %v", err)
	}
	if !adapter.openedCalled {
		t.Error("OpenSettings did not call the adapter")
	}
}

func TestPermissionsOpenSettingsError(t *testing.T) {
	adapter := &fakePermissionAdapter{openErr: errors.New("boom")}
	svc := NewPermissions(adapter, filepath.Join(t.TempDir(), "permission-prompt.json"))

	if err := svc.OpenSettings(); err == nil {
		t.Fatal("OpenSettings: want an error, got nil")
	}
}

func TestPermissionsPromptDismissedDefaultsFalse(t *testing.T) {
	adapter := &fakePermissionAdapter{}
	svc := NewPermissions(adapter, filepath.Join(t.TempDir(), "permission-prompt.json"))

	if svc.PromptDismissed() {
		t.Error("PromptDismissed() = true, want false before DismissPrompt is ever called")
	}
}

func TestPermissionsDismissPromptPersists(t *testing.T) {
	adapter := &fakePermissionAdapter{}
	path := filepath.Join(t.TempDir(), "vault", "permission-prompt.json")
	svc := NewPermissions(adapter, path)

	if err := svc.DismissPrompt(); err != nil {
		t.Fatalf("DismissPrompt: %v", err)
	}
	if !svc.PromptDismissed() {
		t.Error("PromptDismissed() = false right after DismissPrompt, want true")
	}

	// A fresh service instance backed by the same path (simulating the next
	// app launch) must see the dismissal too.
	relaunched := NewPermissions(adapter, path)
	if !relaunched.PromptDismissed() {
		t.Error("PromptDismissed() on a fresh instance = false, want true (should persist across launches)")
	}
}

func TestPermissionsPromptDismissedMissingFile(t *testing.T) {
	adapter := &fakePermissionAdapter{}
	svc := NewPermissions(adapter, filepath.Join(t.TempDir(), "missing", "permission-prompt.json"))

	if svc.PromptDismissed() {
		t.Error("PromptDismissed() = true for a never-written file, want false")
	}
}
