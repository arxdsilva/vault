//go:build darwin

package platform

// ApplicationsDir returns /Applications, added as a default pin (see
// core.DefaultPins) — Finder's sidebar equivalent. Windows/Linux have no
// single canonical "installed apps" folder in the same sense, so they get
// none (see applications_other.go).
func ApplicationsDir() string {
	return "/Applications"
}
