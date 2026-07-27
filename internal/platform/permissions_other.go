//go:build !darwin && !windows

package platform

import "github.com/asilva/vault/internal/core"

// grantedAdapter reports the permission as always granted: Linux (and any
// other non-macOS, non-Windows target) has no per-folder consent system or
// opt-in write-blocking feature to guard against.
type grantedAdapter struct{}

func (grantedAdapter) Check() core.PermissionStatus { return core.PermissionStatus{Granted: true} }
func (grantedAdapter) OpenSettings() error          { return nil }

// NewPermissions returns the OS-appropriate permission adapter for the
// current build target.
func NewPermissions() core.PermissionAdapter {
	return grantedAdapter{}
}
