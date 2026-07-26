//go:build darwin

package platform

import (
	"github.com/asilva/vault/internal/core"
	"github.com/asilva/vault/internal/platform/darwin"
)

// NewVolumes returns the OS-appropriate volume lister/ejector for the
// current build target.
func NewVolumes() core.VolumeAdapter {
	return darwin.NewVolumes()
}
