package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPinned(t *testing.T) {
	t.Run("starts empty", func(t *testing.T) {
		if got := NewPinned().List(); len(got) != 0 {
			t.Errorf("List() = %v, want none", got)
		}
	})

	t.Run("Add appends a path", func(t *testing.T) {
		p := NewPinned()
		p.Add("/a")
		p.Add("/b")
		if got := p.List(); len(got) != 2 || got[0] != "/a" || got[1] != "/b" {
			t.Errorf("List() = %v, want [/a /b]", got)
		}
	})

	t.Run("Add is idempotent for an already-pinned path", func(t *testing.T) {
		p := NewPinned()
		p.Add("/a")
		p.Add("/a")
		if got := p.List(); len(got) != 1 {
			t.Errorf("List() = %v, want [/a]", got)
		}
	})

	t.Run("Remove drops a pinned path", func(t *testing.T) {
		p := NewPinned()
		p.Add("/a")
		p.Add("/b")
		p.Remove("/a")
		if got := p.List(); len(got) != 1 || got[0] != "/b" {
			t.Errorf("List() = %v, want [/b]", got)
		}
	})

	t.Run("Remove of an unpinned path is a no-op", func(t *testing.T) {
		p := NewPinned()
		p.Add("/a")
		p.Remove("/missing")
		if got := p.List(); len(got) != 1 || got[0] != "/a" {
			t.Errorf("List() = %v, want [/a]", got)
		}
	})

	t.Run("Reorder replaces the pin order", func(t *testing.T) {
		p := NewPinned()
		p.Add("/a")
		p.Add("/b")
		p.Reorder([]string{"/b", "/a"})
		if got := p.List(); len(got) != 2 || got[0] != "/b" || got[1] != "/a" {
			t.Errorf("List() = %v, want [/b /a]", got)
		}
	})
}

func TestPinnedSaveLoad(t *testing.T) {
	t.Run("round-trips pins through a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		p := NewPinned()
		p.Add("/a")
		p.Add("/b")
		if err := p.Save(path); err != nil {
			t.Fatalf("Save: %v", err)
		}

		loaded, err := LoadPinned(path, nil)
		if err != nil {
			t.Fatalf("LoadPinned: %v", err)
		}
		if got := loaded.List(); len(got) != 2 || got[0] != "/a" || got[1] != "/b" {
			t.Errorf("List() = %v, want [/a /b]", got)
		}
	})

	t.Run("loading a missing file seeds it with the given defaults", func(t *testing.T) {
		p, err := LoadPinned(filepath.Join(t.TempDir(), "missing.json"), []string{"/a", "/b"})
		if err != nil {
			t.Fatalf("LoadPinned: %v", err)
		}
		if got := p.List(); len(got) != 2 || got[0] != "/a" || got[1] != "/b" {
			t.Errorf("List() = %v, want [/a /b]", got)
		}
	})

	t.Run("loading a missing file with no defaults returns an empty, ready-to-use Pinned", func(t *testing.T) {
		p, err := LoadPinned(filepath.Join(t.TempDir(), "missing.json"), nil)
		if err != nil {
			t.Fatalf("LoadPinned: %v", err)
		}
		if got := p.List(); len(got) != 0 {
			t.Errorf("List() = %v, want none", got)
		}
	})

	t.Run("an existing file is not reseeded with defaults", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		writeFile(t, path, `["/a"]`)
		p, err := LoadPinned(path, []string{"/default"})
		if err != nil {
			t.Fatalf("LoadPinned: %v", err)
		}
		if got := p.List(); len(got) != 1 || got[0] != "/a" {
			t.Errorf("List() = %v, want [/a]", got)
		}
	})

	t.Run("a corrupt file returns an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		writeFile(t, path, "{not json")
		if _, err := LoadPinned(path, nil); err == nil {
			t.Fatal("LoadPinned: got nil error, want non-nil")
		}
	})
}

func TestDefaultPins(t *testing.T) {
	t.Run("includes home and its Desktop/Documents/Downloads subfolders that exist", func(t *testing.T) {
		home := t.TempDir()
		for _, sub := range []string{"Desktop", "Documents"} {
			if err := os.Mkdir(filepath.Join(home, sub), 0o755); err != nil {
				t.Fatalf("Mkdir: %v", err)
			}
		}
		// Downloads deliberately omitted to verify missing subfolders are skipped.

		got := DefaultPins(home)
		want := []string{home, filepath.Join(home, "Desktop"), filepath.Join(home, "Documents")}
		if len(got) != len(want) {
			t.Fatalf("DefaultPins() = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("DefaultPins()[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("a non-existent home yields no pins", func(t *testing.T) {
		got := DefaultPins(filepath.Join(t.TempDir(), "missing"))
		if len(got) != 0 {
			t.Errorf("DefaultPins() = %v, want none", got)
		}
	})

	t.Run("includes an extra candidate that exists", func(t *testing.T) {
		home := t.TempDir()
		extra := t.TempDir()

		got := DefaultPins(home, extra)
		found := false
		for _, p := range got {
			if p == extra {
				found = true
			}
		}
		if !found {
			t.Errorf("DefaultPins(%q, %q) = %v, want it to include the extra candidate", home, extra, got)
		}
	})

	t.Run("silently drops an extra candidate that doesn't exist (e.g. an OS with no equivalent)", func(t *testing.T) {
		home := t.TempDir()

		got := DefaultPins(home, "")
		if len(got) != 1 || got[0] != home {
			t.Errorf("DefaultPins(%q, \"\") = %v, want [%q]", home, got, home)
		}
	})
}
