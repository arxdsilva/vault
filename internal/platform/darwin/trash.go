//go:build darwin

// Package darwin implements the macOS platform adapter: trash semantics and
// any other OS-specific behavior that must stay out of internal/core.
package darwin

import (
	"fmt"
	"os"
	"path/filepath"
)

// Trash implements core.Trasher for macOS by moving files into the user's
// ~/.Trash, the same location Finder empties from. Dir overrides that
// location (used by tests); the zero value resolves ~/.Trash lazily.
type Trash struct {
	Dir string
}

// NewTrash returns a Trash that resolves the real ~/.Trash on first use.
func NewTrash() *Trash {
	return &Trash{}
}

func (t *Trash) dir() (string, error) {
	if t.Dir != "" {
		return t.Dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".Trash"), nil
}

// Location returns ~/.Trash, creating it if it doesn't exist yet — a real,
// browsable folder (trashed files keep their original names), the same one
// Finder's Trash icon shows.
func (t *Trash) Location() (string, error) {
	dir, err := t.dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// Trash moves path into the trash directory, disambiguating a name
// collision with a numeric suffix the way Finder does.
func (t *Trash) Trash(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	dir, err := t.dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	dest := uniquePath(filepath.Join(dir, filepath.Base(path)))
	return os.Rename(path, dest)
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
