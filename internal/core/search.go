package core

import (
	"sort"
	"time"

	"github.com/sahilm/fuzzy"
)

// SearchResult is one ranked search hit.
type SearchResult struct {
	Entry IndexEntry
}

// Search fuzzy-matches query against idx, ranked by match quality, then
// recency of access, then frequency of access (spec §5.2). An empty query
// returns no results — the caller falls back to the Home screen instead.
func Search(idx *Index, frecency *Frecency, query string, now time.Time, limit int) []SearchResult {
	if query == "" {
		return nil
	}

	entries := idx.Snapshot()
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}

	matches := fuzzy.Find(query, names)
	type ranked struct {
		entry      IndexEntry
		matchScore int
		lastAccess time.Time
		count      int
	}
	candidates := make([]ranked, len(matches))
	for i, m := range matches {
		e := entries[m.Index]
		last, count := frecency.Stats(e.Path)
		candidates[i] = ranked{entry: e, matchScore: m.Score, lastAccess: last, count: count}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.matchScore != b.matchScore {
			return a.matchScore > b.matchScore
		}
		if !a.lastAccess.Equal(b.lastAccess) {
			return a.lastAccess.After(b.lastAccess)
		}
		return a.count > b.count
	})

	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}

	results := make([]SearchResult, len(candidates))
	for i, c := range candidates {
		results[i] = SearchResult{Entry: c.entry}
	}
	return results
}
