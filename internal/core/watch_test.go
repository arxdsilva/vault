package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	t.Run("reports a created file", func(t *testing.T) {
		dir := t.TempDir()
		w, err := NewWatcher(dir)
		if err != nil {
			t.Fatalf("NewWatcher: %v", err)
		}
		defer w.Close()

		path := filepath.Join(dir, "new.txt")
		writeFile(t, path, "hi")

		change := awaitChange(t, w, Created)
		if filepath.Clean(change.Path) != filepath.Clean(path) {
			t.Errorf("change.Path = %q, want %q", change.Path, path)
		}
	})

	t.Run("reports a removed file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "gone.txt")
		writeFile(t, path, "hi")

		w, err := NewWatcher(dir)
		if err != nil {
			t.Fatalf("NewWatcher: %v", err)
		}
		defer w.Close()

		if err := os.Remove(path); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		change := awaitChange(t, w, Removed)
		if filepath.Clean(change.Path) != filepath.Clean(path) {
			t.Errorf("change.Path = %q, want %q", change.Path, path)
		}
	})

	t.Run("reports a renamed file", func(t *testing.T) {
		dir := t.TempDir()
		oldPath := filepath.Join(dir, "before.txt")
		writeFile(t, oldPath, "hi")

		w, err := NewWatcher(dir)
		if err != nil {
			t.Fatalf("NewWatcher: %v", err)
		}
		defer w.Close()

		if err := os.Rename(oldPath, filepath.Join(dir, "after.txt")); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		awaitChange(t, w, Renamed)
	})

	t.Run("reports a modified file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "existing.txt")
		writeFile(t, path, "hi")

		w, err := NewWatcher(dir)
		if err != nil {
			t.Fatalf("NewWatcher: %v", err)
		}
		defer w.Close()

		writeFile(t, path, "changed")

		awaitChange(t, w, Modified)
	})

	t.Run("Close stops delivering events", func(t *testing.T) {
		dir := t.TempDir()
		w, err := NewWatcher(dir)
		if err != nil {
			t.Fatalf("NewWatcher: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		select {
		case _, ok := <-w.Events:
			if ok {
				t.Error("Events delivered a value after Close")
			}
		case <-time.After(200 * time.Millisecond):
			t.Error("Events channel did not close within timeout")
		}
	})

	t.Run("missing directory returns an error", func(t *testing.T) {
		_, err := NewWatcher(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("NewWatcher: got nil error, want non-nil")
		}
	})
}

func awaitChange(t *testing.T, w *Watcher, want ChangeOp) Change {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case change := <-w.Events:
			if change.Op == want {
				return change
			}
		case err := <-w.Errors:
			t.Fatalf("watcher error: %v", err)
		case <-deadline:
			t.Fatalf("timed out waiting for change op %v", want)
		}
	}
}
