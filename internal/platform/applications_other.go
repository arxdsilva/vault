//go:build !darwin

package platform

// ApplicationsDir reports no default Applications pin: Windows/Linux have
// no single canonical "installed apps" folder the way macOS's /Applications
// is one (see applications_darwin.go).
func ApplicationsDir() string {
	return ""
}
