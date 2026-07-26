package service

import (
	"context"
	"sync"

	"github.com/asilva/vault/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Pinned is the Wails service exposing the user's pinned locations (spec
// §5.1 Home screen, §4.3 pinned rail) to the frontend.
type Pinned struct {
	path string

	mu    sync.RWMutex
	store *core.Pinned
}

// NewPinned returns a Pinned persisted at path.
func NewPinned(path string) *Pinned {
	return &Pinned{path: path}
}

// ServiceName names this service for Wails logging/debugging.
func (p *Pinned) ServiceName() string { return "Pinned" }

// ServiceStartup loads the persisted pin list.
func (p *Pinned) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	store, err := core.LoadPinned(p.path)
	if err != nil {
		return err
	}
	p.mu.Lock()
	p.store = store
	p.mu.Unlock()
	return nil
}

// ServiceShutdown persists the pin list.
func (p *Pinned) ServiceShutdown() error {
	store := p.getStore()
	if store == nil {
		return nil
	}
	return store.Save(p.path)
}

func (p *Pinned) getStore() *core.Pinned {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.store
}

// List returns the pinned paths in display order. Always non-nil, so it
// serializes as [] rather than null before the frontend has had a pin list
// to work with.
func (p *Pinned) List() []string {
	store := p.getStore()
	if store == nil {
		return []string{}
	}
	return store.List()
}

// Add pins path and persists the change immediately, so a pin survives a
// crash rather than only being saved at clean shutdown.
func (p *Pinned) Add(path string) error {
	store := p.getStore()
	if store == nil {
		return nil
	}
	store.Add(path)
	return store.Save(p.path)
}

// Remove unpins path and persists the change immediately.
func (p *Pinned) Remove(path string) error {
	store := p.getStore()
	if store == nil {
		return nil
	}
	store.Remove(path)
	return store.Save(p.path)
}

// Reorder replaces the pin order (from a drag-drop reorder) and persists
// the change immediately.
func (p *Pinned) Reorder(paths []string) error {
	store := p.getStore()
	if store == nil {
		return nil
	}
	store.Reorder(paths)
	return store.Save(p.path)
}
