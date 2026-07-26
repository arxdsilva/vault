package core

import (
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
// search (spec §5.2). Indexer.Visit populates it from disk directory by
// directory; Upsert/Remove keep it in sync with incremental fsnotify events
// afterward.
type Index struct {
	mu      sync.RWMutex
	entries map[string]IndexEntry // keyed by Path
	content map[string]string     // keyed by Path; lowercased content of small text files only
}

// NewIndex returns an empty index.
func NewIndex() *Index {
	return &Index{entries: map[string]IndexEntry{}, content: map[string]string{}}
}

// Upsert adds or replaces the entry for e.Path.
func (idx *Index) Upsert(e IndexEntry) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.entries[e.Path] = e
}

// Remove drops the entry (and any indexed content) for path, if present.
func (idx *Index) Remove(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.entries, path)
	delete(idx.content, path)
}

// Get returns the indexed entry for path, if present.
func (idx *Index) Get(path string) (IndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	e, ok := idx.entries[path]
	return e, ok
}

// UpsertContent stores lowerContent (expected already lowercased) for path,
// so MatchContentPaths can substring-match it later. Only meant for the
// small, text-classified files indexFileContent selects — see search.go's
// content-match tier.
func (idx *Index) UpsertContent(path, lowerContent string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.content[path] = lowerContent
}

// MatchContentPaths returns the paths of indexed files whose stored content
// contains lowerQuery (case-insensitive substring match; lowerQuery is
// expected already lowercased).
func (idx *Index) MatchContentPaths(lowerQuery string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var matches []string
	for path, content := range idx.content {
		if strings.Contains(content, lowerQuery) {
			matches = append(matches, path)
		}
	}
	return matches
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
