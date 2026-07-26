// Package core holds the platform-neutral file-system logic shared by every
// OS: directory listing, metadata, file operations, and change notification.
// OS-specific behavior (trash semantics, path quirks, watcher backends) stays
// out of this package; see internal/platform/<os>.
package core

import "time"

// Entry describes one file or directory as shown in a listing.
type Entry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}
