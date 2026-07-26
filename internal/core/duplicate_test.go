package core

import (
	"path/filepath"
	"testing"
)

func TestDuplicate(t *testing.T) {
	t.Run("duplicates a file as 'name copy.ext'", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "report.txt")
		writeFile(t, src, "hello")

		dup, err := Duplicate(src)
		if err != nil {
			t.Fatalf("Duplicate: %v", err)
		}
		want := filepath.Join(dir, "report copy.txt")
		if dup != want {
			t.Errorf("dup = %q, want %q", dup, want)
		}
		if got := readFile(t, dup); got != "hello" {
			t.Errorf("dup content = %q, want %q", got, "hello")
		}
		assertExists(t, src)
	})

	t.Run("a second duplicate becomes 'name copy 2.ext'", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "report.txt")
		writeFile(t, src, "hello")

		if _, err := Duplicate(src); err != nil {
			t.Fatalf("Duplicate: %v", err)
		}
		dup2, err := Duplicate(src)
		if err != nil {
			t.Fatalf("Duplicate: %v", err)
		}
		want := filepath.Join(dir, "report copy 2.txt")
		if dup2 != want {
			t.Errorf("dup2 = %q, want %q", dup2, want)
		}
	})

	t.Run("duplicates a directory recursively", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "project")
		mustMkdirAll(t, src)
		writeFile(t, filepath.Join(src, "f.txt"), "x")

		dup, err := Duplicate(src)
		if err != nil {
			t.Fatalf("Duplicate: %v", err)
		}
		if dup != filepath.Join(dir, "project copy") {
			t.Errorf("dup = %q, want %q", dup, filepath.Join(dir, "project copy"))
		}
		if got := readFile(t, filepath.Join(dup, "f.txt")); got != "x" {
			t.Errorf("dup/f.txt = %q, want %q", got, "x")
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		_, err := Duplicate(filepath.Join(t.TempDir(), "missing.txt"))
		if err == nil {
			t.Fatal("Duplicate: got nil error, want non-nil")
		}
	})
}
