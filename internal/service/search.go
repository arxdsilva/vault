package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/asilva/vault/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// RecentEntry is one result from Search.Recent.
type RecentEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	IsDir      bool      `json:"isDir"`
	LastAccess time.Time `json:"lastAccess"`
}

// Search is the Wails service exposing fuzzy search over root (typically
// the user's home directory) to the frontend.
type Search struct {
	root         string
	frecencyPath string

	mu       sync.RWMutex
	indexer  *core.Indexer
	frecency *core.Frecency
}

// NewSearch returns a Search over root, persisting frecency data at
// frecencyPath.
func NewSearch(root, frecencyPath string) *Search {
	return &Search{root: root, frecencyPath: frecencyPath}
}

// ServiceName names this service for Wails logging/debugging.
func (s *Search) ServiceName() string { return "Search" }

// ServiceStartup loads persisted frecency data synchronously (fast: a small
// JSON file) and starts building the search index in the background so
// startup never blocks on walking the whole tree (spec §2.3, §8).
func (s *Search) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	frecency, err := core.LoadFrecency(s.frecencyPath)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.frecency = frecency
	s.mu.Unlock()

	go func() {
		indexer, err := core.StartIndexer(s.root)
		if err != nil {
			return // search just stays empty; browsing is unaffected
		}
		s.mu.Lock()
		s.indexer = indexer
		s.mu.Unlock()
	}()
	return nil
}

// ServiceShutdown stops the indexer and persists frecency data.
func (s *Search) ServiceShutdown() error {
	s.mu.Lock()
	indexer := s.indexer
	frecency := s.frecency
	s.mu.Unlock()

	if indexer != nil {
		indexer.Close()
	}
	if frecency != nil {
		return frecency.Save(s.frecencyPath)
	}
	return nil
}

// Query returns up to limit fuzzy matches for query, ranked per
// core.Search. It returns no results until the background index build has
// produced an indexer.
func (s *Search) Query(query string, limit int) []core.IndexEntry {
	s.mu.RLock()
	indexer, frecency := s.indexer, s.frecency
	s.mu.RUnlock()
	if indexer == nil || frecency == nil {
		return nil
	}

	results := core.Search(indexer.Index(), frecency, query, time.Now(), limit)
	entries := make([]core.IndexEntry, len(results))
	for i, r := range results {
		entries[i] = r.Entry
	}
	return entries
}

// Touch records that path was just accessed, boosting its future search
// ranking (spec §5.2 frecency).
func (s *Search) Touch(path string) {
	s.mu.RLock()
	frecency := s.frecency
	s.mu.RUnlock()
	if frecency == nil {
		return
	}
	frecency.Touch(path, time.Now())
}

// Recent returns up to limit most-recently-accessed paths for the Home
// screen's Recent section (spec §5.1). Paths that no longer exist on disk
// are skipped rather than returned as broken entries.
func (s *Search) Recent(limit int) []RecentEntry {
	s.mu.RLock()
	frecency := s.frecency
	s.mu.RUnlock()
	if frecency == nil {
		return nil
	}

	out := make([]RecentEntry, 0, limit)
	// Over-fetch: some touched paths may have been deleted since, so ask
	// for more than limit to still return a full page when possible.
	for _, r := range frecency.Recent(limit * 2) {
		info, err := os.Stat(r.Path)
		if err != nil {
			continue
		}
		out = append(out, RecentEntry{
			Name:       filepath.Base(r.Path),
			Path:       r.Path,
			IsDir:      info.IsDir(),
			LastAccess: r.LastAccess,
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}
