package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/arxdsilva/vault/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// VolumesChangeEventName is the Wails event emitted whenever the set of
// mounted volumes changes — a drive plugged in, ejected, or unmounted.
const VolumesChangeEventName = "vault:volumeschange"

// volumesPollInterval is how often Volumes polls for mount/unmount changes.
// Short enough that a plugged-in USB drive feels instant (spec §4.3).
const volumesPollInterval = time.Second

// Volumes is the Wails service exposing mounted volumes (external/USB
// drives, network shares) to the frontend, alongside pinned locations in
// the side panel.
type Volumes struct {
	adapter core.VolumeAdapter
	emitter Emitter
	poll    time.Duration

	mu   sync.Mutex
	stop chan struct{}
	done chan struct{}
}

// NewVolumes returns a Volumes service backed by adapter for the current OS
// and emitter for change notifications.
func NewVolumes(adapter core.VolumeAdapter, emitter Emitter) *Volumes {
	return &Volumes{adapter: adapter, emitter: emitter, poll: volumesPollInterval}
}

// ServiceName names this service for Wails logging/debugging.
func (v *Volumes) ServiceName() string { return "Volumes" }

// ServiceStartup begins polling for mount/unmount changes.
func (v *Volumes) ServiceStartup(_ context.Context, _ application.ServiceOptions) error {
	v.mu.Lock()
	v.stop = make(chan struct{})
	v.done = make(chan struct{})
	stop, done := v.stop, v.done
	v.mu.Unlock()

	// Snapshot synchronously so a mount/unmount that happens right after
	// startup is never missed to a race with the poll goroutine's first read.
	prev := v.snapshot()
	go v.watch(stop, done, prev)
	return nil
}

// ServiceShutdown stops the poll loop.
func (v *Volumes) ServiceShutdown() error {
	v.mu.Lock()
	stop, done := v.stop, v.done
	v.mu.Unlock()
	if stop == nil {
		return nil
	}
	close(stop)
	<-done
	return nil
}

// List returns the currently mounted volumes.
func (v *Volumes) List() ([]core.Volume, error) {
	return v.adapter.List()
}

// Eject unmounts the volume at path.
func (v *Volumes) Eject(path string) error {
	return v.adapter.Eject(path)
}

func (v *Volumes) watch(stop, done chan struct{}, prev string) {
	defer close(done)
	ticker := time.NewTicker(v.poll)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cur := v.snapshot()
			if cur != prev {
				prev = cur
				v.emitter.Emit(VolumesChangeEventName)
			}
		case <-stop:
			return
		}
	}
}

// snapshot returns a comparable fingerprint of the current volume set, so
// watch can detect a mount/unmount without caring what actually changed.
func (v *Volumes) snapshot() string {
	vols, err := v.adapter.List()
	if err != nil {
		return ""
	}
	paths := make([]string, len(vols))
	for i, vol := range vols {
		paths[i] = vol.Path
	}
	sort.Strings(paths)
	return strings.Join(paths, "\x00")
}
