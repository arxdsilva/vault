package service

import (
	"errors"
	"testing"

	"github.com/asilva/vault/internal/core"
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
	svc := NewPermissions(adapter)

	if got := svc.Status(); !got.Granted {
		t.Errorf("Status() = %+v, want Granted", got)
	}
}

func TestPermissionsOpenSettings(t *testing.T) {
	adapter := &fakePermissionAdapter{}
	svc := NewPermissions(adapter)

	if err := svc.OpenSettings(); err != nil {
		t.Fatalf("OpenSettings: %v", err)
	}
	if !adapter.openedCalled {
		t.Error("OpenSettings did not call the adapter")
	}
}

func TestPermissionsOpenSettingsError(t *testing.T) {
	adapter := &fakePermissionAdapter{openErr: errors.New("boom")}
	svc := NewPermissions(adapter)

	if err := svc.OpenSettings(); err == nil {
		t.Fatal("OpenSettings: want an error, got nil")
	}
}
