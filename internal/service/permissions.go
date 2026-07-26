package service

import "github.com/asilva/vault/internal/core"

// Permissions is the Wails service exposing the OS-level filesystem
// permission check (macOS Full Disk Access) to the frontend, so Vault can
// guide the user to grant it once at startup instead of a separate system
// prompt for every protected folder (Desktop, Documents, Downloads, …) as
// it's first visited.
type Permissions struct {
	adapter core.PermissionAdapter
}

// NewPermissions returns a Permissions service backed by adapter for the
// current OS.
func NewPermissions(adapter core.PermissionAdapter) *Permissions {
	return &Permissions{adapter: adapter}
}

// ServiceName names this service for Wails logging/debugging.
func (p *Permissions) ServiceName() string { return "Permissions" }

// Status reports whether the OS has granted Vault broad filesystem access.
func (p *Permissions) Status() core.PermissionStatus {
	return p.adapter.Check()
}

// OpenSettings opens the OS's settings surface for granting the permission.
func (p *Permissions) OpenSettings() error {
	return p.adapter.OpenSettings()
}
