package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Pinned is the user's ordered list of pinned locations shown in the Home
// screen and the pinned rail (spec §5.1, §4.3).
type Pinned struct {
	mu    sync.Mutex
	paths []string
}

// NewPinned returns an empty Pinned list.
func NewPinned() *Pinned {
	return &Pinned{}
}

// List returns the pinned paths in display order.
func (p *Pinned) List() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, len(p.paths))
	copy(out, p.paths)
	return out
}

// Add pins path, appending it. Pinning an already-pinned path is a no-op.
func (p *Pinned) Add(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, existing := range p.paths {
		if existing == path {
			return
		}
	}
	p.paths = append(p.paths, path)
}

// Remove unpins path. Removing a path that isn't pinned is a no-op.
func (p *Pinned) Remove(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, existing := range p.paths {
		if existing != path {
			p.paths[n] = existing
			n++
		}
	}
	p.paths = p.paths[:n]
}

// Reorder replaces the pin order with paths (from a drag-drop reorder).
func (p *Pinned) Reorder(paths []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paths = append([]string(nil), paths...)
}

// Save persists the pin list to path as JSON, creating parent directories
// as needed.
func (p *Pinned) Save(path string) error {
	p.mu.Lock()
	paths := append([]string(nil), p.paths...)
	p.mu.Unlock()

	b, err := json.Marshal(paths)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// LoadPinned reads a pin list previously written by Save. A missing file is
// not an error: it returns an empty, ready-to-use Pinned.
func LoadPinned(path string) (*Pinned, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewPinned(), nil
	}
	if err != nil {
		return nil, err
	}

	var paths []string
	if err := json.Unmarshal(b, &paths); err != nil {
		return nil, err
	}
	return &Pinned{paths: paths}, nil
}
