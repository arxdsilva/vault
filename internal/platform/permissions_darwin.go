//go:build darwin

package platform

import (
	"github.com/arxdsilva/vault/internal/core"
	"github.com/arxdsilva/vault/internal/platform/darwin"
)

// NewPermissions returns the OS-appropriate permission adapter for the
// current build target.
func NewPermissions() core.PermissionAdapter {
	return darwin.NewPermissions()
}
