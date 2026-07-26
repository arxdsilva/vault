package core

// Trasher moves a path to the OS trash/recycle bin instead of deleting it
// permanently. Each OS provides its own implementation under
// internal/platform/<os>; core never deletes a file outright.
type Trasher interface {
	Trash(path string) error
}
