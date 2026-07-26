package mcp

import (
	"path/filepath"
	"testing"

	"github.com/asilva/vault/internal/core"
)

type fakeTrasher struct {
	trashed []string
}

func (f *fakeTrasher) Trash(path string) error {
	f.trashed = append(f.trashed, path)
	return nil
}

// newTestServer starts a Server rooted at a fresh temp dir, cleaned up
// automatically via t.Cleanup, so tests never touch the host's real
// file system (per this repo's testing rule).
func newTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	s, err := NewServer(root, &fakeTrasher{}, filepath.Join(root, ".vault", "pinned.json"), filepath.Join(root, ".vault", "frecency.json"))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return s, root
}

func TestResolvePath(t *testing.T) {
	s, root := newTestServer(t)

	t.Run("an empty path resolves to root", func(t *testing.T) {
		got, err := s.resolvePath("")
		if err != nil {
			t.Fatalf("resolvePath: %v", err)
		}
		if got != root {
			t.Errorf("resolvePath(\"\") = %q, want %q", got, root)
		}
	})

	t.Run("a relative path resolves inside root", func(t *testing.T) {
		got, err := s.resolvePath("sub/file.txt")
		if err != nil {
			t.Fatalf("resolvePath: %v", err)
		}
		want := filepath.Join(root, "sub", "file.txt")
		if got != want {
			t.Errorf("resolvePath() = %q, want %q", got, want)
		}
	})

	t.Run("an absolute path inside root is accepted", func(t *testing.T) {
		want := filepath.Join(root, "file.txt")
		got, err := s.resolvePath(want)
		if err != nil {
			t.Fatalf("resolvePath: %v", err)
		}
		if got != want {
			t.Errorf("resolvePath() = %q, want %q", got, want)
		}
	})

	t.Run("a relative path escaping root via .. is rejected", func(t *testing.T) {
		if _, err := s.resolvePath("../outside"); err == nil {
			t.Fatal("resolvePath: got nil error, want non-nil")
		}
	})

	t.Run("an absolute path outside root is rejected", func(t *testing.T) {
		if _, err := s.resolvePath(filepath.Join(t.TempDir(), "elsewhere")); err == nil {
			t.Fatal("resolvePath: got nil error, want non-nil")
		}
	})
}

func TestNewServerAndClose(t *testing.T) {
	root := t.TempDir()
	s, err := NewServer(root, &fakeTrasher{}, filepath.Join(root, "pinned.json"), filepath.Join(root, "frecency.json"))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if s.MCPServer() == nil {
		t.Fatal("MCPServer() = nil")
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

var _ core.Trasher = (*fakeTrasher)(nil)
