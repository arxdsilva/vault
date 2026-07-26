package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func startPinned(t *testing.T, path string) *Pinned {
	t.Helper()
	p := NewPinned(path)
	if err := p.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("ServiceStartup: %v", err)
	}
	return p
}

func TestPinnedService(t *testing.T) {
	t.Run("List returns an empty slice, never null/nil, before ServiceStartup", func(t *testing.T) {
		p := NewPinned(filepath.Join(t.TempDir(), "pinned.json"))
		got := p.List()
		if got == nil {
			t.Fatal("List() = nil, want a non-nil empty slice (JSON must serialize as [], not null)")
		}
		if len(got) != 0 {
			t.Errorf("List() = %v, want none", got)
		}
	})

	t.Run("starts empty and lists added pins", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		p := startPinned(t, path)

		if got := p.List(); len(got) != 0 {
			t.Fatalf("List() = %v, want none", got)
		}

		if err := p.Add("/a"); err != nil {
			t.Fatalf("Add: %v", err)
		}
		if got := p.List(); len(got) != 1 || got[0] != "/a" {
			t.Errorf("List() = %v, want [/a]", got)
		}
	})

	t.Run("Add persists immediately, before shutdown", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		p := startPinned(t, path)

		if err := p.Add("/a"); err != nil {
			t.Fatalf("Add: %v", err)
		}

		reloaded := NewPinned(path)
		if err := reloaded.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		if got := reloaded.List(); len(got) != 1 || got[0] != "/a" {
			t.Errorf("reloaded List() = %v, want [/a]", got)
		}
	})

	t.Run("Remove drops a pin and persists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		p := startPinned(t, path)
		if err := p.Add("/a"); err != nil {
			t.Fatalf("Add: %v", err)
		}

		if err := p.Remove("/a"); err != nil {
			t.Fatalf("Remove: %v", err)
		}
		if got := p.List(); len(got) != 0 {
			t.Errorf("List() = %v, want none", got)
		}
	})

	t.Run("Reorder replaces the pin order and persists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		p := startPinned(t, path)
		if err := p.Add("/a"); err != nil {
			t.Fatalf("Add: %v", err)
		}
		if err := p.Add("/b"); err != nil {
			t.Fatalf("Add: %v", err)
		}

		if err := p.Reorder([]string{"/b", "/a"}); err != nil {
			t.Fatalf("Reorder: %v", err)
		}
		if got := p.List(); len(got) != 2 || got[0] != "/b" || got[1] != "/a" {
			t.Errorf("List() = %v, want [/b /a]", got)
		}
	})

	t.Run("ServiceShutdown persists the current pin list", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "pinned.json")
		p := startPinned(t, path)
		if err := p.Add("/a"); err != nil {
			t.Fatalf("Add: %v", err)
		}
		if err := p.ServiceShutdown(); err != nil {
			t.Fatalf("ServiceShutdown: %v", err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected pinned file to exist: %v", err)
		}
	})
}
