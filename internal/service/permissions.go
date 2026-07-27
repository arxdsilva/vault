package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/arxdsilva/vault/internal/core"
)

type permissionPromptState struct {
	Dismissed bool `json:"dismissed"`
}

// Permissions is the Wails service exposing the OS-level filesystem
// permission check (macOS Full Disk Access) to the frontend, so Vault can
// guide the user to grant it once at startup instead of a separate system
// prompt for every protected folder (Desktop, Documents, Downloads, …) as
// it's first visited.
//
// Whether the startup prompt has been dismissed is persisted to path, so
// declining it once doesn't nag on every subsequent launch while the
// permission stays ungranted. This is independent of Status, which always
// reflects the real OS grant — a folder that fails to open still surfaces
// permission guidance inline regardless of this flag.
type Permissions struct {
	adapter core.PermissionAdapter
	path    string

	mu        sync.Mutex
	dismissed bool
}

// NewPermissions returns a Permissions service backed by adapter for the
// current OS, persisting startup-prompt dismissal to path.
func NewPermissions(adapter core.PermissionAdapter, path string) *Permissions {
	return &Permissions{
		adapter:   adapter,
		path:      path,
		dismissed: loadPromptDismissed(path),
	}
}

// ServiceName names this service for Wails logging/debugging.
func (p *Permissions) ServiceName() string { return "Permissions" }

// Status reports whether the OS has granted Vault broad filesystem access.
func (p *Permissions) Status() core.PermissionStatus {
	return p.adapter.Check()
}

// PromptDismissed reports whether the user has already dismissed the
// startup permission prompt, so it isn't shown again every launch.
func (p *Permissions) PromptDismissed() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dismissed
}

// DismissPrompt persists that the user has handled the startup permission
// prompt (via "Not now" or after being sent to Settings).
func (p *Permissions) DismissPrompt() error {
	p.mu.Lock()
	p.dismissed = true
	p.mu.Unlock()

	b, err := json.Marshal(permissionPromptState{Dismissed: true})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p.path, b, 0o644)
}

// OpenSettings opens the OS's settings surface for granting the permission.
func (p *Permissions) OpenSettings() error {
	return p.adapter.OpenSettings()
}

func loadPromptDismissed(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var state permissionPromptState
	if err := json.Unmarshal(b, &state); err != nil {
		return false
	}
	return state.Dismissed
}
