package core

import (
	"fmt"
	"os"
)

// DefaultReadFileMaxBytes is the size cap ReadFile enforces when callers
// don't have a more specific limit in mind (e.g. an AI agent reading an
// arbitrary file).
const DefaultReadFileMaxBytes = 200_000

// ReadFile returns path's contents as a string. It errors rather than
// truncating when the file exceeds maxBytes, so a caller can't mistake a
// partial read for the whole file.
func ReadFile(path string, maxBytes int) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}
	if info.Size() > int64(maxBytes) {
		return "", fmt.Errorf("%s is %d bytes, exceeds the %d byte limit", path, info.Size(), maxBytes)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
