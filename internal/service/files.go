// Package service holds the thin, Wails-bound service structs exposed to
// the frontend. They call only into internal/core, never a platform
// adapter directly, so the UI/adapter boundary stays enforced at this layer.
package service

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/asilva/vault/internal/core"
)

// thumbnailMaxSize is the max edge length (px) generated thumbnails are
// scaled to fit within — 2x a ~80px grid tile for HiDPI screens.
const thumbnailMaxSize = 160

// thumbnailWorkers bounds concurrent thumbnail decode/resize work.
const thumbnailWorkers = 4

// Emitter delivers a named event with a payload to the frontend. Satisfied
// by *application.EventManager (app.Event) in production.
type Emitter interface {
	Emit(name string, data ...any) bool
}

// ChangeEventName is the Wails event emitted for every filesystem change in
// the currently watched directory.
const ChangeEventName = "vault:fschange"

// ChangeEvent is the payload delivered with ChangeEventName.
type ChangeEvent struct {
	Dir  string `json:"dir"`
	Path string `json:"path"`
	Op   string `json:"op"`
}

// undoOp records one reversible operation (rename or move). Copy, NewFolder
// and Delete are deliberately not undoable (spec §5.6): Delete already goes
// through the OS trash, and Copy/NewFolder create rather than relocate.
type undoOp struct {
	from, to string // the operation moved/renamed from -> to
}

// Files is the Wails service exposing file-system operations to the
// frontend.
type Files struct {
	trash   core.Trasher
	emitter Emitter

	mu      sync.Mutex
	watcher *core.Watcher

	undoMu    sync.Mutex
	undoStack []undoOp

	thumbOnce   sync.Once
	thumbnailer *core.Thumbnailer
	cacheDir    string // empty means "resolve the default lazily" (see thumbnailCacheDir)
}

// New returns a Files service backed by trash for delete operations and
// emitter for change notifications.
func New(trash core.Trasher, emitter Emitter) *Files {
	return &Files{trash: trash, emitter: emitter}
}

// ServiceName names this service for Wails logging/debugging.
func (f *Files) ServiceName() string { return "Files" }

// ServiceShutdown stops any active watch/thumbnail workers when the app exits.
func (f *Files) ServiceShutdown() error {
	f.Unwatch()
	if f.thumbnailer != nil {
		f.thumbnailer.Close()
	}
	return nil
}

// Thumbnail returns the path of a cached, downscaled thumbnail for the
// image at path, generating one via a bounded worker pool if needed (spec
// §5.3 grid view, §3 async thumbnail generation). Non-image files (video
// included — see core.CacheThumbnail) return an error; the frontend falls
// back to a generic icon.
func (f *Files) Thumbnail(path string) (string, error) {
	f.thumbOnce.Do(func() {
		f.thumbnailer = core.NewThumbnailer(thumbnailWorkers)
		if f.cacheDir == "" {
			f.cacheDir = thumbnailCacheDir()
		}
	})
	if f.cacheDir == "" {
		return "", errors.New("no thumbnail cache directory available")
	}
	return f.thumbnailer.Generate(f.cacheDir, path, thumbnailMaxSize)
}

func thumbnailCacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "vault", "thumbnails")
}

// List returns the entries directly inside dir.
func (f *Files) List(dir string) ([]core.Entry, error) {
	return core.ListDir(dir)
}

// HomeDir returns the current user's home directory.
func (f *Files) HomeDir() (string, error) {
	return os.UserHomeDir()
}

// Rename renames oldPath to newPath. The rename can be reversed with Undo.
func (f *Files) Rename(oldPath, newPath string) error {
	if err := core.Rename(oldPath, newPath); err != nil {
		return err
	}
	f.pushUndo(undoOp{from: oldPath, to: newPath})
	return nil
}

// Move relocates src to dst. The move can be reversed with Undo.
func (f *Files) Move(src, dst string) error {
	if err := core.Move(src, dst); err != nil {
		return err
	}
	f.pushUndo(undoOp{from: src, to: dst})
	return nil
}

// Copy copies src to dst.
func (f *Files) Copy(src, dst string) error {
	return core.Copy(src, dst)
}

// NewFolder creates a new folder named name inside parent and returns the
// path actually created.
func (f *Files) NewFolder(parent, name string) (string, error) {
	return core.NewFolder(parent, name)
}

// Duplicate copies path into a disambiguated sibling ("name copy.ext") and
// returns the path created.
func (f *Files) Duplicate(path string) (string, error) {
	return core.Duplicate(path)
}

// Paste copies src into destDir, disambiguating a name collision, and
// returns the path created. Used for clipboard-style copy/paste.
func (f *Files) Paste(src, destDir string) (string, error) {
	return core.CopyInto(src, destDir)
}

// Compress zips path into a sibling "<name>.zip" and returns the path
// created.
func (f *Files) Compress(path string) (string, error) {
	return core.Compress(path)
}

// Delete moves path to the OS trash/recycle bin.
func (f *Files) Delete(path string) error {
	return f.trash.Trash(path)
}

func (f *Files) pushUndo(op undoOp) {
	f.undoMu.Lock()
	defer f.undoMu.Unlock()
	f.undoStack = append(f.undoStack, op)
}

// Undo reverses the most recent rename or move, in last-in-first-out order.
// It returns an error if there is nothing left to undo.
func (f *Files) Undo() error {
	f.undoMu.Lock()
	if len(f.undoStack) == 0 {
		f.undoMu.Unlock()
		return errors.New("nothing to undo")
	}
	op := f.undoStack[len(f.undoStack)-1]
	f.undoStack = f.undoStack[:len(f.undoStack)-1]
	f.undoMu.Unlock()

	return core.Move(op.to, op.from)
}

// Watch starts watching dir for changes, replacing any previous watch.
// Changes are delivered as ChangeEventName events until Watch is called
// again with a different directory or Unwatch is called.
func (f *Files) Watch(dir string) error {
	w, err := core.NewWatcher(dir)
	if err != nil {
		return err
	}

	f.mu.Lock()
	prev := f.watcher
	f.watcher = w
	f.mu.Unlock()

	if prev != nil {
		prev.Close()
	}

	go f.forward(w, dir)
	return nil
}

// Unwatch stops the active watch, if any.
func (f *Files) Unwatch() {
	f.mu.Lock()
	w := f.watcher
	f.watcher = nil
	f.mu.Unlock()

	if w != nil {
		w.Close()
	}
}

func (f *Files) forward(w *core.Watcher, dir string) {
	for change := range w.Events {
		f.emitter.Emit(ChangeEventName, ChangeEvent{
			Dir:  dir,
			Path: change.Path,
			Op:   opName(change.Op),
		})
	}
}

func opName(op core.ChangeOp) string {
	switch op {
	case core.Created:
		return "created"
	case core.Removed:
		return "removed"
	case core.Renamed:
		return "renamed"
	default:
		return "modified"
	}
}
