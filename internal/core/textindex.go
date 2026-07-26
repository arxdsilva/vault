package core

import (
	"os"
	"strings"
)

// indexableTextExt mirrors frontend/src/lib/previewKind.ts's classification
// of files as previewable plain text or code — content search only indexes
// files in one of these buckets (plus extensionless files, e.g. README,
// Makefile), skipping binaries/media where content matching wouldn't be
// meaningful.
var indexableTextExt = map[string]bool{
	"txt": true, "log": true, "cfg": true, "ini": true, "env": true,
	"js": true, "jsx": true, "ts": true, "tsx": true, "go": true, "py": true,
	"rb": true, "java": true, "c": true, "cpp": true, "h": true, "rs": true,
	"json": true, "yaml": true, "yml": true, "toml": true, "md": true,
	"sh": true, "html": true, "css": true, "scss": true, "sql": true,
	"php": true, "swift": true, "kt": true,
}

func isIndexableTextFile(name string) bool {
	i := strings.LastIndex(name, ".")
	if i <= 0 {
		return true
	}
	return indexableTextExt[strings.ToLower(name[i+1:])]
}

// indexFileContent reads path's content into idx for later content search
// (see Search's content-match tier) if it's a small enough indexable text
// file. Best-effort: skips silently (no content stored) on a read error,
// oversized file, or non-text extension, rather than failing the caller's
// walk/watch over one file.
func indexFileContent(idx *Index, path, name string, size int64) {
	if size > DefaultReadFileMaxBytes || !isIndexableTextFile(name) {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	idx.UpsertContent(path, strings.ToLower(string(b)))
}
