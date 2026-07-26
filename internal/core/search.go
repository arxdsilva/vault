package core

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
)

// SearchResult is one ranked search hit.
type SearchResult struct {
	Entry IndexEntry
}

type ranked struct {
	entry      IndexEntry
	matchScore int
	lastAccess time.Time
	count      int
}

// Search fuzzy/content-matches query against idx, ranked by match quality,
// then recency of access, then frequency of access (frecency) (spec §5.2).
// An empty query returns no results — the caller falls back to the Home
// screen instead.
//
// If scopeDir is non-empty, matches under scopeDir get their own
// allocation of up to limit results, independent of how many matches exist
// elsewhere in the tree. Without this, a single global rank-then-truncate
// pass can silently drop a file the user is looking for: sahilm/fuzzy
// scores shorter names higher for the same prefix match, so a folder full
// of similarly-prefixed short names (e.g. "Screenshot 1.png" .. "Screenshot
// 200.png") can systematically crowd a longer, more specific match (e.g.
// "Screenshot final report.png") out of a shared top-N cutoff — every time,
// not just occasionally. Scoping guarantees the folder you're actually
// browsing isn't competing with the rest of the tree for its share of the
// limit.
func Search(idx *Index, frecency *Frecency, query string, now time.Time, limit int, scopeDir string) []SearchResult {
	if query == "" {
		return nil
	}

	all := rankCandidates(idx, frecency, query, now)

	final := all
	if scopeDir != "" {
		var inScope, outScope []ranked
		for _, c := range all {
			if isUnderDir(c.entry.Path, scopeDir) {
				inScope = append(inScope, c)
			} else {
				outScope = append(outScope, c)
			}
		}
		if limit > 0 && len(inScope) > limit {
			inScope = inScope[:limit]
		}
		final = inScope
		if limit <= 0 {
			final = append(final, outScope...)
		} else if remaining := limit - len(inScope); remaining > 0 {
			if remaining > len(outScope) {
				remaining = len(outScope)
			}
			final = append(final, outScope[:remaining]...)
		}
	}

	if limit > 0 && len(final) > limit {
		final = final[:limit]
	}

	results := make([]SearchResult, len(final))
	for i, c := range final {
		results[i] = SearchResult{Entry: c.entry}
	}
	return results
}

// isUnderDir reports whether path is dir itself or nested inside it.
func isUnderDir(path, dir string) bool {
	if path == dir {
		return true
	}
	sep := string(filepath.Separator)
	if !strings.HasSuffix(dir, sep) {
		dir += sep
	}
	return strings.HasPrefix(path, dir)
}

// rankCandidates returns every match for query in idx, in final rank order
// (name matches by quality, then recency, then frequency; content-only
// matches appended after), before any limit or scope allocation is applied.
func rankCandidates(idx *Index, frecency *Frecency, query string, now time.Time) []ranked {
	entries := idx.Snapshot()
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}

	matches := fuzzy.Find(query, names)
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

	return append(candidates, contentCandidates...)
}
