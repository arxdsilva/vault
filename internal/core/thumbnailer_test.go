package core

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func TestThumbnailer(t *testing.T) {
	t.Run("generates a thumbnail via a worker", func(t *testing.T) {
		srcDir := t.TempDir()
		cacheDir := t.TempDir()
		src := filepath.Join(srcDir, "img.png")
		writeTestPNG(t, src, 200, 100)

		th := NewThumbnailer(2)
		defer th.Close()

		out, err := th.Generate(cacheDir, src, 64)
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		w, h := decodePNGSize(t, out)
		if w != 64 || h != 32 {
			t.Errorf("thumbnail size = %dx%d, want 64x32", w, h)
		}
	})

	t.Run("propagates a generation error", func(t *testing.T) {
		th := NewThumbnailer(2)
		defer th.Close()

		_, err := th.Generate(t.TempDir(), filepath.Join(t.TempDir(), "missing.png"), 64)
		if err == nil {
			t.Fatal("Generate: got nil error, want non-nil")
		}
	})

	t.Run("serves many concurrent requests without racing", func(t *testing.T) {
		srcDir := t.TempDir()
		cacheDir := t.TempDir()
		paths := make([]string, 10)
		for i := range paths {
			paths[i] = filepath.Join(srcDir, fmt.Sprintf("img%d.png", i))
			writeTestPNG(t, paths[i], 80, 80)
		}

		th := NewThumbnailer(4)
		defer th.Close()

		var wg sync.WaitGroup
		for _, p := range paths {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				if _, err := th.Generate(cacheDir, p, 32); err != nil {
					t.Errorf("Generate(%q): %v", p, err)
				}
			}(p)
		}
		wg.Wait()
	})
}
