package core

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestCompress(t *testing.T) {
	t.Run("zips a single file", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "report.txt")
		writeFile(t, src, "hello")

		zipPath, err := Compress(src)
		if err != nil {
			t.Fatalf("Compress: %v", err)
		}
		if zipPath != filepath.Join(dir, "report.zip") {
			t.Errorf("zipPath = %q, want %q", zipPath, filepath.Join(dir, "report.zip"))
		}

		names := zipEntryNames(t, zipPath)
		if len(names) != 1 || names[0] != "report.txt" {
			t.Errorf("zip entries = %v, want [report.txt]", names)
		}
	})

	t.Run("zips a directory recursively, preserving relative structure", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "project")
		mustMkdirAll(t, filepath.Join(src, "nested"))
		writeFile(t, filepath.Join(src, "top.txt"), "top")
		writeFile(t, filepath.Join(src, "nested", "deep.txt"), "deep")

		zipPath, err := Compress(src)
		if err != nil {
			t.Fatalf("Compress: %v", err)
		}

		names := zipEntryNames(t, zipPath)
		want := map[string]bool{
			filepath.ToSlash(filepath.Join("project", "top.txt")):            true,
			filepath.ToSlash(filepath.Join("project", "nested", "deep.txt")): true,
		}
		if len(names) != len(want) {
			t.Fatalf("zip entries = %v, want %v", names, want)
		}
		for _, n := range names {
			if !want[n] {
				t.Errorf("unexpected zip entry %q", n)
			}
		}
	})

	t.Run("disambiguates when the destination zip already exists", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "report.txt")
		writeFile(t, src, "hello")
		writeFile(t, filepath.Join(dir, "report.zip"), "existing")

		zipPath, err := Compress(src)
		if err != nil {
			t.Fatalf("Compress: %v", err)
		}
		if zipPath != filepath.Join(dir, "report 2.zip") {
			t.Errorf("zipPath = %q, want %q", zipPath, filepath.Join(dir, "report 2.zip"))
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		_, err := Compress(filepath.Join(t.TempDir(), "missing.txt"))
		if err == nil {
			t.Fatal("Compress: got nil error, want non-nil")
		}
	})
}

func zipEntryNames(t *testing.T, path string) []string {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("zip.OpenReader(%q): %v", path, err)
	}
	defer r.Close()
	names := make([]string, len(r.File))
	for i, f := range r.File {
		names[i] = f.Name
	}
	return names
}

func TestCompressSkipsNothingUnexpected(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.txt")
	writeFile(t, src, "x")

	zipPath, err := Compress(src)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	info, err := os.Stat(zipPath)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("zip file is empty")
	}
}
