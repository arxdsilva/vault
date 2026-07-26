package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// CacheThumbnail decodes the image at srcPath, downscales it to fit within
// maxSize x maxSize (preserving aspect ratio, never upscaling), and writes a
// PNG under cacheDir. It returns the cache file's path; a thumbnail already
// cached for this exact source (path + size + mtime + maxSize) is reused
// rather than regenerated.
//
// Video thumbnails aren't covered here — extracting a frame needs a video
// decoder this project doesn't depend on, so the grid falls back to a
// generic icon for video files (see docs/STATUS.md).
func CacheThumbnail(cacheDir, srcPath string, maxSize int) (string, error) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", err
	}

	cachePath, err := thumbnailCachePath(cacheDir, srcPath, info, maxSize)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	src, err := decodeImage(srcPath)
	if err != nil {
		return "", err
	}
	thumb := scaleToFit(src, maxSize)

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	out, err := os.Create(cachePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if err := png.Encode(out, thumb); err != nil {
		return "", err
	}
	return cachePath, nil
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func thumbnailCachePath(cacheDir, srcPath string, info os.FileInfo, maxSize int) (string, error) {
	abs, err := filepath.Abs(srcPath)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf("%s|%d|%d|%d", abs, info.ModTime().UnixNano(), info.Size(), maxSize)
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(cacheDir, hex.EncodeToString(sum[:])+".png"), nil
}

// scaleToFit nearest-neighbor resizes src to fit within maxSize x maxSize,
// preserving aspect ratio. Images already within bounds are left as-is.
func scaleToFit(src image.Image, maxSize int) *image.RGBA {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	scale := 1.0
	if w > maxSize || h > maxSize {
		if w > h {
			scale = float64(maxSize) / float64(w)
		} else {
			scale = float64(maxSize) / float64(h)
		}
	}
	dw := max(1, int(float64(w)*scale))
	dh := max(1, int(float64(h)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		sy := bounds.Min.Y + y*h/dh
		for x := 0; x < dw; x++ {
			sx := bounds.Min.X + x*w/dw
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
