package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDir(t *testing.T) {
	t.Run("lists files and directories with metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "a.txt"), "hello")
		if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}

		entries, err := ListDir(dir)
		if err != nil {
			t.Fatalf("ListDir: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("got %d entries, want 2: %+v", len(entries), entries)
		}

		byName := map[string]Entry{}
		for _, e := range entries {
			byName[e.Name] = e
		}

		file, ok := byName["a.txt"]
		if !ok {
			t.Fatalf("missing entry for a.txt: %+v", entries)
		}
		if file.IsDir {
			t.Error("a.txt: IsDir = true, want false")
		}
		if file.Size != 5 {
			t.Errorf("a.txt: Size = %d, want 5", file.Size)
		}
		if file.Path != filepath.Join(dir, "a.txt") {
			t.Errorf("a.txt: Path = %q, want %q", file.Path, filepath.Join(dir, "a.txt"))
		}

		sub, ok := byName["sub"]
		if !ok {
			t.Fatalf("missing entry for sub: %+v", entries)
		}
		if !sub.IsDir {
			t.Error("sub: IsDir = false, want true")
		}
	})

	t.Run("missing directory returns an error", func(t *testing.T) {
		_, err := ListDir(filepath.Join(t.TempDir(), "does-not-exist"))
		if err == nil {
			t.Fatal("ListDir: got nil error, want non-nil")
		}
	})

	t.Run("empty directory returns an empty, non-nil slice", func(t *testing.T) {
		entries, err := ListDir(t.TempDir())
		if err != nil {
			t.Fatalf("ListDir: %v", err)
		}
		if entries == nil {
			t.Error("entries = nil, want empty slice")
		}
		if len(entries) != 0 {
			t.Errorf("len(entries) = %d, want 0", len(entries))
		}
	})
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
