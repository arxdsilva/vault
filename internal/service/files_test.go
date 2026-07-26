package service

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 200, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
}

type fakeTrasher struct {
	trashed []string
	err     error
}

func (f *fakeTrasher) Trash(path string) error {
	if f.err != nil {
		return f.err
	}
	f.trashed = append(f.trashed, path)
	return nil
}

type emitCall struct {
	name string
	data []any
}

type fakeEmitter struct {
	calls chan emitCall
}

func newFakeEmitter() *fakeEmitter {
	return &fakeEmitter{calls: make(chan emitCall, 16)}
}

func (f *fakeEmitter) Emit(name string, data ...any) bool {
	f.calls <- emitCall{name: name, data: data}
	return true
}

func (f *fakeEmitter) await(t *testing.T) emitCall {
	t.Helper()
	select {
	case c := <-f.calls:
		return c
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for an emitted event")
		return emitCall{}
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %q to exist: %v", path, err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %q to be gone, stat err = %v", path, err)
	}
}

func TestFilesList(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "hi")

	svc := New(&fakeTrasher{}, newFakeEmitter())
	entries, err := svc.List(dir)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "a.txt" {
		t.Errorf("entries = %+v, want a single a.txt entry", entries)
	}
}

func TestFilesStat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	writeFile(t, path, "hi")

	svc := New(&fakeTrasher{}, newFakeEmitter())
	e, err := svc.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if e.Name != "a.txt" || e.IsDir {
		t.Errorf("Stat(%q) = %+v, want a.txt file entry", path, e)
	}
}

func TestFilesHomeDir(t *testing.T) {
	svc := New(&fakeTrasher{}, newFakeEmitter())
	home, err := svc.HomeDir()
	if err != nil {
		t.Fatalf("HomeDir: %v", err)
	}
	if home == "" {
		t.Error("HomeDir: got empty string")
	}
}

func TestFilesRenameMoveCopyNewFolder(t *testing.T) {
	dir := t.TempDir()
	svc := New(&fakeTrasher{}, newFakeEmitter())

	src := filepath.Join(dir, "src.txt")
	writeFile(t, src, "content")

	renamed := filepath.Join(dir, "renamed.txt")
	if err := svc.Rename(src, renamed); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	copied := filepath.Join(dir, "copied.txt")
	if err := svc.Copy(renamed, copied); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	moved := filepath.Join(dir, "moved.txt")
	if err := svc.Move(copied, moved); err != nil {
		t.Fatalf("Move: %v", err)
	}
	if _, err := os.Stat(moved); err != nil {
		t.Errorf("expected %q to exist: %v", moved, err)
	}

	folder, err := svc.NewFolder(dir, "Untitled Folder")
	if err != nil {
		t.Fatalf("NewFolder: %v", err)
	}
	if _, err := os.Stat(folder); err != nil {
		t.Errorf("expected %q to exist: %v", folder, err)
	}
}

func TestFilesPaste(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()
	svc := New(&fakeTrasher{}, newFakeEmitter())
	src := filepath.Join(srcDir, "report.txt")
	writeFile(t, src, "content")

	got, err := svc.Paste(src, destDir)
	if err != nil {
		t.Fatalf("Paste: %v", err)
	}
	want := filepath.Join(destDir, "report.txt")
	if got != want {
		t.Errorf("got = %q, want %q", got, want)
	}
	assertExists(t, src)
	assertExists(t, got)
}

func TestFilesThumbnail(t *testing.T) {
	t.Run("generates a thumbnail for an image", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "photo.png")
		writeTestPNG(t, src, 200, 100)

		svc := New(&fakeTrasher{}, newFakeEmitter())
		svc.cacheDir = t.TempDir()

		out, err := svc.Thumbnail(src)
		if err != nil {
			t.Fatalf("Thumbnail: %v", err)
		}
		assertExists(t, out)
	})

	t.Run("a non-image file returns an error", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "notes.txt")
		writeFile(t, src, "text")

		svc := New(&fakeTrasher{}, newFakeEmitter())
		svc.cacheDir = t.TempDir()

		if _, err := svc.Thumbnail(src); err == nil {
			t.Fatal("Thumbnail: got nil error, want non-nil")
		}
	})
}

