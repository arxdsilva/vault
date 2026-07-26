package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// IndexEntry is the lightweight record held in the search Index — just
// enough to match and display a result, not a full Entry.
type IndexEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

// Index is an in-memory, concurrency-safe set of indexed paths used for
// search (spec §5.2). BuildIndex populates one from disk; Upsert/Remove keep
// it in sync with incremental fsnotify events afterward.
type Index struct {
	mu      sync.RWMutex
	entries map[string]IndexEntry // keyed by Path
}

// NewIndex returns an empty index.
func NewIndex() *Index {
	return &Index{entries: map[string]IndexEntry{}}
}

// BuildIndex walks root recursively and returns a populated Index. Hidden
// files/directories (dotfiles) are excluded, matching Finder's default
// search scope; unreadable subtrees are skipped rather than failing the
// whole walk.
func BuildIndex(root string) (*Index, error) {
	idx := NewIndex()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return fs.SkipDir
			}
			return err
		}
		if path == root {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		idx.Upsert(IndexEntry{Name: d.Name(), Path: path, IsDir: d.IsDir()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return idx, nil
}

// Upsert adds or replaces the entry for e.Path.
func (idx *Index) Upsert(e IndexEntry) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries[e.Path] = e
}

// Remove drops the entry for path, if present.
func (idx *Index) Remove(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.entries, path)
}

// Snapshot returns a point-in-time copy of all indexed entries.
func (idx *Index) Snapshot() []IndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	out := make([]IndexEntry, 0, len(idx.entries))
	for _, e := range idx.entries {
		out = append(out, e)
	}
	return out
}
