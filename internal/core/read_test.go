package core

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	t.Run("returns a file's contents", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "note.txt")
		writeFile(t, path, "hello vault")

		got, err := ReadFile(path, DefaultReadFileMaxBytes)
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if got != "hello vault" {
			t.Errorf("ReadFile() = %q, want %q", got, "hello vault")
		}
	})

	t.Run("a file over maxBytes returns an error instead of truncating", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "big.txt")
		writeFile(t, path, strings.Repeat("x", 10))

		if _, err := ReadFile(path, 5); err == nil {
			t.Fatal("ReadFile: got nil error, want non-nil")
		}
	})

	t.Run("a directory returns an error", func(t *testing.T) {
		if _, err := ReadFile(t.TempDir(), DefaultReadFileMaxBytes); err == nil {
			t.Fatal("ReadFile: got nil error, want non-nil")
		}
	})

	t.Run("a missing path returns an error", func(t *testing.T) {
		if _, err := ReadFile(filepath.Join(t.TempDir(), "missing.txt"), DefaultReadFileMaxBytes); err == nil {
			t.Fatal("ReadFile: got nil error, want non-nil")
		}
	})
}
