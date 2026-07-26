package core

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
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

func decodePNGSize(t *testing.T, path string) (int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%q): %v", path, err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		t.Fatalf("DecodeConfig: %v", err)
	}
	return cfg.Width, cfg.Height
}

func TestCacheThumbnail(t *testing.T) {
	t.Run("downscales a large image to fit maxSize", func(t *testing.T) {
		srcDir := t.TempDir()
		cacheDir := t.TempDir()
		src := filepath.Join(srcDir, "big.png")
		writeTestPNG(t, src, 400, 200)

		out, err := CacheThumbnail(cacheDir, src, 100)
		if err != nil {
			t.Fatalf("CacheThumbnail: %v", err)
		}
		w, h := decodePNGSize(t, out)
		if w != 100 || h != 50 {
			t.Errorf("thumbnail size = %dx%d, want 100x50", w, h)
		}
	})

	t.Run("does not upscale an image already smaller than maxSize", func(t *testing.T) {
		srcDir := t.TempDir()
		cacheDir := t.TempDir()
		src := filepath.Join(srcDir, "small.png")
		writeTestPNG(t, src, 40, 20)

		out, err := CacheThumbnail(cacheDir, src, 100)
		if err != nil {
			t.Fatalf("CacheThumbnail: %v", err)
		}
		w, h := decodePNGSize(t, out)
		if w != 40 || h != 20 {
			t.Errorf("thumbnail size = %dx%d, want 40x20 (unscaled)", w, h)
		}
	})

	t.Run("a second call for the same source reuses the cached file", func(t *testing.T) {
		srcDir := t.TempDir()
		cacheDir := t.TempDir()
		src := filepath.Join(srcDir, "img.png")
		writeTestPNG(t, src, 200, 200)

		first, err := CacheThumbnail(cacheDir, src, 64)
		if err != nil {
			t.Fatalf("CacheThumbnail: %v", err)
		}
		second, err := CacheThumbnail(cacheDir, src, 64)
		if err != nil {
			t.Fatalf("CacheThumbnail: %v", err)
		}
		if first != second {
			t.Errorf("cache path changed between calls: %q vs %q", first, second)
		}
	})

	t.Run("a changed source regenerates rather than reusing a stale cache", func(t *testing.T) {
		srcDir := t.TempDir()
		cacheDir := t.TempDir()
		src := filepath.Join(srcDir, "img.png")
		writeTestPNG(t, src, 200, 100)

		first, err := CacheThumbnail(cacheDir, src, 64)
		if err != nil {
			t.Fatalf("CacheThumbnail: %v", err)
		}

		// Overwrite with different content (different dimensions, so a
		// different encoded size) and a fresh mtime.
		writeTestPNG(t, src, 100, 200)

		second, err := CacheThumbnail(cacheDir, src, 64)
		if err != nil {
			t.Fatalf("CacheThumbnail: %v", err)
		}
		if first == second {
			t.Error("expected a new cache path after the source changed")
		}
		w, h := decodePNGSize(t, second)
		if w != 32 || h != 64 {
			t.Errorf("regenerated thumbnail size = %dx%d, want 32x64", w, h)
		}
	})

	t.Run("missing source returns an error", func(t *testing.T) {
		_, err := CacheThumbnail(t.TempDir(), filepath.Join(t.TempDir(), "missing.png"), 64)
		if err == nil {
			t.Fatal("CacheThumbnail: got nil error, want non-nil")
		}
	})

	t.Run("a non-image file returns a decode error", func(t *testing.T) {
		srcDir := t.TempDir()
		src := filepath.Join(srcDir, "notes.txt")
		writeFile(t, src, "just text")

		_, err := CacheThumbnail(t.TempDir(), src, 64)
		if err == nil {
			t.Fatal("CacheThumbnail: got nil error, want non-nil")
		}
	})
}