func TestFilesDuplicate(t *testing.T) {
	dir := t.TempDir()
	svc := New(&fakeTrasher{}, newFakeEmitter())
	src := filepath.Join(dir, "report.txt")
	writeFile(t, src, "content")

	dup, err := svc.Duplicate(src)
	if err != nil {
		t.Fatalf("Duplicate: %v", err)
	}
	if dup != filepath.Join(dir, "report copy.txt") {
		t.Errorf("dup = %q, want %q", dup, filepath.Join(dir, "report copy.txt"))
	}
	assertExists(t, src)
	assertExists(t, dup)
}

func TestFilesCompress(t *testing.T) {
	dir := t.TempDir()
	svc := New(&fakeTrasher{}, newFakeEmitter())
	src := filepath.Join(dir, "report.txt")
	writeFile(t, src, "content")

	zipPath, err := svc.Compress(src)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if zipPath != filepath.Join(dir, "report.zip") {
		t.Errorf("zipPath = %q, want %q", zipPath, filepath.Join(dir, "report.zip"))
	}
	assertExists(t, zipPath)
}

func TestFilesUndo(t *testing.T) {
	t.Run("undoes a rename", func(t *testing.T) {
		dir := t.TempDir()
		svc := New(&fakeTrasher{}, newFakeEmitter())
		original := filepath.Join(dir, "original.txt")
		renamed := filepath.Join(dir, "renamed.txt")
		writeFile(t, original, "x")

		if err := svc.Rename(original, renamed); err != nil {
			t.Fatalf("Rename: %v", err)
		}
		if err := svc.Undo(); err != nil {
			t.Fatalf("Undo: %v", err)
		}
		assertExists(t, original)
		assertMissing(t, renamed)
	})

	t.Run("undoes a move", func(t *testing.T) {
		dir := t.TempDir()
		svc := New(&fakeTrasher{}, newFakeEmitter())
		src := filepath.Join(dir, "src.txt")
		dst := filepath.Join(dir, "dst.txt")
		writeFile(t, src, "x")

		if err := svc.Move(src, dst); err != nil {
			t.Fatalf("Move: %v", err)
		}
		if err := svc.Undo(); err != nil {
			t.Fatalf("Undo: %v", err)
		}
		assertExists(t, src)
		assertMissing(t, dst)
	})

	t.Run("undoes multiple operations in last-in-first-out order", func(t *testing.T) {
		dir := t.TempDir()
		svc := New(&fakeTrasher{}, newFakeEmitter())
		a := filepath.Join(dir, "a.txt")
		b := filepath.Join(dir, "b.txt")
		c := filepath.Join(dir, "c.txt")
		writeFile(t, a, "x")

		if err := svc.Rename(a, b); err != nil {
			t.Fatalf("Rename: %v", err)
		}
		if err := svc.Rename(b, c); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		if err := svc.Undo(); err != nil {
			t.Fatalf("Undo: %v", err)
		}
		assertExists(t, b)
		assertMissing(t, c)

		if err := svc.Undo(); err != nil {
			t.Fatalf("Undo: %v", err)
		}
		assertExists(t, a)
		assertMissing(t, b)
	})

	t.Run("Copy, NewFolder and Delete are not undoable", func(t *testing.T) {
		dir := t.TempDir()
		trasher := &fakeTrasher{}
		svc := New(trasher, newFakeEmitter())
		src := filepath.Join(dir, "src.txt")
		writeFile(t, src, "x")

		if _, err := svc.NewFolder(dir, "Sub"); err != nil {
			t.Fatalf("NewFolder: %v", err)
		}
		if err := svc.Copy(src, filepath.Join(dir, "copy.txt")); err != nil {
			t.Fatalf("Copy: %v", err)
		}
		if err := svc.Delete(src); err != nil {
			t.Fatalf("Delete: %v", err)
		}

		if err := svc.Undo(); err == nil {
			t.Fatal("Undo: got nil error, want non-nil (nothing undoable was recorded)")
		}
	})

	t.Run("an empty history returns an error rather than panicking", func(t *testing.T) {
		svc := New(&fakeTrasher{}, newFakeEmitter())
		if err := svc.Undo(); err == nil {
			t.Fatal("Undo: got nil error, want non-nil")
		}
	})

	t.Run("a failed Rename is not recorded for undo", func(t *testing.T) {
		dir := t.TempDir()
		svc := New(&fakeTrasher{}, newFakeEmitter())
		if err := svc.Rename(filepath.Join(dir, "missing.txt"), filepath.Join(dir, "new.txt")); err == nil {
			t.Fatal("Rename: got nil error, want non-nil")
		}
		if err := svc.Undo(); err == nil {
			t.Fatal("Undo: got nil error, want non-nil (the failed rename should not be undoable)")
		}
	})
}

