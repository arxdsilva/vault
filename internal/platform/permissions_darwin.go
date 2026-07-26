//go:build darwin

package platform

import (
	"github.com/asilva/vault/internal/core"
	"github.com/asilva/vault/internal/platform/darwin"
)

// NewPermissions returns the OS-appropriate permission adapter for the
// current build target.
func NewPermissions() core.PermissionAdapter {
	return darwin.NewPermissions()
}
