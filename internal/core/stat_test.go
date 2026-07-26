package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStat(t *testing.T) {
	t.Run("returns metadata for a file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "a.txt")
		writeFile(t, path, "hello")

		e, err := Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if e.Name != "a.txt" {
			t.Errorf("Name = %q, want a.txt", e.Name)
		}
		if e.Path != path {
			t.Errorf("Path = %q, want %q", e.Path, path)
		}
		if e.IsDir {
			t.Error("IsDir = true, want false")
		}
		if e.Size != 5 {
			t.Errorf("Size = %d, want 5", e.Size)
		}
	})

	t.Run("returns metadata for a directory", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "sub")
		if err := os.Mkdir(sub, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}

		e, err := Stat(sub)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		if e.Name != "sub" {
			t.Errorf("Name = %q, want sub", e.Name)
		}
		if !e.IsDir {
			t.Error("IsDir = false, want true")
		}
	})

	t.Run("missing path returns an error", func(t *testing.T) {
		_, err := Stat(filepath.Join(t.TempDir(), "does-not-exist"))
		if err == nil {
			t.Fatal("Stat: got nil error, want non-nil")
		}
	})
}
