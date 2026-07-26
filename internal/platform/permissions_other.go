//go:build !darwin

package platform

import "github.com/asilva/vault/internal/core"

// grantedAdapter reports Full Disk Access-equivalent as always granted:
// only macOS gates protected folders (Desktop, Documents, Downloads, …)
// behind a consent prompt per location.
type grantedAdapter struct{}

func (grantedAdapter) Check() core.PermissionStatus { return core.PermissionStatus{Granted: true} }
func (grantedAdapter) OpenSettings() error          { return nil }

// NewPermissions returns the OS-appropriate permission adapter for the
// current build target.
func NewPermissions() core.PermissionAdapter {
	return grantedAdapter{}
}
