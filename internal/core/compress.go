package core

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Compress zips src (a file or directory) into a sibling "<name>.zip",
// disambiguating with a numeric suffix if that name is already taken. It
// returns the path of the zip file created.
func Compress(src string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	base := filepath.Base(src)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]
	zipPath := uniquePath(filepath.Join(filepath.Dir(src), stem+".zip"))

	out, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	w := zip.NewWriter(out)
	if info.IsDir() {
		err = addDirToZip(w, src, base)
	} else {
		err = addFileToZip(w, src, base)
	}
	if err != nil {
		w.Close()
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return zipPath, nil
}

func addDirToZip(w *zip.Writer, dir, prefix string) error {
	des, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, de := range des {
		childPath := filepath.Join(dir, de.Name())
		childName := prefix + "/" + de.Name()
		if de.IsDir() {
			if err := addDirToZip(w, childPath, childName); err != nil {
				return err
			}
			continue
		}
		if err := addFileToZip(w, childPath, childName); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(w *zip.Writer, path, name string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	entry, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, in)
	return err
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s %d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
