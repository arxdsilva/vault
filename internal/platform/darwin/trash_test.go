//go:build darwin

package darwin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTrash(t *testing.T) {
	t.Run("moves a file into the trash directory", func(t *testing.T) {
		trashDir := t.TempDir()
		srcDir := t.TempDir()
		src := filepath.Join(srcDir, "doomed.txt")
		if err := os.WriteFile(src, []byte("bye"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		tr := &Trash{Dir: trashDir}
		if err := tr.Trash(src); err != nil {
			t.Fatalf("Trash: %v", err)
		}

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Errorf("source still exists at %q, stat err = %v", src, err)
		}
		dest := filepath.Join(trashDir, "doomed.txt")
		b, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("ReadFile(%q): %v", dest, err)
		}
		if string(b) != "bye" {
			t.Errorf("trashed content = %q, want %q", b, "bye")
		}
	})

	t.Run("disambiguates a name collision", func(t *testing.T) {
		trashDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(trashDir, "doomed.txt"), []byte("already here"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		srcDir := t.TempDir()
		src := filepath.Join(srcDir, "doomed.txt")
		if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		tr := &Trash{Dir: trashDir}
		if err := tr.Trash(src); err != nil {
			t.Fatalf("Trash: %v", err)
		}

		b, err := os.ReadFile(filepath.Join(trashDir, "doomed 2.txt"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(b) != "new" {
			t.Errorf("disambiguated content = %q, want %q", b, "new")
		}
	})

	t.Run("creates the trash directory if missing", func(t *testing.T) {
		trashDir := filepath.Join(t.TempDir(), "does-not-exist-yet")
		srcDir := t.TempDir()
		src := filepath.Join(srcDir, "f.txt")
		if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		tr := &Trash{Dir: trashDir}
		if err := tr.Trash(src); err != nil {
			t.Fatalf("Trash: %v", err)
		}
		if _, err := os.Stat(filepath.Join(trashDir, "f.txt")); err != nil {
			t.Errorf("expected trashed file to exist: %v", err)
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		tr := &Trash{Dir: t.TempDir()}
		if err := tr.Trash(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
			t.Fatal("Trash: got nil error, want non-nil")
		}
	})

	t.Run("Location returns and creates the trash directory", func(t *testing.T) {
		trashDir := filepath.Join(t.TempDir(), "not-created-yet")
		tr := &Trash{Dir: trashDir}

		got, err := tr.Location()
		if err != nil {
			t.Fatalf("Location: %v", err)
		}
		if got != trashDir {
			t.Errorf("Location() = %q, want %q", got, trashDir)
		}
		if info, err := os.Stat(trashDir); err != nil || !info.IsDir() {
			t.Errorf("expected Location to create %q as a directory", trashDir)
		}
	})

	t.Run("NewTrash resolves ~/.Trash under the current user's home", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		tr := NewTrash()
		dir, err := tr.dir()
		if err != nil {
			t.Fatalf("dir: %v", err)
		}
		if dir != filepath.Join(home, ".Trash") {
			t.Errorf("dir = %q, want %q", dir, filepath.Join(home, ".Trash"))
		}
	})
}
