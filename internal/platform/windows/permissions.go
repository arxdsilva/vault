//go:build windows

package windows

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asilva/vault/internal/core"
)

// controlledFolderAccessURL opens Windows Security directly to the
// Ransomware protection page, where Controlled Folder Access is toggled
// and apps are allow-listed.
const controlledFolderAccessURL = "windowsdefender://ransomwareprotection"

// Permissions checks and requests the closest Windows analogue to macOS's
// per-folder consent: Controlled Folder Access, an opt-in Windows Defender
// ransomware-protection feature that — only when a user has turned it on —
// silently blocks unlisted apps from writing to Desktop, Documents,
// Pictures, and other protected folders. Off by default for most Windows
// installs, unlike macOS's TCC which always applies.
type Permissions struct {
	// ExecutablePath overrides the running executable's path matched
	// against the allow-list (used by tests). The zero value resolves
	// os.Executable lazily.
	ExecutablePath string
	// run executes an external command and returns its combined output;
	// overridable in tests so Check never has to shell out to PowerShell
	// for real.
	run func(name string, args ...string) ([]byte, error)
	// open launches a URL/path via rundll32; overridable in tests.
	open func(target string) error
}

// NewPermissions returns a Permissions ready for use against the real
// Controlled Folder Access configuration.
func NewPermissions() *Permissions {
	return &Permissions{}
}

func (p *Permissions) runner() func(string, ...string) ([]byte, error) {
	if p.run != nil {
		return p.run
	}
	return func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
}

func (p *Permissions) executablePath() (string, error) {
	if p.ExecutablePath != "" {
		return p.ExecutablePath, nil
	}
	return os.Executable()
}

// Check reports whether Controlled Folder Access — if the user has turned
// it on — would block Vault. Any failure to query Defender's preferences
// (PowerShell unavailable, Defender disabled, non-standard Windows edition)
// is treated as granted, since a feature that can't be queried can't be
// actively blocking anything either.
func (p *Permissions) Check() core.PermissionStatus {
	blocking, err := p.controlledFolderAccessBlocking()
	if err != nil || !blocking {
		return core.PermissionStatus{Granted: true}
	}

	exe, err := p.executablePath()
	if err == nil {
		if allowed, err := p.allowedApplications(); err == nil {
			for _, a := range allowed {
				if strings.EqualFold(filepath.Clean(a), filepath.Clean(exe)) {
					return core.PermissionStatus{Granted: true}
				}
			}
		}
	}

	return core.PermissionStatus{
		Granted: false,
		Title:   "Allow Vault through Controlled folder access",
		Message: "Windows' ransomware protection is blocking Vault from writing to Desktop, Documents, " +
			"Pictures, and other protected folders.\n\n" +
			"Add Vault to the allowed apps list, then quit and reopen it.",
		SettingsLabel: "Open Windows Security",
	}
}

// controlledFolderAccessBlocking reports whether Controlled Folder Access
// is actively blocking writes. Defender's EnableControlledFolderAccess is
// 1 for "Enabled (Block)"; 0 (Disabled) and 2 (Audit Mode, log-only) don't
// block anything.
func (p *Permissions) controlledFolderAccessBlocking() (bool, error) {
	out, err := p.runner()("powershell", "-NoProfile", "-Command", "(Get-MpPreference).EnableControlledFolderAccess")
	if err != nil {
		return false, err
	}
	mode, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return false, err
	}
	return mode == 1, nil
}

// allowedApplications lists the full paths Controlled Folder Access
// currently allow-lists.
func (p *Permissions) allowedApplications() ([]string, error) {
	out, err := p.runner()("powershell", "-NoProfile", "-Command",
		"(Get-MpPreference).ControlledFolderAccessAllowedApplications")
	if err != nil {
		return nil, err
	}
	var apps []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			apps = append(apps, line)
		}
	}
	return apps, nil
}

func (p *Permissions) opener() func(string) error {
	if p.open != nil {
		return p.open
	}
	return func(target string) error {
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Run()
	}
}

// OpenSettings opens Windows Security directly to the Ransomware protection
// page, where the user can allow-list Vault under Controlled folder access.
func (p *Permissions) OpenSettings() error {
	return p.opener()(controlledFolderAccessURL)
}
