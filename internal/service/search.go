package service

import (
	"context"
	"log"
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

// ServiceStartup loads persisted frecency data and starts the indexer, both
// synchronously — neither touches the disk tree (spec §2.3, §8): the index
// starts empty and only grows as Visit is called for folders the user
// actually browses to (spec §5.2 follow-up). Eagerly walking the whole home
// directory up front used to leave search dead for minutes on a real
// machine (huge folders like ~/Library, macOS TCC permission errors
// aborting the walk entirely).
func (s *Search) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	frecency, err := core.LoadFrecency(s.frecencyPath)
	if err != nil {
		return err
	}
	indexer, err := core.StartIndexer(s.root)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.frecency = frecency
	s.indexer = indexer
	s.mu.Unlock()
	return nil
}

// Visit indexes dir's immediate contents and keeps it live via a watch, so
// search covers it from now on. The frontend calls this whenever the user
// navigates to a folder; the MCP server's list_directory tool calls it too,
// so an agent's search tool covers whatever it has already listed. A no-op
// if the indexer isn't up yet (there's nothing useful to report back, and
// the frontend doesn't wait on this call).
func (s *Search) Visit(dir string) error {
	s.mu.RLock()
	indexer := s.indexer
	s.mu.RUnlock()
	if indexer == nil {
		return nil
	}
	return indexer.Visit(dir)
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
// core.Search, over whatever folders have been indexed via Visit so far. If
// dir is non-empty, matches under dir (typically the folder currently being
// browsed) are guaranteed their own share of limit rather than competing
// with the rest of the index for a single shared cutoff — see core.Search's
// doc comment for why that matters.
func (s *Search) Query(query string, limit int, dir string) []core.IndexEntry {
	s.mu.RLock()
	indexer, frecency := s.indexer, s.frecency
	s.mu.RUnlock()
	if indexer == nil || frecency == nil {
		log.Printf("search: query %q dir %q dropped — service not started yet", query, dir)
		return nil
	}

	results := core.Search(indexer.Index(), frecency, query, time.Now(), limit, dir)
	entries := make([]core.IndexEntry, len(results))
	for i, r := range results {
		entries[i] = r.Entry
	}
	log.Printf("search: query %q dir %q limit %d -> %d results (index size %d)",
		query, dir, limit, len(entries), len(indexer.Index().Snapshot()))
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
