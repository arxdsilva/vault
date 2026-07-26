package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// CopyInto copies src into destDir under its own name, disambiguating with
// "<name> copy<ext>" / "<name> copy N<ext>" (Finder-style) if that name is
// already taken in destDir — notably when pasting back into src's own
// directory. It returns the path created.
func CopyInto(src, destDir string) (string, error) {
	if _, err := os.Stat(src); err != nil {
		return "", err
	}

	base := filepath.Base(src)
	candidate := filepath.Join(destDir, base)

	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		if err := Copy(src, candidate); err != nil {
			return "", err
		}
		return candidate, nil
	}

	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	candidate = filepath.Join(destDir, stem+" copy"+ext)
	for i := 2; ; i++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			break
		}
		candidate = filepath.Join(destDir, fmt.Sprintf("%s copy %d%s", stem, i, ext))
	}

	if err := Copy(src, candidate); err != nil {
		return "", err
	}
	return candidate, nil
}
