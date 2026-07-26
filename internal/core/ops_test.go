package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRename(t *testing.T) {
	t.Run("renames a file", func(t *testing.T) {
		dir := t.TempDir()
		oldPath := filepath.Join(dir, "old.txt")
		newPath := filepath.Join(dir, "new.txt")
		writeFile(t, oldPath, "content")

		if err := Rename(oldPath, newPath); err != nil {
			t.Fatalf("Rename: %v", err)
		}
		assertExists(t, newPath)
		assertMissing(t, oldPath)
	})

	t.Run("refuses to overwrite an existing destination", func(t *testing.T) {
		dir := t.TempDir()
		oldPath := filepath.Join(dir, "old.txt")
		newPath := filepath.Join(dir, "new.txt")
		writeFile(t, oldPath, "content")
		writeFile(t, newPath, "existing")

		if err := Rename(oldPath, newPath); err == nil {
			t.Fatal("Rename: got nil error, want non-nil (destination exists)")
		}
		if got := readFile(t, newPath); got != "existing" {
			t.Errorf("destination was overwritten: got %q", got)
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		dir := t.TempDir()
		err := Rename(filepath.Join(dir, "missing.txt"), filepath.Join(dir, "new.txt"))
		if err == nil {
			t.Fatal("Rename: got nil error, want non-nil (missing source)")
		}
	})
}

func TestCopy(t *testing.T) {
	t.Run("copies a file", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		writeFile(t, src, "content")

		if err := Copy(src, dst); err != nil {
			t.Fatalf("Copy: %v", err)
		}
		assertExists(t, src)
		if got := readFile(t, dst); got != "content" {
			t.Errorf("dst content = %q, want %q", got, "content")
		}
	})

	t.Run("copies a directory recursively", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src")
		dst := filepath.Join(dir, "dst")
		if err := os.MkdirAll(filepath.Join(src, "nested"), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		writeFile(t, filepath.Join(src, "top.txt"), "top")
		writeFile(t, filepath.Join(src, "nested", "deep.txt"), "deep")

		if err := Copy(src, dst); err != nil {
			t.Fatalf("Copy: %v", err)
		}
		if got := readFile(t, filepath.Join(dst, "top.txt")); got != "top" {
			t.Errorf("dst/top.txt = %q, want %q", got, "top")
		}
		if got := readFile(t, filepath.Join(dst, "nested", "deep.txt")); got != "deep" {
			t.Errorf("dst/nested/deep.txt = %q, want %q", got, "deep")
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		dir := t.TempDir()
		err := Copy(filepath.Join(dir, "missing.txt"), filepath.Join(dir, "dst.txt"))
		if err == nil {
			t.Fatal("Copy: got nil error, want non-nil (missing source)")
		}
	})

	t.Run("destination parent missing returns an error", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		writeFile(t, src, "content")

		err := Copy(src, filepath.Join(dir, "no-such-parent", "dst.txt"))
		if err == nil {
			t.Fatal("Copy: got nil error, want non-nil (missing destination parent)")
		}
	})

	t.Run("unreadable source directory returns an error", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("running as root: permission bits don't restrict access")
		}
		dir := t.TempDir()
		src := filepath.Join(dir, "src")
		if err := os.Mkdir(src, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		writeFile(t, filepath.Join(src, "f.txt"), "f")
		if err := os.Chmod(src, 0o000); err != nil {
			t.Fatalf("Chmod: %v", err)
		}
		t.Cleanup(func() { os.Chmod(src, 0o755) })

		err := Copy(src, filepath.Join(dir, "dst"))
		if err == nil {
			t.Fatal("Copy: got nil error, want non-nil (unreadable source directory)")
		}
	})
}

func TestMove(t *testing.T) {
	t.Run("moves a file within the same directory", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		writeFile(t, src, "content")

		if err := Move(src, dst); err != nil {
			t.Fatalf("Move: %v", err)
		}
		assertMissing(t, src)
		if got := readFile(t, dst); got != "content" {
			t.Errorf("dst content = %q, want %q", got, "content")
		}
	})

	t.Run("moves a directory", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src")
		dst := filepath.Join(dir, "dst")
		if err := os.Mkdir(src, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		writeFile(t, filepath.Join(src, "f.txt"), "f")

		if err := Move(src, dst); err != nil {
			t.Fatalf("Move: %v", err)
		}
		assertMissing(t, src)
		if got := readFile(t, filepath.Join(dst, "f.txt")); got != "f" {
			t.Errorf("dst/f.txt = %q, want %q", got, "f")
		}
	})

	t.Run("refuses to overwrite an existing destination", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		writeFile(t, src, "content")
		writeFile(t, dst, "existing")

		if err := Move(src, dst); err == nil {
			t.Fatal("Move: got nil error, want non-nil (destination exists)")
		}
		assertExists(t, src)
		if got := readFile(t, dst); got != "existing" {
			t.Errorf("destination was overwritten: got %q", got)
		}
	})
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %q to exist: %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %q to be gone, stat err = %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(b)
}
