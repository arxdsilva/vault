package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildIndex(t *testing.T) {
	t.Run("indexes files and directories recursively", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, filepath.Join(root, "sub"))
		writeFile(t, filepath.Join(root, "top.txt"), "x")
		writeFile(t, filepath.Join(root, "sub", "nested.txt"), "y")

		idx, err := BuildIndex(root)
		if err != nil {
			t.Fatalf("BuildIndex: %v", err)
		}

		names := namesOf(idx.Snapshot())
		if !names["top.txt"] || !names["sub"] || !names["nested.txt"] {
			t.Errorf("snapshot missing expected entries: %v", names)
		}
	})

	t.Run("excludes hidden files and directories", func(t *testing.T) {
		root := t.TempDir()
		mustMkdirAll(t, filepath.Join(root, ".git"))
		writeFile(t, filepath.Join(root, ".git", "config"), "x")
		writeFile(t, filepath.Join(root, ".hidden.txt"), "x")
		writeFile(t, filepath.Join(root, "visible.txt"), "x")

		idx, err := BuildIndex(root)
		if err != nil {
			t.Fatalf("BuildIndex: %v", err)
		}

		names := namesOf(idx.Snapshot())
		if names[".git"] || names["config"] || names[".hidden.txt"] {
			t.Errorf("snapshot should exclude hidden entries: %v", names)
		}
		if !names["visible.txt"] {
			t.Errorf("snapshot missing visible.txt: %v", names)
		}
	})

	t.Run("skips unreadable subtrees instead of failing", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root: permission bits don't restrict access")
		}
		root := t.TempDir()
		locked := filepath.Join(root, "locked")
		mustMkdirAll(t, locked)
		writeFile(t, filepath.Join(locked, "secret.txt"), "x")
		writeFile(t, filepath.Join(root, "visible.txt"), "x")
		if err := os.Chmod(locked, 0o000); err != nil {
			t.Fatalf("Chmod: %v", err)
		}
		t.Cleanup(func() { os.Chmod(locked, 0o755) })

		idx, err := BuildIndex(root)
		if err != nil {
			t.Fatalf("BuildIndex: %v", err)
		}
		names := namesOf(idx.Snapshot())
		if !names["visible.txt"] {
			t.Errorf("snapshot missing visible.txt: %v", names)
		}
		if names["secret.txt"] {
			t.Errorf("snapshot should not reach into an unreadable subtree: %v", names)
		}
	})

	t.Run("missing root returns an error", func(t *testing.T) {
		_, err := BuildIndex(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("BuildIndex: got nil error, want non-nil")
		}
	})

	t.Run("indexes the content of a small text file", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "notes.txt")
		writeFile(t, path, "The Quarterly Report is attached.")

		idx, err := BuildIndex(root)
		if err != nil {
			t.Fatalf("BuildIndex: %v", err)
		}

		matches := idx.MatchContentPaths(strings.ToLower("quarterly"))
		if len(matches) != 1 || matches[0] != path {
			t.Errorf("MatchContentPaths() = %v, want [%s]", matches, path)
		}
	})

	t.Run("does not index the content of a non-text file", func(t *testing.T) {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "photo.png"), "quarterly report")

		idx, err := BuildIndex(root)
		if err != nil {
			t.Fatalf("BuildIndex: %v", err)
		}

		if matches := idx.MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() = %v, want none", matches)
		}
	})
}

func TestIndexUpsertRemove(t *testing.T) {
	idx := NewIndex()

	idx.Upsert(IndexEntry{Name: "a.txt", Path: "/root/a.txt"})
	idx.Upsert(IndexEntry{Name: "b.txt", Path: "/root/b.txt"})
	if len(idx.Snapshot()) != 2 {
		t.Fatalf("len(Snapshot()) = %d, want 2", len(idx.Snapshot()))
	}

	// Upserting the same path again updates in place rather than duplicating.
	idx.Upsert(IndexEntry{Name: "a-renamed.txt", Path: "/root/a.txt"})
	snap := idx.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("len(Snapshot()) after re-upsert = %d, want 2", len(snap))
	}
	names := namesOf(snap)
	if !names["a-renamed.txt"] {
		t.Errorf("expected updated name a-renamed.txt, got %v", names)
	}

	idx.Remove("/root/a.txt")
	snap = idx.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("len(Snapshot()) after Remove = %d, want 1", len(snap))
	}
	if snap[0].Path != "/root/b.txt" {
		t.Errorf("remaining entry = %+v, want /root/b.txt", snap[0])
	}
}

func TestIndexContent(t *testing.T) {
	t.Run("MatchContentPaths finds a case-insensitive substring", func(t *testing.T) {
		idx := NewIndex()
		idx.Upsert(IndexEntry{Name: "notes.txt", Path: "/notes.txt"})
		idx.UpsertContent("/notes.txt", strings.ToLower("The Quarterly Report is attached."))

		matches := idx.MatchContentPaths(strings.ToLower("quarterly"))
		if len(matches) != 1 || matches[0] != "/notes.txt" {
			t.Errorf("MatchContentPaths() = %v, want [/notes.txt]", matches)
		}
	})

	t.Run("MatchContentPaths finds nothing when the substring is absent", func(t *testing.T) {
		idx := NewIndex()
		idx.Upsert(IndexEntry{Name: "notes.txt", Path: "/notes.txt"})
		idx.UpsertContent("/notes.txt", strings.ToLower("nothing relevant here"))

		if matches := idx.MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() = %v, want none", matches)
		}
	})

	t.Run("Get returns the indexed entry for a path", func(t *testing.T) {
		idx := NewIndex()
		idx.Upsert(IndexEntry{Name: "a.txt", Path: "/a.txt"})

		e, ok := idx.Get("/a.txt")
		if !ok || e.Name != "a.txt" {
			t.Errorf("Get() = %+v, %v, want a.txt entry, true", e, ok)
		}

		if _, ok := idx.Get("/missing.txt"); ok {
			t.Error("Get() on a missing path = true, want false")
		}
	})

	t.Run("Remove drops content along with the entry", func(t *testing.T) {
		idx := NewIndex()
		idx.Upsert(IndexEntry{Name: "a.txt", Path: "/a.txt"})
		idx.UpsertContent("/a.txt", "quarterly")

		idx.Remove("/a.txt")

		if matches := idx.MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() after Remove = %v, want none", matches)
		}
	})
}

func namesOf(entries []IndexEntry) map[string]bool {
	m := make(map[string]bool, len(entries))
	for _, e := range entries {
		m[e.Name] = true
	}
	return m
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
}
