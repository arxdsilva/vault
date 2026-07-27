package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/arxdsilva/vault/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type fakeVolumeAdapter struct {
	mu       sync.Mutex
	vols     []core.Volume
	listErr  error
	ejected  []string
	ejectErr error
}

func (f *fakeVolumeAdapter) List() ([]core.Volume, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]core.Volume, len(f.vols))
	copy(out, f.vols)
	return out, nil
}

func (f *fakeVolumeAdapter) Eject(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ejectErr != nil {
		return f.ejectErr
	}
	f.ejected = append(f.ejected, path)
	return nil
}

func (f *fakeVolumeAdapter) setVolumes(vols []core.Volume) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.vols = vols
}

func TestVolumesList(t *testing.T) {
	adapter := &fakeVolumeAdapter{vols: []core.Volume{{Path: "/Volumes/USB", Name: "USB"}}}
	svc := NewVolumes(adapter, newFakeEmitter())

	got, err := svc.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].Path != "/Volumes/USB" {
		t.Errorf("List() = %+v, want the adapter's volume", got)
	}
}

func TestVolumesListError(t *testing.T) {
	adapter := &fakeVolumeAdapter{listErr: errors.New("boom")}
	svc := NewVolumes(adapter, newFakeEmitter())

	if _, err := svc.List(); err == nil {
		t.Fatal("List: want an error, got nil")
	}
}

func TestVolumesEject(t *testing.T) {
	adapter := &fakeVolumeAdapter{}
	svc := NewVolumes(adapter, newFakeEmitter())

	if err := svc.Eject("/Volumes/USB"); err != nil {
		t.Fatalf("Eject: %v", err)
	}
	if len(adapter.ejected) != 1 || adapter.ejected[0] != "/Volumes/USB" {
		t.Errorf("adapter.ejected = %v, want [/Volumes/USB]", adapter.ejected)
	}
}

func TestVolumesEjectError(t *testing.T) {
	adapter := &fakeVolumeAdapter{ejectErr: errors.New("busy")}
	svc := NewVolumes(adapter, newFakeEmitter())

	if err := svc.Eject("/Volumes/USB"); err == nil {
		t.Fatal("Eject: want an error, got nil")
	}
}

func TestVolumesWatchEmitsOnMountAndUnmount(t *testing.T) {
	adapter := &fakeVolumeAdapter{}
	emitter := newFakeEmitter()
	svc := NewVolumes(adapter, emitter)
	svc.poll = 5 * time.Millisecond

	if err := svc.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("ServiceStartup: %v", err)
	}
	defer func() {
		if err := svc.ServiceShutdown(); err != nil {
			t.Errorf("ServiceShutdown: %v", err)
		}
	}()

	// Plug in a drive: the poll loop should notice and emit a change event.
	adapter.setVolumes([]core.Volume{{Path: "/Volumes/USB", Name: "USB"}})
	if got := emitter.await(t); got.name != VolumesChangeEventName {
		t.Errorf("event = %q, want %q", got.name, VolumesChangeEventName)
	}

	// Unplug it: another change event.
	adapter.setVolumes(nil)
	if got := emitter.await(t); got.name != VolumesChangeEventName {
		t.Errorf("event = %q, want %q", got.name, VolumesChangeEventName)
	}
}

func TestVolumesWatchStopsOnShutdown(t *testing.T) {
	adapter := &fakeVolumeAdapter{}
	svc := NewVolumes(adapter, newFakeEmitter())
	svc.poll = 5 * time.Millisecond

	if err := svc.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("ServiceStartup: %v", err)
	}

	done := make(chan struct{})
	go func() {
		if err := svc.ServiceShutdown(); err != nil {
			t.Errorf("ServiceShutdown: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ServiceShutdown did not return; watch loop likely still running")
	}
}
