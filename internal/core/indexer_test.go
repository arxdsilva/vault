package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIndexer(t *testing.T) {
	t.Run("missing root returns an error", func(t *testing.T) {
		_, err := StartIndexer(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("StartIndexer: got nil error, want non-nil")
		}
	})

	t.Run("index starts empty until something is visited", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "a.txt"), "x")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if len(ix.Index().Snapshot()) != 0 {
			t.Errorf("index should start empty, got %v", ix.Index().Snapshot())
		}
	})

	t.Run("Visit indexes a directory's immediate contents", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, filepath.Join(root, "sub"))
		writeFile(t, filepath.Join(root, "top.txt"), "x")
		writeFile(t, filepath.Join(root, "sub", "nested.txt"), "y")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		names := namesOf(ix.Index().Snapshot())
		if !names["top.txt"] || !names["sub"] {
			t.Errorf("snapshot missing expected entries: %v", names)
		}
		if names["nested.txt"] {
			t.Errorf("Visit should not descend into subdirectories: %v", names)
		}
	})

	t.Run("Visit excludes hidden files and directories", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, filepath.Join(root, ".git"))
		writeFile(t, filepath.Join(root, ".git", "config"), "x")
		writeFile(t, filepath.Join(root, ".hidden.txt"), "x")
		writeFile(t, filepath.Join(root, "visible.txt"), "x")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		names := namesOf(ix.Index().Snapshot())
		if names[".git"] || names[".hidden.txt"] {
			t.Errorf("snapshot should exclude hidden entries: %v", names)
		}
		if !names["visible.txt"] {
			t.Errorf("snapshot missing visible.txt: %v", names)
		}
	})

	t.Run("Visit on a missing directory returns an error", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if err := ix.Visit(filepath.Join(root, "missing")); err == nil {
			t.Fatal("Visit: got nil error, want non-nil")
		}
	})

	t.Run("Visit indexes the content of a small text file", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "notes.txt")
		writeFile(t, path, "The Quarterly Report is attached.")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		matches := ix.Index().MatchContentPaths(strings.ToLower("quarterly"))
		if len(matches) != 1 || matches[0] != path {
			t.Errorf("MatchContentPaths() = %v, want [%s]", matches, path)
		}
	})

	t.Run("Visit does not index the content of a non-text file", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "photo.png"), "quarterly report")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		if matches := ix.Index().MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() = %v, want none", matches)
		}
	})

	t.Run("indexes a file created after a directory is visited", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		writeFile(t, filepath.Join(root, "new.txt"), "x")

		awaitIndexed(t, ix, "new.txt", true)
	})

	t.Run("does not pick up a file created before its directory is visited", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		writeFile(t, filepath.Join(root, "unseen.txt"), "x")
		time.Sleep(100 * time.Millisecond)

		if namesOf(ix.Index().Snapshot())["unseen.txt"] {
			t.Error("index should not include files from a never-visited directory")
		}
	})

	t.Run("a newly created subdirectory appears as an entry but isn't recursed into", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		sub := filepath.Join(root, "sub")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		awaitIndexed(t, ix, "sub", true)

		writeFile(t, filepath.Join(sub, "deep.txt"), "x")
		time.Sleep(100 * time.Millisecond)
		if namesOf(ix.Index().Snapshot())["deep.txt"] {
			t.Error("a new subdirectory should not be auto-watched until it's separately visited")
		}

		if err := ix.Visit(sub); err != nil {
			t.Fatalf("Visit(sub): %v", err)
		}
		if !namesOf(ix.Index().Snapshot())["deep.txt"] {
			t.Error("visiting the subdirectory should index its contents")
		}
	})

	t.Run("removes a deleted file from the index", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "gone.txt")
		writeFile(t, path, "x")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		if err := os.Remove(path); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		awaitIndexed(t, ix, "gone.txt", false)
	})

	t.Run("ignores hidden files created after a directory is visited", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		writeFile(t, filepath.Join(root, "visible.txt"), "x")
		awaitIndexed(t, ix, "visible.txt", true)

		writeFile(t, filepath.Join(root, ".hidden.txt"), "x")
		time.Sleep(200 * time.Millisecond)
		if namesOf(ix.Index().Snapshot())[".hidden.txt"] {
			t.Error("index should not include .hidden.txt")
		}
	})

	t.Run("Close stops indexing further changes", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}
		if err := ix.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		writeFile(t, filepath.Join(root, "after-close.txt"), "x")
		time.Sleep(200 * time.Millisecond)
		if namesOf(ix.Index().Snapshot())["after-close.txt"] {
			t.Error("index should not update after Close")
		}
	})

	t.Run("indexes the content of a file created after a directory is visited", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		path := filepath.Join(root, "notes.txt")
		writeFile(t, path, "The Quarterly Report is attached.")

		awaitContentIndexed(t, ix, path, "quarterly")
	})

	t.Run("re-indexes content when a file is modified", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "notes.txt")
		writeFile(t, path, "first draft")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		writeFile(t, path, "the quarterly report")
		awaitContentIndexed(t, ix, path, "quarterly")
	})

	t.Run("removing a file drops its content from the index", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "notes.txt")
		writeFile(t, path, "quarterly report")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()
		if err := ix.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		awaitContentIndexed(t, ix, path, "quarterly")

		if err := os.Remove(path); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		awaitIndexed(t, ix, "notes.txt", false)
		if matches := ix.Index().MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() after removal = %v, want none", matches)
		}
	})
}

func awaitContentIndexed(t *testing.T, ix *Indexer, path, lowerQuery string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		for _, p := range ix.Index().MatchContentPaths(lowerQuery) {
			if p == path {
				return
			}
		}
		select {
		case <-time.After(20 * time.Millisecond):
		case <-deadline:
			t.Fatalf("timed out waiting for %q to be content-indexed for query %q", path, lowerQuery)
		}
	}
}

func awaitIndexed(t *testing.T, ix *Indexer, name string, present bool) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		if namesOf(ix.Index().Snapshot())[name] == present {
			return
		}
		select {
		case <-time.After(20 * time.Millisecond):
		case <-deadline:
			t.Fatalf("timed out waiting for %q present=%v", name, present)
		}
	}
}
