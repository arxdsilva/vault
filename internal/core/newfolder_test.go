package core

import (
	"path/filepath"
	"testing"
)

func TestNewFolder(t *testing.T) {
	t.Run("creates a folder with the given name", func(t *testing.T) {
		dir := t.TempDir()
		path, err := NewFolder(dir, "Untitled Folder")
		if err != nil {
			t.Fatalf("NewFolder: %v", err)
		}
		if path != filepath.Join(dir, "Untitled Folder") {
			t.Errorf("path = %q, want %q", path, filepath.Join(dir, "Untitled Folder"))
		}
		assertExists(t, path)
	})

	t.Run("disambiguates when the name is taken", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := NewFolder(dir, "Untitled Folder"); err != nil {
			t.Fatalf("NewFolder: %v", err)
		}
		path, err := NewFolder(dir, "Untitled Folder")
		if err != nil {
			t.Fatalf("NewFolder: %v", err)
		}
		if path != filepath.Join(dir, "Untitled Folder 2") {
			t.Errorf("path = %q, want %q", path, filepath.Join(dir, "Untitled Folder 2"))
		}
	})

	t.Run("missing parent returns an error", func(t *testing.T) {
		_, err := NewFolder(filepath.Join(t.TempDir(), "missing"), "Untitled Folder")
		if err == nil {
			t.Fatal("NewFolder: got nil error, want non-nil")
		}
	})
}
