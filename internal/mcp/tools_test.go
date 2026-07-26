package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestListDirectory(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "a.txt"), "a")
	writeFile(t, filepath.Join(root, "b.txt"), "b")

	_, out, err := s.listDirectory(context.Background(), nil, ListDirectoryIn{Path: ""})
	if err != nil {
		t.Fatalf("listDirectory: %v", err)
	}
	if len(out.Entries) != 2 {
		t.Errorf("Entries = %v, want 2 entries", out.Entries)
	}
}

func TestListDirectoryRejectsPathOutsideRoot(t *testing.T) {
	s, _ := newTestServer(t)
	if _, _, err := s.listDirectory(context.Background(), nil, ListDirectoryIn{Path: "../escape"}); err == nil {
		t.Fatal("listDirectory: got nil error, want non-nil")
	}
}

func TestReadFileTool(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "note.txt"), "hello")

	_, out, err := s.readFile(context.Background(), nil, ReadFileIn{Path: "note.txt"})
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}
	if out.Content != "hello" {
		t.Errorf("Content = %q, want %q", out.Content, "hello")
	}
}

func TestMoveFile(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "src.txt"), "x")

	if _, _, err := s.moveFile(context.Background(), nil, MoveFileIn{Src: "src.txt", Dst: "dst.txt"}); err != nil {
		t.Fatalf("moveFile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "dst.txt")); err != nil {
		t.Errorf("expected dst.txt to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "src.txt")); !os.IsNotExist(err) {
		t.Errorf("expected src.txt to be gone, stat err = %v", err)
	}
}

func TestRenameFile(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "old.txt"), "x")

	if _, _, err := s.renameFile(context.Background(), nil, RenameFileIn{Path: "old.txt", NewPath: "new.txt"}); err != nil {
		t.Fatalf("renameFile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "new.txt")); err != nil {
		t.Errorf("expected new.txt to exist: %v", err)
	}
}

func TestPasteInto(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "src.txt"), "x")
	if err := os.Mkdir(filepath.Join(root, "dest"), 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	_, out, err := s.pasteInto(context.Background(), nil, PasteIntoIn{Src: "src.txt", DestDir: "dest"})
	if err != nil {
		t.Fatalf("pasteInto: %v", err)
	}
	if _, err := os.Stat(out.Path); err != nil {
		t.Errorf("expected %q to exist: %v", out.Path, err)
	}
}

func TestNewFolderTool(t *testing.T) {
	s, root := newTestServer(t)

	_, out, err := s.newFolder(context.Background(), nil, NewFolderIn{Parent: "", Name: "Untitled"})
	if err != nil {
		t.Fatalf("newFolder: %v", err)
	}
	if out.Path != filepath.Join(root, "Untitled") {
		t.Errorf("Path = %q, want %q", out.Path, filepath.Join(root, "Untitled"))
	}
}

func TestDeleteFile(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "gone.txt"), "x")

	if _, _, err := s.deleteFile(context.Background(), nil, DeleteFileIn{Path: "gone.txt"}); err != nil {
		t.Fatalf("deleteFile: %v", err)
	}
	trasher := s.trasher.(*fakeTrasher)
	if len(trasher.trashed) != 1 || trasher.trashed[0] != filepath.Join(root, "gone.txt") {
		t.Errorf("trashed = %v, want [%s]", trasher.trashed, filepath.Join(root, "gone.txt"))
	}
}

func TestListPinnedTool(t *testing.T) {
	root := t.TempDir()
	pinnedPath := filepath.Join(root, "pinned.json")
	writeFile(t, pinnedPath, `["/a", "/b"]`)

	s, err := NewServer(root, &fakeTrasher{}, pinnedPath, filepath.Join(root, "frecency.json"))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	_, out, err := s.listPinned(context.Background(), nil, ListPinnedIn{})
	if err != nil {
		t.Fatalf("listPinned: %v", err)
	}
	if len(out.Paths) != 2 || out.Paths[0] != "/a" || out.Paths[1] != "/b" {
		t.Errorf("Paths = %v, want [/a /b]", out.Paths)
	}
}

func TestSearchTool(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "report.txt"), "x")

	// Search only covers directories that have been listed (spec §5.2
	// follow-up); list_directory indexes synchronously, so no polling needed.
	if _, _, err := s.listDirectory(context.Background(), nil, ListDirectoryIn{Path: ""}); err != nil {
		t.Fatalf("listDirectory: %v", err)
	}

	_, out, err := s.search(context.Background(), nil, SearchIn{Query: "report"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(out.Results) == 0 {
		t.Fatal("search: report.txt not found")
	}
}

func TestSearchToolFindsNothingBeforeListing(t *testing.T) {
	s, root := newTestServer(t)
	writeFile(t, filepath.Join(root, "report.txt"), "x")

	_, out, err := s.search(context.Background(), nil, SearchIn{Query: "report"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(out.Results) != 0 {
		t.Errorf("search before any list_directory call = %+v, want none", out.Results)
	}
}

func TestSearchToolWithDir(t *testing.T) {
	s, root := newTestServer(t)
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	writeFile(t, filepath.Join(sub, "report.txt"), "x")

	// An agent exploring the tree lists root, then drills into sub — both
	// listings index their own contents (spec §5.2 follow-up).
	if _, _, err := s.listDirectory(context.Background(), nil, ListDirectoryIn{Path: ""}); err != nil {
		t.Fatalf("listDirectory(root): %v", err)
	}
	if _, _, err := s.listDirectory(context.Background(), nil, ListDirectoryIn{Path: "sub"}); err != nil {
		t.Fatalf("listDirectory(sub): %v", err)
	}

	_, out, err := s.search(context.Background(), nil, SearchIn{Query: "report", Dir: "sub"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(out.Results) == 0 || out.Results[0].Entry.Path != filepath.Join(sub, "report.txt") {
		t.Fatalf("results = %+v, want report.txt under sub", out.Results)
	}
}

func TestSearchToolRejectsDirOutsideRoot(t *testing.T) {
	s, _ := newTestServer(t)
	if _, _, err := s.search(context.Background(), nil, SearchIn{Query: "report", Dir: "../escape"}); err == nil {
		t.Fatal("search: got nil error, want non-nil")
	}
}
