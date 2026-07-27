//go:build linux

package linux

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrash(t *testing.T) {
	t.Run("moves a file into files/ and writes a paired .trashinfo", func(t *testing.T) {
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

		b, err := os.ReadFile(filepath.Join(trashDir, "files", "doomed.txt"))
		if err != nil {
			t.Fatalf("ReadFile(files/doomed.txt): %v", err)
		}
		if string(b) != "bye" {
			t.Errorf("trashed content = %q, want %q", b, "bye")
		}

		info, err := os.ReadFile(filepath.Join(trashDir, "info", "doomed.txt.trashinfo"))
		if err != nil {
			t.Fatalf("ReadFile(info/doomed.txt.trashinfo): %v", err)
		}
		if !strings.Contains(string(info), "[Trash Info]") {
			t.Errorf(".trashinfo missing header: %q", info)
		}
		if !strings.Contains(string(info), "Path=") {
			t.Errorf(".trashinfo missing Path=: %q", info)
		}
		if !strings.Contains(string(info), "DeletionDate=") {
			t.Errorf(".trashinfo missing DeletionDate=: %q", info)
		}
	})

	t.Run("disambiguates a name collision, keeping the file and .trashinfo paired", func(t *testing.T) {
		trashDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(trashDir, "files"), 0o700); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(filepath.Join(trashDir, "files", "doomed.txt"), []byte("already here"), 0o644); err != nil {
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

		b, err := os.ReadFile(filepath.Join(trashDir, "files", "doomed 2.txt"))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(b) != "new" {
			t.Errorf("disambiguated content = %q, want %q", b, "new")
		}
		if _, err := os.Stat(filepath.Join(trashDir, "info", "doomed 2.txt.trashinfo")); err != nil {
			t.Errorf("expected a paired doomed 2.txt.trashinfo: %v", err)
		}
	})

	t.Run("creates the trash directories if missing", func(t *testing.T) {
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
		if _, err := os.Stat(filepath.Join(trashDir, "files", "f.txt")); err != nil {
			t.Errorf("expected trashed file to exist: %v", err)
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		tr := &Trash{Dir: t.TempDir()}
		if err := tr.Trash(filepath.Join(t.TempDir(), "missing.txt")); err == nil {
			t.Fatal("Trash: got nil error, want non-nil")
		}
	})

	t.Run("Location returns and creates the files subdirectory", func(t *testing.T) {
		trashDir := t.TempDir()
		tr := &Trash{Dir: trashDir}

		got, err := tr.Location()
		if err != nil {
			t.Fatalf("Location: %v", err)
		}
		want := filepath.Join(trashDir, "files")
		if got != want {
			t.Errorf("Location() = %q, want %q", got, want)
		}
		if info, err := os.Stat(got); err != nil || !info.IsDir() {
			t.Errorf("expected Location to create %q as a directory", got)
		}
		if info, err := os.Stat(filepath.Join(trashDir, "info")); err != nil || !info.IsDir() {
			t.Errorf("expected Location to also create the paired info/ directory")
		}
	})

	t.Run("NewTrash resolves ~/.local/share/Trash under the current user's home", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("XDG_DATA_HOME", "")

		tr := NewTrash()
		dir, err := tr.dir()
		if err != nil {
			t.Fatalf("dir: %v", err)
		}
		want := filepath.Join(home, ".local", "share", "Trash")
		if dir != want {
			t.Errorf("dir = %q, want %q", dir, want)
		}
	})

	t.Run("NewTrash prefers XDG_DATA_HOME when set", func(t *testing.T) {
		dataHome := t.TempDir()
		t.Setenv("XDG_DATA_HOME", dataHome)

		tr := NewTrash()
		dir, err := tr.dir()
		if err != nil {
			t.Fatalf("dir: %v", err)
		}
		want := filepath.Join(dataHome, "Trash")
		if dir != want {
			t.Errorf("dir = %q, want %q", dir, want)
		}
	})
}

func TestEncodeTrashPath(t *testing.T) {
	got := encodeTrashPath("/home/user/my file.txt")
	if got != "/home/user/my%20file.txt" {
		t.Errorf("encodeTrashPath() = %q, want %q", got, "/home/user/my%20file.txt")
	}
}
