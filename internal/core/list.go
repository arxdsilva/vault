package core

import (
	"os"
	"path/filepath"
)

// ListDir returns the entries directly inside dir (non-recursive).
func ListDir(dir string) ([]Entry, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(des))
	for _, de := range des {
		info, err := de.Info()
		if err != nil {
			// Entry vanished between ReadDir and Info (e.g. removed by another
			// process); skip it rather than failing the whole listing.
			continue
		}
		entries = append(entries, Entry{
			Name:    de.Name(),
			Path:    filepath.Join(dir, de.Name()),
			IsDir:   de.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return entries, nil
}
