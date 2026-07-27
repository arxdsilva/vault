//go:build linux

package platform

import (
	"github.com/arxdsilva/vault/internal/core"
	"github.com/arxdsilva/vault/internal/platform/linux"
)

// NewVolumes returns the OS-appropriate volume lister/ejector for the
// current build target.
func NewVolumes() core.VolumeAdapter {
	return linux.NewVolumes()
}
