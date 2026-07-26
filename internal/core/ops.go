package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Rename renames oldPath to newPath in place. It refuses to overwrite an
// existing destination.
func Rename(oldPath, newPath string) error {
	if err := refuseExisting(newPath); err != nil {
		return err
	}
	return os.Rename(oldPath, newPath)
}

// Move relocates src to dst, falling back to copy-then-remove when a direct
// rename isn't possible (e.g. moving across volumes). It refuses to
// overwrite an existing destination.
//
// The copy-then-remove fallback only triggers for cross-volume moves, which
// aren't reproducible with a single t.TempDir(); it's exercised by Copy's
// own tests instead.
func Move(src, dst string) error {
	if err := refuseExisting(dst); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := Copy(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// Copy copies src to dst. Directories are copied recursively.
func Copy(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst, info)
	}
	return copyFile(src, dst, info)
}

func refuseExisting(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", filepath.Base(path))
	}
	return nil
}

func copyDir(src, dst string, info os.FileInfo) error {
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	des, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, de := range des {
		srcChild := filepath.Join(src, de.Name())
		dstChild := filepath.Join(dst, de.Name())
		childInfo, err := de.Info()
		if err != nil {
			return err
		}
		if de.IsDir() {
			if err := copyDir(srcChild, dstChild, childInfo); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(srcChild, dstChild, childInfo); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, info os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
