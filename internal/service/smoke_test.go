package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asilva/vault/internal/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/fileserver"
)

// TestSmokeLaunchSearchOpenPreview exercises the real service stack across
// the flow spec §9 calls out for an e2e smoke test: launch (list a
// directory) → browse into a subfolder → search → open (list the found
// result's directory) → preview (stream its bytes, as the frontend's
// <img>/<video>/... would). Search only covers folders that have actually
// been browsed to (spec §5.2 follow-up), so each List call here is paired
// with a Visit, mirroring what the frontend does on navigation. It stands
// in for a full browser-driven e2e test, which would need a Wails-specific
// UI automation harness this project doesn't set up.
func TestSmokeLaunchSearchOpenPreview(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "Photos"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	photoPath := filepath.Join(home, "Photos", "vacation-sunset.png")
	writeTestPNG(t, photoPath, 64, 32)
	writeFile(t, filepath.Join(home, "notes.txt"), "unrelated")

	search := NewSearch(home, filepath.Join(t.TempDir(), "frecency.json"))
	if err := search.ServiceStartup(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatalf("Search.ServiceStartup: %v", err)
	}
	defer search.ServiceShutdown()

	// Launch: the home directory lists its immediate contents.
	files := New(&fakeTrasher{}, newFakeEmitter())
	launchEntries, err := files.List(home)
	if err != nil {
		t.Fatalf("List(home): %v", err)
	}
	if !hasName(launchEntries, "Photos") {
		t.Fatalf("launch listing missing Photos: %+v", launchEntries)
	}
	if err := search.Visit(home); err != nil {
		t.Fatalf("Visit(home): %v", err)
	}

	// Browse into Photos — the photo only becomes searchable once its own
	// folder has been visited, not just its ancestor.
	photosDir := filepath.Join(home, "Photos")
	if _, err := files.List(photosDir); err != nil {
		t.Fatalf("List(photosDir): %v", err)
	}
	if err := search.Visit(photosDir); err != nil {
		t.Fatalf("Visit(photosDir): %v", err)
	}

	// Search: find the photo by a fuzzy fragment of its name.
	foundPath := awaitSearchResult(t, search, "sunset", "vacation-sunset.png")
	search.Touch(foundPath)

	// Open: reveal the result by listing its parent directory.
	parent := filepath.Dir(foundPath)
	openEntries, err := files.List(parent)
	if err != nil {
		t.Fatalf("List(parent): %v", err)
	}
	if !hasName(openEntries, "vacation-sunset.png") {
		t.Fatalf("open listing missing the photo: %+v", openEntries)
	}

	// Preview: the frontend streams file bytes from the fileserver service
	// mounted at /preview (see main.go) via <img src>.
	preview := fileserver.NewWithConfig(&fileserver.Config{RootPath: "/"})
	req := httptest.NewRequest(http.MethodGet, previewRequestPath(foundPath), nil)
	rec := httptest.NewRecorder()
	preview.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	want, err := os.ReadFile(foundPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if rec.Body.String() != string(want) {
		t.Error("preview response body doesn't match the file's actual bytes")
	}
}

// previewRequestPath mirrors the frontend's previewUrl: each path segment is
// percent-encoded independently so the separators survive.
func previewRequestPath(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return "/" + strings.Join(segments, "/")
}

func hasName(entries []core.Entry, name string) bool {
	for _, e := range entries {
		if e.Name == name {
			return true
		}
	}
	return false
}

func awaitSearchResult(t *testing.T, search *Search, query, wantName string) string {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		for _, e := range search.Query(query, 10, "") {
			if e.Name == wantName {
				return e.Path
			}
		}
		select {
		case <-time.After(20 * time.Millisecond):
		case <-deadline:
			t.Fatalf("timed out waiting for %q in search results for %q", wantName, query)
		}
	}
}
