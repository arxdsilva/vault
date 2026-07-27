// Package mcp exposes Vault's file-system core as an MCP (Model Context
// Protocol) server, so external AI clients (Claude Desktop, Cursor, and any
// other MCP-speaking tool) can browse/search/operate on the user's files
// through the same logic the GUI uses — see internal/core. Every path a
// caller supplies is confined to a configured root (resolvePath) since an AI
// agent is a different trust boundary than a human clicking through the GUI.
package mcp

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arxdsilva/vault/internal/core"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server holds the state backing the MCP tool handlers: the root directory
// tools are confined to, the OS trasher, and the search index/frecency
// ranker (kept alive for the process lifetime, mirroring
// internal/service.Search's startup/shutdown but without any Wails
// dependency).
type Server struct {
	root         string
	trasher      core.Trasher
	pinnedPath   string
	frecencyPath string

	indexer  *core.Indexer
	frecency *core.Frecency
}

// NewServer starts the search indexer rooted at root and loads the shared
// frecency ranking data, returning a Server ready to build an MCP server
// from via MCPServer.
func NewServer(root string, trasher core.Trasher, pinnedPath, frecencyPath string) (*Server, error) {
	frecency, err := core.LoadFrecency(frecencyPath)
	if err != nil {
		return nil, err
	}
	indexer, err := core.StartIndexer(root)
	if err != nil {
		return nil, err
	}
	return &Server{
		root:         root,
		trasher:      trasher,
		pinnedPath:   pinnedPath,
		frecencyPath: frecencyPath,
		indexer:      indexer,
		frecency:     frecency,
	}, nil
}

// Close stops the indexer and persists frecency data, mirroring
// internal/service.Search's ServiceShutdown.
func (s *Server) Close() error {
	if err := s.indexer.Close(); err != nil {
		return err
	}
	return s.frecency.Save(s.frecencyPath)
}

// resolvePath resolves requested against root, rejecting any path that
// escapes it (e.g. via "..") so tool handlers never touch anything outside
// the configured root.
func (s *Server) resolvePath(requested string) (string, error) {
	if requested == "" {
		requested = s.root
	}
	abs := requested
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(s.root, abs)
	}
	abs = filepath.Clean(abs)

	rel, err := filepath.Rel(s.root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside the allowed root %q", requested, s.root)
	}
	return abs, nil
}

// MCPServer builds an MCP server with every Vault tool registered.
func (s *Server) MCPServer() *sdkmcp.Server {
	srv := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "vault", Version: "0.1.0"}, nil)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "list_directory",
		Description: "List the entries (files and folders) of a directory.",
	}, s.listDirectory)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "search",
		Description: "Fuzzy-search indexed files by name, ranked by match quality, recency, and frequency of use. Only covers directories already listed via list_directory in this session — call list_directory on the relevant area first if search comes back empty.",
	}, s.search)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "read_file",
		Description: "Read a text file's contents (capped at ~200KB; errors instead of truncating larger files).",
	}, s.readFile)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "move_file",
		Description: "Move a file or folder to a new path.",
	}, s.moveFile)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "rename_file",
		Description: "Rename a file or folder in place.",
	}, s.renameFile)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "paste_into",
		Description: "Copy a file or folder into a destination directory, disambiguating the name if one already exists there (like Finder's paste).",
	}, s.pasteInto)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "new_folder",
		Description: "Create a new folder inside a parent directory.",
	}, s.newFolder)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "delete_file",
		Description: "Move a file or folder to the OS trash. Never permanently deletes.",
	}, s.deleteFile)

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "list_pinned",
		Description: "List the user's pinned folders/files shown in Vault's sidebar.",
	}, s.listPinned)

	return srv
}
