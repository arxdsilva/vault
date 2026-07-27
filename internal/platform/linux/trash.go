//go:build linux

package linux

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Trash implements core.Trasher for Linux via the XDG Trash specification
// (freedesktop.org): trashed files move into $XDG_DATA_HOME/Trash/files
// keeping their original name (disambiguated on collision, same as
// Finder), paired with a small ".trashinfo" metadata file recording the
// original path and deletion time — the same format every XDG-compliant
// file manager (GNOME Files, Dolphin, ...) reads and writes, so Vault's
// trash interoperates with them instead of creating a second, separate one.
type Trash struct {
	// Dir overrides $XDG_DATA_HOME/Trash (used by tests); the zero value
	// resolves the real one lazily.
	Dir string
}

// NewTrash returns a Trash that resolves the real XDG trash directory on
// first use.
func NewTrash() *Trash {
	return &Trash{}
}

func (t *Trash) dir() (string, error) {
	if t.Dir != "" {
		return t.Dir, nil
	}
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "Trash"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash"), nil
}

// Location returns the trash's "files" subdirectory — the real, browsable
// folder holding trashed items under their original names — creating both
// it and its paired "info" metadata directory if they don't exist yet.
func (t *Trash) Location() (string, error) {
	dir, err := t.dir()
	if err != nil {
		return "", err
	}
	filesDir := filepath.Join(dir, "files")
	if err := os.MkdirAll(filesDir, 0o700); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(dir, "info"), 0o700); err != nil {
		return "", err
	}
	return filesDir, nil
}

// Trash moves path into the XDG trash, disambiguating a name collision
// with a numeric suffix so the trashed file and its .trashinfo stay paired
// under the same name.
func (t *Trash) Trash(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err != nil {
		return err
	}

	dir, err := t.dir()
	if err != nil {
		return err
	}
	filesDir := filepath.Join(dir, "files")
	infoDir := filepath.Join(dir, "info")
	if err := os.MkdirAll(filesDir, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(infoDir, 0o700); err != nil {
		return err
	}

	name := uniqueTrashName(filesDir, infoDir, filepath.Base(abs))
	if err := os.Rename(abs, filepath.Join(filesDir, name)); err != nil {
		return err
	}

	info := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		encodeTrashPath(abs), time.Now().Format("2006-01-02T15:04:05"))
	return os.WriteFile(filepath.Join(infoDir, name+".trashinfo"), []byte(info), 0o644)
}

// uniqueTrashName returns base, or base disambiguated with a numeric
// suffix, such that neither <filesDir>/<name> nor
// <infoDir>/<name>.trashinfo already exists.
func uniqueTrashName(filesDir, infoDir, base string) string {
	taken := func(name string) bool {
		if _, err := os.Stat(filepath.Join(filesDir, name)); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(infoDir, name+".trashinfo")); err == nil {
			return true
		}
		return false
	}
	if !taken(base) {
		return base
	}
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d%s", stem, i, ext)
		if !taken(candidate) {
			return candidate
		}
	}
}

// encodeTrashPath percent-encodes path per the XDG Trash spec (a URI path,
// RFC 2396), preserving "/" as a path separator.
func encodeTrashPath(path string) string {
	u := url.URL{Path: path}
	return u.EscapedPath()
}
