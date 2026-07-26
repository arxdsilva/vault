package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// Duplicate copies path into a sibling named "<name> copy<ext>", or
// "<name> copy N<ext>" if that's already taken (Finder-style), and returns
// the path created.
func Duplicate(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		return "", err
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]

	candidate := filepath.Join(dir, stem+" copy"+ext)
	for i := 2; ; i++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			break
		}
		candidate = filepath.Join(dir, fmt.Sprintf("%s copy %d%s", stem, i, ext))
	}

	if err := Copy(path, candidate); err != nil {
		return "", err
	}
	return candidate, nil
}