func TestFilesDelete(t *testing.T) {
	t.Run("delegates to the injected trasher", func(t *testing.T) {
		trasher := &fakeTrasher{}
		svc := New(trasher, newFakeEmitter())

		if err := svc.Delete("/some/path.txt"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if len(trasher.trashed) != 1 || trasher.trashed[0] != "/some/path.txt" {
			t.Errorf("trashed = %v, want [/some/path.txt]", trasher.trashed)
		}
	})

	t.Run("propagates the trasher's error", func(t *testing.T) {
		trasher := &fakeTrasher{err: fmt.Errorf("boom")}
		svc := New(trasher, newFakeEmitter())

		if err := svc.Delete("/some/path.txt"); err == nil {
			t.Fatal("Delete: got nil error, want non-nil")
		}
	})
}

func TestFilesWatch(t *testing.T) {
	t.Run("emits an event when the watched directory changes", func(t *testing.T) {
		dir := t.TempDir()
		emitter := newFakeEmitter()
		svc := New(&fakeTrasher{}, emitter)

		if err := svc.Watch(dir); err != nil {
			t.Fatalf("Watch: %v", err)
		}
		defer svc.Unwatch()

		path := filepath.Join(dir, "new.txt")
		writeFile(t, path, "hi")

		call := emitter.await(t)
		if call.name != ChangeEventName {
			t.Errorf("event name = %q, want %q", call.name, ChangeEventName)
		}
		if len(call.data) != 1 {
			t.Fatalf("event data = %v, want 1 payload", call.data)
		}
		change, ok := call.data[0].(ChangeEvent)
		if !ok {
			t.Fatalf("payload type = %T, want ChangeEvent", call.data[0])
		}
		if change.Dir != dir {
			t.Errorf("change.Dir = %q, want %q", change.Dir, dir)
		}
		if filepath.Clean(change.Path) != filepath.Clean(path) {
			t.Errorf("change.Path = %q, want %q", change.Path, path)
		}
	})

	t.Run("watching a new directory replaces the previous watch", func(t *testing.T) {
		dirA := t.TempDir()
		dirB := t.TempDir()
		emitter := newFakeEmitter()
		svc := New(&fakeTrasher{}, emitter)

		if err := svc.Watch(dirA); err != nil {
			t.Fatalf("Watch(dirA): %v", err)
		}
		if err := svc.Watch(dirB); err != nil {
			t.Fatalf("Watch(dirB): %v", err)
		}
		defer svc.Unwatch()

		writeFile(t, filepath.Join(dirB, "new.txt"), "hi")
		call := emitter.await(t)
		if call.data[0].(ChangeEvent).Dir != dirB {
			t.Errorf("expected the event to come from dirB, got %+v", call.data[0])
		}

		writeFile(t, filepath.Join(dirA, "new.txt"), "hi")
		select {
		case c := <-emitter.calls:
			t.Fatalf("expected no more events after switching watches, got %+v", c)
		case <-time.After(300 * time.Millisecond):
		}
	})

	t.Run("Unwatch stops delivering events", func(t *testing.T) {
		dir := t.TempDir()
		emitter := newFakeEmitter()
		svc := New(&fakeTrasher{}, emitter)

		if err := svc.Watch(dir); err != nil {
			t.Fatalf("Watch: %v", err)
		}
		svc.Unwatch()

		writeFile(t, filepath.Join(dir, "new.txt"), "hi")
		select {
		case c := <-emitter.calls:
			t.Fatalf("expected no events after Unwatch, got %+v", c)
		case <-time.After(300 * time.Millisecond):
		}
	})

	t.Run("missing directory returns an error", func(t *testing.T) {
		svc := New(&fakeTrasher{}, newFakeEmitter())
		if err := svc.Watch(filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Fatal("Watch: got nil error, want non-nil")
		}
	})
}
