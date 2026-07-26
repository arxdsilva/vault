package core

import (
	"os"
	"path/filepath"
)

// Stat returns metadata for a single path (unlike ListDir, which lists a
// directory's children) — used to describe a path itself, e.g. a pinned
// folder's own size/mod time for a hover card.
func Stat(path string) (Entry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Entry{}, err
	}
	return Entry{
		Name:    filepath.Base(path),
		Path:    path,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}
