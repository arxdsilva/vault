package mcp

import (
	"context"
	"time"

	"github.com/asilva/vault/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

const defaultSearchLimit = 20

type ListDirectoryIn struct {
	Path string `json:"path" jsonschema:"directory to list, absolute or relative to the server root"`
}

type ListDirectoryOut struct {
	Entries []core.Entry `json:"entries"`
}

func (s *Server) listDirectory(_ context.Context, _ *sdkmcp.CallToolRequest, in ListDirectoryIn) (*sdkmcp.CallToolResult, ListDirectoryOut, error) {
	path, err := s.resolvePath(in.Path)
	if err != nil {
		return nil, ListDirectoryOut{}, err
	}
	entries, err := core.ListDir(path)
	if err != nil {
		return nil, ListDirectoryOut{}, err
	}
	return nil, ListDirectoryOut{Entries: entries}, nil
}

type SearchIn struct {
	Query string `json:"query" jsonschema:"fuzzy search query matched against indexed file/folder names"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of results (default 20)"`
}

type SearchOut struct {
	Results []core.SearchResult `json:"results"`
}

func (s *Server) search(_ context.Context, _ *sdkmcp.CallToolRequest, in SearchIn) (*sdkmcp.CallToolResult, SearchOut, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	results := core.Search(s.indexer.Index(), s.frecency, in.Query, time.Now(), limit)
	return nil, SearchOut{Results: results}, nil
}

type ReadFileIn struct {
	Path string `json:"path" jsonschema:"path of the text file to read"`
}

type ReadFileOut struct {
	Content string `json:"content"`
}

func (s *Server) readFile(_ context.Context, _ *sdkmcp.CallToolRequest, in ReadFileIn) (*sdkmcp.CallToolResult, ReadFileOut, error) {
	path, err := s.resolvePath(in.Path)
	if err != nil {
		return nil, ReadFileOut{}, err
	}
	content, err := core.ReadFile(path, core.DefaultReadFileMaxBytes)
	if err != nil {
		return nil, ReadFileOut{}, err
	}
	return nil, ReadFileOut{Content: content}, nil
}

type MoveFileIn struct {
	Src string `json:"src" jsonschema:"path of the file or folder to move"`
	Dst string `json:"dst" jsonschema:"destination path"`
}

type MoveFileOut struct{}

func (s *Server) moveFile(_ context.Context, _ *sdkmcp.CallToolRequest, in MoveFileIn) (*sdkmcp.CallToolResult, MoveFileOut, error) {
	src, err := s.resolvePath(in.Src)
	if err != nil {
		return nil, MoveFileOut{}, err
	}
	dst, err := s.resolvePath(in.Dst)
	if err != nil {
		return nil, MoveFileOut{}, err
	}
	return nil, MoveFileOut{}, core.Move(src, dst)
}

type RenameFileIn struct {
	Path    string `json:"path" jsonschema:"path of the file or folder to rename"`
	NewPath string `json:"newPath" jsonschema:"new path (same parent directory, different name)"`
}

type RenameFileOut struct{}

func (s *Server) renameFile(_ context.Context, _ *sdkmcp.CallToolRequest, in RenameFileIn) (*sdkmcp.CallToolResult, RenameFileOut, error) {
	oldPath, err := s.resolvePath(in.Path)
	if err != nil {
		return nil, RenameFileOut{}, err
	}
	newPath, err := s.resolvePath(in.NewPath)
	if err != nil {
		return nil, RenameFileOut{}, err
	}
	return nil, RenameFileOut{}, core.Rename(oldPath, newPath)
}

type PasteIntoIn struct {
	Src     string `json:"src" jsonschema:"path of the file or folder to copy"`
	DestDir string `json:"destDir" jsonschema:"destination directory"`
}

type PasteIntoOut struct {
	Path string `json:"path" jsonschema:"resulting path of the copy"`
}

func (s *Server) pasteInto(_ context.Context, _ *sdkmcp.CallToolRequest, in PasteIntoIn) (*sdkmcp.CallToolResult, PasteIntoOut, error) {
	src, err := s.resolvePath(in.Src)
	if err != nil {
		return nil, PasteIntoOut{}, err
	}
	destDir, err := s.resolvePath(in.DestDir)
	if err != nil {
		return nil, PasteIntoOut{}, err
	}
	path, err := core.CopyInto(src, destDir)
	if err != nil {
		return nil, PasteIntoOut{}, err
	}
	return nil, PasteIntoOut{Path: path}, nil
}

type NewFolderIn struct {
	Parent string `json:"parent" jsonschema:"parent directory the new folder is created in"`
	Name   string `json:"name" jsonschema:"name of the new folder"`
}

type NewFolderOut struct {
	Path string `json:"path" jsonschema:"path of the folder actually created"`
}

func (s *Server) newFolder(_ context.Context, _ *sdkmcp.CallToolRequest, in NewFolderIn) (*sdkmcp.CallToolResult, NewFolderOut, error) {
	parent, err := s.resolvePath(in.Parent)
	if err != nil {
		return nil, NewFolderOut{}, err
	}
	path, err := core.NewFolder(parent, in.Name)
	if err != nil {
		return nil, NewFolderOut{}, err
	}
	return nil, NewFolderOut{Path: path}, nil
}

type DeleteFileIn struct {
	Path string `json:"path" jsonschema:"path of the file or folder to move to trash"`
}

type DeleteFileOut struct{}

func (s *Server) deleteFile(_ context.Context, _ *sdkmcp.CallToolRequest, in DeleteFileIn) (*sdkmcp.CallToolResult, DeleteFileOut, error) {
	path, err := s.resolvePath(in.Path)
	if err != nil {
		return nil, DeleteFileOut{}, err
	}
	return nil, DeleteFileOut{}, s.trasher.Trash(path)
}

type ListPinnedIn struct{}

type ListPinnedOut struct {
	Paths []string `json:"paths"`
}

func (s *Server) listPinned(_ context.Context, _ *sdkmcp.CallToolRequest, _ ListPinnedIn) (*sdkmcp.CallToolResult, ListPinnedOut, error) {
	pinned, err := core.LoadPinned(s.pinnedPath, nil)
	if err != nil {
		return nil, ListPinnedOut{}, err
	}
	return nil, ListPinnedOut{Paths: pinned.List()}, nil
}
