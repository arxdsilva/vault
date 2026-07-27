package core

// Trasher moves a path to the OS trash/recycle bin instead of deleting it
// permanently, and reports where that trash lives so it can be browsed like
// any other folder (a pinned rail entry, spec §4.3). Each OS provides its
// own implementation under internal/platform/<os>; core never deletes a
// file outright.
//
// Location's result is a plain, real directory on macOS (~/.Trash) and
// Linux (XDG: ~/.local/share/Trash/files) — trashed files keep their
// original names there, so normal folder browsing/restore-by-move works
// unmodified. On Windows it's the per-user $Recycle.Bin folder, whose
// entries are opaque, shell-obfuscated names (Explorer's own Recycle Bin
// view resolves them via Shell API this project doesn't implement yet) —
// browsable, but without human-readable names until that's added.
type Trasher interface {
	Trash(path string) error
	Location() (string, error)
}
