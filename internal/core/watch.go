package core

import "github.com/fsnotify/fsnotify"

// ChangeOp classifies a filesystem change delivered by a Watcher.
type ChangeOp int

const (
	Modified ChangeOp = iota
	Created
	Removed
	Renamed
)

// Change describes a single filesystem event.
type Change struct {
	Path string
	Op   ChangeOp
}

// Watcher watches a single directory (non-recursively) for changes. Events
// and Errors are delivered on their respective channels from a dedicated
// goroutine started by NewWatcher; call Close to stop it.
type Watcher struct {
	fsw    *fsnotify.Watcher
	Events chan Change
	Errors chan error
	done   chan struct{}
}

// NewWatcher starts watching dir and returns once the watch is registered.
func NewWatcher(dir string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return nil, err
	}

	w := &Watcher{
		fsw:    fsw,
		Events: make(chan Change),
		Errors: make(chan error),
		done:   make(chan struct{}),
	}
	go w.run()
	return w, nil
}

func (w *Watcher) run() {
	defer close(w.Events)
	defer close(w.Errors)
	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return
			}
			select {
			case w.Events <- Change{Path: ev.Name, Op: mapOp(ev.Op)}:
			case <-w.done:
				return
			}
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return
			}
			select {
			case w.Errors <- err:
			case <-w.done:
				return
			}
		case <-w.done:
			return
		}
	}
}

func mapOp(op fsnotify.Op) ChangeOp {
	switch {
	case op&fsnotify.Create != 0:
		return Created
	case op&fsnotify.Remove != 0:
		return Removed
	case op&fsnotify.Rename != 0:
		return Renamed
	default:
		return Modified
	}
}

// Close stops the watcher and its goroutine. Events and Errors are closed
// once the goroutine exits.
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsw.Close()
}
