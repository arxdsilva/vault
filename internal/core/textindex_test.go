package core

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestIsIndexableTextFile(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"notes.txt", true},
		{"main.go", true},
		{"App.tsx", true},
		{"README", true},   // extensionless: treated as text, mirrors previewKind.ts
		{"Makefile", true}, // extensionless
		{"photo.png", false},
		{"movie.mp4", false},
		{"archive.zip", false},
		{"binary.exe", false},
	}
	for _, c := range cases {
		if got := isIndexableTextFile(c.name); got != c.want {
			t.Errorf("isIndexableTextFile(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestIndexFileContent(t *testing.T) {
	t.Run("indexes content of a small text file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "notes.txt")
		writeFile(t, path, "The Quarterly Report is attached.")

		idx := NewIndex()
		indexFileContent(idx, path, "notes.txt", int64(len("The Quarterly Report is attached.")))

		matches := idx.MatchContentPaths(strings.ToLower("quarterly"))
		if len(matches) != 1 || matches[0] != path {
			t.Errorf("MatchContentPaths() = %v, want [%s]", matches, path)
		}
	})

	t.Run("skips a file over the size cap", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "big.txt")
		writeFile(t, path, "quarterly report")

		idx := NewIndex()
		indexFileContent(idx, path, "big.txt", DefaultReadFileMaxBytes+1)

		if matches := idx.MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() = %v, want none (file over size cap)", matches)
		}
	})

	t.Run("skips a non-text extension", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "photo.png")
		writeFile(t, path, "quarterly report")

		idx := NewIndex()
		indexFileContent(idx, path, "photo.png", int64(len("quarterly report")))

		if matches := idx.MatchContentPaths("quarterly"); len(matches) != 0 {
			t.Errorf("MatchContentPaths() = %v, want none (non-text extension)", matches)
		}
	})

	t.Run("a missing file is a silent no-op", func(t *testing.T) {
		idx := NewIndex()
		indexFileContent(idx, filepath.Join(t.TempDir(), "missing.txt"), "missing.txt", 10)
		if matches := idx.MatchContentPaths(""); len(matches) != 0 {
			t.Errorf("MatchContentPaths() = %v, want none", matches)
		}
	})
}
