package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIndexer(t *testing.T) {
	t.Run("builds an initial index of the tree", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "a.txt"), "x")

		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		if !namesOf(ix.Index().Snapshot())["a.txt"] {
			t.Errorf("initial index missing a.txt: %v", ix.Index().Snapshot())
		}
	})

	t.Run("indexes a file created after startup", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		writeFile(t, filepath.Join(root, "new.txt"), "x")

		awaitIndexed(t, ix, "new.txt", true)
	})

	t.Run("follows a newly created subdirectory", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

		sub := filepath.Join(root, "sub")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		awaitIndexed(t, ix, "sub", true)

		writeFile(t, filepath.Join(sub, "deep.txt"), "x")
		awaitIndexed(t, ix, "deep.txt", true)
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

		if err := os.Remove(path); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		awaitIndexed(t, ix, "gone.txt", false)
	})

	t.Run("ignores hidden files created after startup", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

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
		if err := ix.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		writeFile(t, filepath.Join(root, "after-close.txt"), "x")
		time.Sleep(200 * time.Millisecond)
		if namesOf(ix.Index().Snapshot())["after-close.txt"] {
			t.Error("index should not update after Close")
		}
	})

	t.Run("missing root returns an error", func(t *testing.T) {
		_, err := StartIndexer(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("StartIndexer: got nil error, want non-nil")
		}
	})

	t.Run("indexes the content of a file created after startup", func(t *testing.T) {
		root := t.TempDir()
		ix, err := StartIndexer(root)
		if err != nil {
			t.Fatalf("StartIndexer: %v", err)
		}
		defer ix.Close()

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
