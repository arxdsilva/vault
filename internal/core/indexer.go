package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Indexer keeps an Index in sync with disk by watching every directory
// under root. Each directory gets its own fsnotify watch (portable across
// macOS/Windows/Linux), and new subdirectories are watched as they appear.
//
// This scales fine for typical home directories; a single-stream recursive
// watch (macOS FSEvents, Windows ReadDirectoryChangesW with the recursive
// flag) would use far fewer OS resources at very large file counts and is
// the natural next step if per-directory watches become a bottleneck.
type Indexer struct {
	root string
	idx  *Index
	fsw  *fsnotify.Watcher
	done chan struct{}
}

// StartIndexer builds an initial index of root and starts watching it for
// changes.
func StartIndexer(root string) (*Indexer, error) {
	idx, err := BuildIndex(root)
	if err != nil {
		return nil, err
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ix := &Indexer{root: root, idx: idx, fsw: fsw, done: make(chan struct{})}
	if err := ix.watchTree(root); err != nil {
		fsw.Close()
		return nil, err
	}

	go ix.run()
	return ix, nil
}

// Index returns the live index. Its Snapshot is safe to call concurrently
// with ongoing updates.
func (ix *Indexer) Index() *Index {
	return ix.idx
}

// Close stops watching and updating the index.
func (ix *Indexer) Close() error {
	close(ix.done)
	return ix.fsw.Close()
}

func (ix *Indexer) watchTree(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return fs.SkipDir
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != dir && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}
		return ix.fsw.Add(path)
	})
}

func (ix *Indexer) run() {
	for {
		select {
		case ev, ok := <-ix.fsw.Events:
			if !ok {
				return
			}
			ix.handle(ev)
		case _, ok := <-ix.fsw.Errors:
			if !ok {
				return
			}
		case <-ix.done:
			return
		}
	}
}

func (ix *Indexer) handle(ev fsnotify.Event) {
	name := filepath.Base(ev.Name)
	if strings.HasPrefix(name, ".") {
		return
	}

	if ev.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		ix.idx.Remove(ev.Name)
		return
	}

	if ev.Op&fsnotify.Create != 0 {
		info, err := os.Stat(ev.Name)
		if err != nil {
			return // vanished before we could stat it
		}
		ix.idx.Upsert(IndexEntry{Name: name, Path: ev.Name, IsDir: info.IsDir()})
		if info.IsDir() {
			ix.watchTree(ev.Name)
		} else {
			indexFileContent(ix.idx, ev.Name, name, info.Size())
		}
		return
	}

	if ev.Op&fsnotify.Write != 0 {
		info, err := os.Stat(ev.Name)
		if err != nil || info.IsDir() {
			return
		}
		indexFileContent(ix.idx, ev.Name, name, info.Size())
	}
}
