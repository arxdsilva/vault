package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type frecencyRecord struct {
	Path       string    `json:"path"`
	LastAccess time.Time `json:"lastAccess"`
	Count      int       `json:"count"`
}

// Frecency tracks how recently and how often each path has been opened, so
// search results can be ranked by recency then frequency once fuzzy match
// quality ties (spec §5.2).
type Frecency struct {
	mu      sync.RWMutex
	records map[string]*frecencyRecord
}

// NewFrecency returns an empty tracker.
func NewFrecency() *Frecency {
	return &Frecency{records: map[string]*frecencyRecord{}}
}

// Touch records that path was accessed at t.
func (f *Frecency) Touch(path string, t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.records[path]
	if !ok {
		r = &frecencyRecord{Path: path}
		f.records[path] = r
	}
	r.LastAccess = t
	r.Count++
}

// Stats returns the last access time and access count for path. An
// untouched path returns a zero time and a count of 0.
func (f *Frecency) Stats(path string) (time.Time, int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	r, ok := f.records[path]
	if !ok {
		return time.Time{}, 0
	}
	return r.LastAccess, r.Count
}

// RecentEntry is one result from Frecency.Recent.
type RecentEntry struct {
	Path       string
	LastAccess time.Time
}

// Recent returns up to limit touched paths, most recently accessed first
// (spec §5.1 Home screen).
func (f *Frecency) Recent(limit int) []RecentEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := make([]RecentEntry, 0, len(f.records))
	for _, r := range f.records {
		out = append(out, RecentEntry{Path: r.Path, LastAccess: r.LastAccess})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastAccess.After(out[j].LastAccess) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

// Save persists the tracker to path as JSON, creating parent directories as
// needed.
func (f *Frecency) Save(path string) error {
	f.mu.RLock()
	records := make([]frecencyRecord, 0, len(f.records))
	for _, r := range f.records {
		records = append(records, *r)
	}
	f.mu.RUnlock()

	b, err := json.Marshal(records)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// LoadFrecency reads a tracker previously written by Save. A missing file
// is not an error: it returns an empty, ready-to-use Frecency.
func LoadFrecency(path string) (*Frecency, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewFrecency(), nil
	}
	if err != nil {
		return nil, err
	}

	var records []frecencyRecord
	if err := json.Unmarshal(b, &records); err != nil {
		return nil, err
	}

	f := NewFrecency()
	for _, r := range records {
		rec := r
		f.records[rec.Path] = &rec
	}
	return f, nil
}
