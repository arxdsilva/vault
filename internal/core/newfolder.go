package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewFolder creates a new directory named name inside parent, disambiguating
// with a numeric suffix ("name 2", "name 3", ...) if it's already taken. It
// returns the path actually created.
func NewFolder(parent, name string) (string, error) {
	path := filepath.Join(parent, name)
	for i := 2; ; i++ {
		err := os.Mkdir(path, 0o755)
		if err == nil {
			return path, nil
		}
		if !os.IsExist(err) {
			return "", err
		}
		path = filepath.Join(parent, fmt.Sprintf("%s %d", name, i))
	}
}
