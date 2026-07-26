package core

import (
	"path/filepath"
	"testing"
)

func TestCopyInto(t *testing.T) {
	t.Run("copies into a different directory under the same name", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()
		src := filepath.Join(srcDir, "report.txt")
		writeFile(t, src, "hello")

		got, err := CopyInto(src, destDir)
		if err != nil {
			t.Fatalf("CopyInto: %v", err)
		}
		want := filepath.Join(destDir, "report.txt")
		if got != want {
			t.Errorf("got = %q, want %q", got, want)
		}
		if content := readFile(t, got); content != "hello" {
			t.Errorf("content = %q, want %q", content, "hello")
		}
	})

	t.Run("disambiguates when pasting into the same directory as the source", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "report.txt")
		writeFile(t, src, "hello")

		got, err := CopyInto(src, dir)
		if err != nil {
			t.Fatalf("CopyInto: %v", err)
		}
		want := filepath.Join(dir, "report copy.txt")
		if got != want {
			t.Errorf("got = %q, want %q", got, want)
		}
		assertExists(t, src)
	})

	t.Run("disambiguates when the destination already has a same-named file", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()
		src := filepath.Join(srcDir, "report.txt")
		writeFile(t, src, "new")
		writeFile(t, filepath.Join(destDir, "report.txt"), "existing")

		got, err := CopyInto(src, destDir)
		if err != nil {
			t.Fatalf("CopyInto: %v", err)
		}
		want := filepath.Join(destDir, "report copy.txt")
		if got != want {
			t.Errorf("got = %q, want %q", got, want)
		}
	})

	t.Run("copies a directory recursively", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()
		src := filepath.Join(srcDir, "project")
		mustMkdirAll(t, src)
		writeFile(t, filepath.Join(src, "f.txt"), "x")

		got, err := CopyInto(src, destDir)
		if err != nil {
			t.Fatalf("CopyInto: %v", err)
		}
		if got != filepath.Join(destDir, "project") {
			t.Errorf("got = %q, want %q", got, filepath.Join(destDir, "project"))
		}
		if content := readFile(t, filepath.Join(got, "f.txt")); content != "x" {
			t.Errorf("content = %q, want %q", content, "x")
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		_, err := CopyInto(filepath.Join(t.TempDir(), "missing.txt"), t.TempDir())
		if err == nil {
			t.Fatal("CopyInto: got nil error, want non-nil")
		}
	})
}
