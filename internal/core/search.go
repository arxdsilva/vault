package core

import (
	"sort"
	"strings"
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
	seen := make(map[string]bool, len(matches))
	candidates := make([]ranked, len(matches))
	for i, m := range matches {
		e := entries[m.Index]
		last, count := frecency.Stats(e.Path)
		candidates[i] = ranked{entry: e, matchScore: m.Score, lastAccess: last, count: count}
		seen[e.Path] = true
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

	// Files whose content contains query but whose name didn't already
	// fuzzy-match: a broader-recall tier appended after name matches, since
	// name relevance is the stronger signal (spec §5.2). Substring (not
	// fuzzy) matching is the right semantic for file bodies — fuzzy
	// subsequence matching over whole file contents would match almost
	// anything and isn't useful here.
	var contentCandidates []ranked
	for _, path := range idx.MatchContentPaths(strings.ToLower(query)) {
		if seen[path] {
			continue
		}
		e, ok := idx.Get(path)
		if !ok {
			continue
		}
		last, count := frecency.Stats(path)
		contentCandidates = append(contentCandidates, ranked{entry: e, lastAccess: last, count: count})
		seen[path] = true
	}
	sort.SliceStable(contentCandidates, func(i, j int) bool {
		a, b := contentCandidates[i], contentCandidates[j]
		if !a.lastAccess.Equal(b.lastAccess) {
			return a.lastAccess.After(b.lastAccess)
		}
		return a.count > b.count
	})
	candidates = append(candidates, contentCandidates...)

	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}

	results := make([]SearchResult, len(candidates))
	for i, c := range candidates {
		results[i] = SearchResult{Entry: c.entry}
	}
	return results
}
