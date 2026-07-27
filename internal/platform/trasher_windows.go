//go:build windows

package platform

import (
	"github.com/arxdsilva/vault/internal/core"
	"github.com/arxdsilva/vault/internal/platform/windows"
)

// NewTrasher returns the OS-appropriate core.Trasher for the current build
// target, shared by every entry point (the GUI app, the MCP server) so the
// build-tagged OS selection lives in exactly one place.
func NewTrasher() core.Trasher {
	return windows.NewTrash()
}
