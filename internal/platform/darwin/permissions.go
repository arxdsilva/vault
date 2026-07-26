//go:build darwin

package darwin

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/asilva/vault/internal/core"
)

// settingsURL opens macOS's Full Disk Access pane directly, so the user
// doesn't have to hunt through System Settings > Privacy & Security.
const settingsURL = "x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles"

// Permissions checks and requests macOS Full Disk Access. Granting it lets
// Vault browse Desktop, Documents, Downloads, and removable/network volumes
// without a separate TCC consent prompt the first time each is visited.
type Permissions struct {
	// ProbePath overrides the file whose readability indicates Full Disk
	// Access (used by tests). The zero value resolves the real per-user TCC
	// database lazily.
	ProbePath string
	// open launches a URL/path via the "open" command; overridable in tests.
	open func(target string) error
}

// NewPermissions returns a Permissions ready for use against the real TCC
// database.
func NewPermissions() *Permissions {
	return &Permissions{}
}

func (p *Permissions) probePath() (string, error) {
	if p.ProbePath != "" {
		return p.ProbePath, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Reading the per-user TCC database itself is gated behind Full Disk
	// Access — the same probe used elsewhere to detect it, since Apple
	// exposes no public API for this check.
	return filepath.Join(home, "Library", "Application Support", "com.apple.TCC", "TCC.db"), nil
}

// Check reports whether Full Disk Access is currently granted.
func (p *Permissions) Check() core.PermissionStatus {
	path, err := p.probePath()
	if err != nil {
		return core.PermissionStatus{Granted: false}
	}
	f, err := os.Open(path)
	if err != nil {
		return core.PermissionStatus{Granted: false}
	}
	f.Close()
	return core.PermissionStatus{Granted: true}
}

func (p *Permissions) opener() func(string) error {
	if p.open != nil {
		return p.open
	}
	return func(target string) error {
		return exec.Command("open", target).Run()
	}
}

// OpenSettings opens System Settings directly to the Full Disk Access pane.
func (p *Permissions) OpenSettings() error {
	return p.opener()(settingsURL)
}
