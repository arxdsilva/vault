//go:build windows

package platform

import (
	"github.com/asilva/vault/internal/core"
	"github.com/asilva/vault/internal/platform/windows"
)

// NewPermissions returns the OS-appropriate permission adapter for the
// current build target.
func NewPermissions() core.PermissionAdapter {
	return windows.NewPermissions()
}
