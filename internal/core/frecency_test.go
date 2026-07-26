package core

import (
	"path/filepath"
	"testing"
	"time"
)

func TestFrecencyTouch(t *testing.T) {
	t.Run("an untouched path has zero count and a zero LastAccess", func(t *testing.T) {
		f := NewFrecency()
		last, count := f.Stats("/never/touched")
		if !last.IsZero() {
			t.Errorf("LastAccess = %v, want zero", last)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
	})

	t.Run("touching a path records the access time and increments count", func(t *testing.T) {
		f := NewFrecency()
		t1 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		f.Touch("/a", t1)

		last, count := f.Stats("/a")
		if !last.Equal(t1) {
			t.Errorf("LastAccess = %v, want %v", last, t1)
		}
		if count != 1 {
			t.Errorf("count = %d, want 1", count)
		}

		t2 := t1.Add(time.Hour)
		f.Touch("/a", t2)
		last, count = f.Stats("/a")
		if !last.Equal(t2) {
			t.Errorf("LastAccess = %v, want %v", last, t2)
		}
		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
	})
}

func TestFrecencyRecent(t *testing.T) {
	t.Run("returns touched paths ordered by most recently accessed first", func(t *testing.T) {
		f := NewFrecency()
		t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		f.Touch("/oldest", t0)
		f.Touch("/newest", t0.Add(2*time.Hour))
		f.Touch("/middle", t0.Add(time.Hour))

		got := f.Recent(10)
		want := []string{"/newest", "/middle", "/oldest"}
		if len(got) != len(want) {
			t.Fatalf("len(Recent) = %d, want %d: %+v", len(got), len(want), got)
		}
		for i, w := range want {
			if got[i].Path != w {
				t.Errorf("Recent()[%d].Path = %q, want %q", i, got[i].Path, w)
			}
		}
	})

	t.Run("limit caps the number of results", func(t *testing.T) {
		f := NewFrecency()
		now := time.Now()
		f.Touch("/a", now)
		f.Touch("/b", now)
		f.Touch("/c", now)

		if got := f.Recent(2); len(got) != 2 {
			t.Errorf("len(Recent(2)) = %d, want 2", len(got))
		}
	})

	t.Run("an empty tracker returns no results", func(t *testing.T) {
		if got := NewFrecency().Recent(10); len(got) != 0 {
			t.Errorf("Recent() = %+v, want none", got)
		}
	})
}

func TestFrecencySaveLoad(t *testing.T) {
	t.Run("round-trips touched entries through a file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "frecency.json")

		f := NewFrecency()
		at := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		f.Touch("/a", at)
		f.Touch("/a", at.Add(time.Minute))
		f.Touch("/b", at)

		if err := f.Save(path); err != nil {
			t.Fatalf("Save: %v", err)
		}

		loaded, err := LoadFrecency(path)
		if err != nil {
			t.Fatalf("LoadFrecency: %v", err)
		}

		last, count := loaded.Stats("/a")
		if !last.Equal(at.Add(time.Minute)) {
			t.Errorf("/a LastAccess = %v, want %v", last, at.Add(time.Minute))
		}
		if count != 2 {
			t.Errorf("/a count = %d, want 2", count)
		}

		_, count = loaded.Stats("/b")
		if count != 1 {
			t.Errorf("/b count = %d, want 1", count)
		}
	})

	t.Run("loading a missing file returns an empty, ready-to-use Frecency", func(t *testing.T) {
		f, err := LoadFrecency(filepath.Join(t.TempDir(), "missing.json"))
		if err != nil {
			t.Fatalf("LoadFrecency: %v", err)
		}
		_, count := f.Stats("/anything")
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
	})

	t.Run("a corrupt file returns an error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "frecency.json")
		writeFile(t, path, "{not json")

		if _, err := LoadFrecency(path); err == nil {
			t.Fatal("LoadFrecency: got nil error, want non-nil")
		}
	})
}
