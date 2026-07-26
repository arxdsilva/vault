package core

import (
	"testing"
	"time"
)

func newTestIndex(entries ...IndexEntry) *Index {
	idx := NewIndex()
	for _, e := range entries {
		idx.Upsert(e)
	}
	return idx
}

func TestSearch(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("only fuzzy-matching entries are returned", func(t *testing.T) {
		idx := newTestIndex(
			IndexEntry{Name: "report.txt", Path: "/report.txt"},
			IndexEntry{Name: "budget.xls", Path: "/budget.xls"},
		)
		results := Search(idx, NewFrecency(), "rep", now, 10)
		if len(results) != 1 || results[0].Entry.Name != "report.txt" {
			t.Fatalf("results = %+v, want just report.txt", results)
		}
	})

	t.Run("empty query returns no results", func(t *testing.T) {
		idx := newTestIndex(IndexEntry{Name: "report.txt", Path: "/report.txt"})
		results := Search(idx, NewFrecency(), "", now, 10)
		if len(results) != 0 {
			t.Errorf("results = %+v, want none", results)
		}
	})

	t.Run("a clearly better fuzzy match always ranks first", func(t *testing.T) {
		idx := newTestIndex(
			IndexEntry{Name: "report.txt", Path: "/report.txt"},          // strong match for "report"
			IndexEntry{Name: "old_report_final.docx", Path: "/old.docx"}, // weak match, but heavily used
		)
		f := NewFrecency()
		for i := 0; i < 50; i++ {
			f.Touch("/old.docx", now)
		}

		results := Search(idx, f, "report", now, 10)
		if results[0].Entry.Path != "/report.txt" {
			t.Fatalf("first result = %+v, want /report.txt (match quality beats frecency)", results[0])
		}
	})

	t.Run("among equal fuzzy matches, more recently accessed ranks first", func(t *testing.T) {
		idx := newTestIndex(
			IndexEntry{Name: "report.txt", Path: "/a/report.txt"},
			IndexEntry{Name: "report.txt", Path: "/b/report.txt"},
		)
		f := NewFrecency()
		f.Touch("/a/report.txt", now.Add(-time.Hour))
		f.Touch("/b/report.txt", now)

		results := Search(idx, f, "report", now, 10)
		if results[0].Entry.Path != "/b/report.txt" {
			t.Fatalf("first result = %+v, want /b/report.txt (touched more recently)", results[0])
		}
	})

	t.Run("among equal match and recency, more frequently accessed ranks first", func(t *testing.T) {
		idx := newTestIndex(
			IndexEntry{Name: "report.txt", Path: "/a/report.txt"},
			IndexEntry{Name: "report.txt", Path: "/b/report.txt"},
		)
		f := NewFrecency()
		f.Touch("/a/report.txt", now)
		f.Touch("/b/report.txt", now)
		f.Touch("/b/report.txt", now)

		results := Search(idx, f, "report", now, 10)
		if results[0].Entry.Path != "/b/report.txt" {
			t.Fatalf("first result = %+v, want /b/report.txt (touched more often)", results[0])
		}
	})

	t.Run("never-touched entries rank after touched entries of equal match quality", func(t *testing.T) {
		idx := newTestIndex(
			IndexEntry{Name: "report.txt", Path: "/a/report.txt"},
			IndexEntry{Name: "report.txt", Path: "/b/report.txt"},
		)
		f := NewFrecency()
		f.Touch("/b/report.txt", now)

		results := Search(idx, f, "report", now, 10)
		if results[0].Entry.Path != "/b/report.txt" {
			t.Fatalf("first result = %+v, want /b/report.txt (touched vs. never touched)", results[0])
		}
	})

	t.Run("limit caps the number of results", func(t *testing.T) {
		idx := newTestIndex(
			IndexEntry{Name: "report1.txt", Path: "/report1.txt"},
			IndexEntry{Name: "report2.txt", Path: "/report2.txt"},
			IndexEntry{Name: "report3.txt", Path: "/report3.txt"},
		)
		results := Search(idx, NewFrecency(), "report", now, 2)
		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want 2", len(results))
		}
	})
}
