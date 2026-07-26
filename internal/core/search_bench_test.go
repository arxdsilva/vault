package core

import (
	"fmt"
	"testing"
	"time"
)

// buildSyntheticIndex builds an n-entry index shaped like a real home
// directory: a handful of top-level folders (Documents, Downloads,
// Projects, Photos, Library), each holding many files with varied
// extensions and realistic-length names, so fuzzy matching and sorting
// see the same candidate volume/shape a real search would (spec §8: an
// index of 200k files).
func buildSyntheticIndex(n int) *Index {
	idx := NewIndex()
	dirs := []string{"Documents", "Downloads", "Projects", "Photos", "Library"}
	exts := []string{"txt", "pdf", "png", "jpg", "go", "ts", "md", "json", "zip", "docx"}
	for i := 0; i < n; i++ {
		dir := dirs[i%len(dirs)]
		ext := exts[i%len(exts)]
		name := fmt.Sprintf("report_%d_final.%s", i, ext)
		path := fmt.Sprintf("/home/user/%s/sub%d/%s", dir, i%100, name)
		idx.Upsert(IndexEntry{Name: name, Path: path, IsDir: false})
	}
	return idx
}

// BenchmarkSearch200k measures query latency against a 200k-file index —
// the exact scale spec §8 sets a <100ms (after keystroke) target for.
// "report" is a deliberately broad, high-candidate-count query (every
// synthesized name contains it) so this exercises the worst case: fuzzy-
// score every entry, then sort. A narrower query only gets faster from here.
func BenchmarkSearch200k(b *testing.B) {
	idx := buildSyntheticIndex(200_000)
	frecency := NewFrecency()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(idx, frecency, "report", now, 200, "")
	}
}

// BenchmarkSearch200kScoped is the common in-app case: searching while a
// folder is open, so results are scoped to it (see Search's scopeDir doc).
func BenchmarkSearch200kScoped(b *testing.B) {
	idx := buildSyntheticIndex(200_000)
	frecency := NewFrecency()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(idx, frecency, "report", now, 200, "/home/user/Documents/sub0")
	}
}

// BenchmarkSearch200kNarrow measures a rare, specific query — the common
// case once a user has typed a few characters — against the same index.
func BenchmarkSearch200kNarrow(b *testing.B) {
	idx := buildSyntheticIndex(200_000)
	frecency := NewFrecency()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Search(idx, frecency, "report_199999", now, 200, "")
	}
}
