package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func awaitQuery(t *testing.T, s *Search, query string, want string) []string {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		names := queryNames(s, query)
		for _, n := range names {
			if n == want {
				return names
			}
		}
		select {
		case <-time.After(20 * time.Millisecond):
		case <-deadline:
			t.Fatalf("timed out waiting for %q in results of query %q, got %v", want, query, names)
		}
	}
}

func queryNames(s *Search, query string) []string {
	entries := s.Query(query, 10, "")
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return names
}

func TestSearchQuery(t *testing.T) {
	t.Run("finds indexed files by fuzzy query", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "report.txt"), []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()
		if err := s.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		awaitQuery(t, s, "rep", "report.txt")
	})

	t.Run("does not find a file in a directory that was never visited", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "unseen.txt"), []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()

		if got := s.Query("unseen", 10, ""); len(got) != 0 {
			t.Errorf("Query() = %v, want none — root was never Visit()ed", got)
		}
	})

	t.Run("empty query returns no results", func(t *testing.T) {
		root := t.TempDir()
		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()

		if got := s.Query("", 10, ""); len(got) != 0 {
			t.Errorf("Query(\"\") = %v, want none", got)
		}
	})

	t.Run("before startup, Query returns no results rather than panicking", func(t *testing.T) {
		s := NewSearch(t.TempDir(), filepath.Join(t.TempDir(), "frecency.json"))
		if got := s.Query("anything", 10, ""); got != nil && len(got) != 0 {
			t.Errorf("Query before startup = %v, want none", got)
		}
	})

	t.Run("before startup, Visit is a no-op rather than panicking", func(t *testing.T) {
		s := NewSearch(t.TempDir(), filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.Visit(t.TempDir()); err != nil {
			t.Errorf("Visit before startup = %v, want nil", err)
		}
	})

	t.Run("dir scopes results so an in-scope match isn't crowded out by higher-scoring noise elsewhere", func(t *testing.T) {
		root := t.TempDir()
		downloads := filepath.Join(root, "Downloads")
		if err := os.Mkdir(downloads, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		// Many short, higher-scoring names elsewhere in the tree that would
		// otherwise fill up a small shared limit ahead of the target below.
		for _, letter := range []string{"A", "B", "C", "D", "E"} {
			name := filepath.Join(root, "Report"+letter+".txt")
			if err := os.WriteFile(name, []byte("x"), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
		}
		target := filepath.Join(downloads, "Report final version.txt")
		if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()
		if err := s.Visit(root); err != nil {
			t.Fatalf("Visit(root): %v", err)
		}
		if err := s.Visit(downloads); err != nil {
			t.Fatalf("Visit(downloads): %v", err)
		}

		deadline := time.After(2 * time.Second)
		for {
			entries := s.Query("Report", 2, downloads)
			found := false
			for _, e := range entries {
				if e.Path == target {
					found = true
				}
			}
			if found {
				break
			}
			select {
			case <-time.After(20 * time.Millisecond):
			case <-deadline:
				t.Fatalf("timed out waiting for scoped query to include %q", target)
			}
		}
	})
}

func TestSearchTouch(t *testing.T) {
	t.Run("touching a path ranks it above an equally-matching, never-touched path", func(t *testing.T) {
		root := t.TempDir()
		for _, name := range []string{"a_report.txt", "b_report.txt"} {
			if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
		}

		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()
		if err := s.Visit(root); err != nil {
			t.Fatalf("Visit: %v", err)
		}

		awaitQuery(t, s, "report", "a_report.txt")
		s.Touch(filepath.Join(root, "b_report.txt"))

		names := awaitQuery(t, s, "report", "b_report.txt")
		if names[0] != "b_report.txt" {
			t.Errorf("first result = %q, want b_report.txt (touched)", names[0])
		}
	})
}

func TestSearchRecent(t *testing.T) {
	t.Run("returns touched paths most recently accessed first", func(t *testing.T) {
		root := t.TempDir()
		aPath := filepath.Join(root, "a.txt")
		bPath := filepath.Join(root, "b.txt")
		for _, p := range []string{aPath, bPath} {
			if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
		}

		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()

		s.Touch(aPath)
		s.Touch(bPath)

		recent := s.Recent(10)
		if len(recent) != 2 || recent[0].Path != bPath || recent[1].Path != aPath {
			t.Fatalf("Recent = %+v, want [b, a]", recent)
		}
		if recent[0].Name != "b.txt" {
			t.Errorf("recent[0].Name = %q, want b.txt", recent[0].Name)
		}
		if recent[0].IsDir {
			t.Errorf("recent[0].IsDir = true, want false")
		}
	})

	t.Run("skips a touched path that no longer exists", func(t *testing.T) {
		root := t.TempDir()
		gone := filepath.Join(root, "gone.txt")
		if err := os.WriteFile(gone, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		s := NewSearch(root, filepath.Join(t.TempDir(), "frecency.json"))
		if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
			t.Fatalf("ServiceStartup: %v", err)
		}
		defer s.ServiceShutdown()

		s.Touch(gone)
		if err := os.Remove(gone); err != nil {
			t.Fatalf("Remove: %v", err)
		}

		if recent := s.Recent(10); len(recent) != 0 {
			t.Errorf("Recent = %+v, want none", recent)
		}
	})

	t.Run("before startup, Recent returns no results rather than panicking", func(t *testing.T) {
		s := NewSearch(t.TempDir(), filepath.Join(t.TempDir(), "frecency.json"))
		if recent := s.Recent(10); len(recent) != 0 {
			t.Errorf("Recent = %+v, want none", recent)
		}
	})
}

func TestSearchShutdownPersistsFrecency(t *testing.T) {
	root := t.TempDir()
	frecencyPath := filepath.Join(t.TempDir(), "frecency.json")

	s := NewSearch(root, frecencyPath)
	if err := s.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("ServiceStartup: %v", err)
	}
	s.Touch(filepath.Join(root, "a.txt"))

	if err := s.ServiceShutdown(); err != nil {
		t.Fatalf("ServiceShutdown: %v", err)
	}

	if _, err := os.Stat(frecencyPath); err != nil {
		t.Fatalf("expected frecency file to exist: %v", err)
	}
}
