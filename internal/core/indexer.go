package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Indexer keeps an Index in sync with disk, scoped to whatever directories
// the caller has actually visited (spec §5.2 follow-up): eagerly walking a
// whole home directory made the index unusable for minutes on a real
// machine (huge trees like ~/Library, macOS TCC-protected folders aborting
// the walk). Instead, Visit indexes one directory's immediate contents on
// demand — GUI navigation and the MCP server's list_directory tool both call
// it — and keeps that directory live via its own fsnotify watch
// (per-directory watches are portable across macOS/Windows/Linux). A
// directory's subdirectories appear in the index as entries (so they're
// findable by name) but their own contents aren't indexed until they're
// separately visited.
type Indexer struct {
	idx  *Index
	fsw  *fsnotify.Watcher
	done chan struct{}
}

// StartIndexer starts an Indexer with an empty index, ready for Visit calls.
// root must exist but is otherwise untouched — nothing is read from it until
// something visits it.
func StartIndexer(root string) (*Indexer, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ix := &Indexer{idx: NewIndex(), fsw: fsw, done: make(chan struct{})}
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

// Visit indexes the immediate (non-recursive) contents of dir and starts
// watching dir for changes, so search covers it from now on. Hidden entries
// (dotfiles) are excluded, matching Finder's default search scope. Safe to
// call again on an already-visited dir — a cheap refresh.
func (ix *Indexer) Visit(dir string) error {
	des, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, de := range des {
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		path := filepath.Join(dir, de.Name())
		ix.idx.Upsert(IndexEntry{Name: de.Name(), Path: path, IsDir: de.IsDir()})
		if !de.IsDir() {
			if info, err := de.Info(); err == nil {
				indexFileContent(ix.idx, path, de.Name(), info.Size())
			}
		}
	}

	return ix.fsw.Add(dir)
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
		if !info.IsDir() {
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
