package core

// PermissionStatus reports whether the OS has granted Vault broad
// filesystem access.
type PermissionStatus struct {
	Granted bool `json:"granted"`
}

// PermissionAdapter checks and requests the OS-level permission needed to
// browse protected locations (macOS Full Disk Access and equivalents)
// without a separate consent prompt per folder. Each OS provides its own
// implementation under internal/platform/<os>; OSes with no such concept
// (Windows, Linux) report Granted: true unconditionally.
type PermissionAdapter interface {
	// Check reports the current grant status.
	Check() PermissionStatus
	// OpenSettings opens the OS's settings surface for granting the
	// permission (a no-op where the OS has no such surface).
	OpenSettings() error
}
