package core

import (
	"os"
	"path/filepath"
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
