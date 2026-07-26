// Command vault-mcp exposes Vault's file-system core as an MCP server, so
// Claude Desktop, Cursor, or any other MCP client can browse/search/operate
// on the user's files through the same logic the GUI app uses. See
// README.md for client setup (Claude Desktop / Cursor config snippets).
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/asilva/vault/internal/mcp"
	"github.com/asilva/vault/internal/platform"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	root := flag.String("root", home, "restrict file operations to this directory tree")
	httpAddr := flag.String("http", "", "serve over streamable HTTP at this address (e.g. :8090) instead of stdio")
	flag.Parse()

	pinnedPath := filepath.Join(configDir, "vault", "pinned.json")
	frecencyPath := filepath.Join(configDir, "vault", "frecency.json")

	server, err := mcp.NewServer(*root, platform.NewTrasher(), pinnedPath, frecencyPath)
	if err != nil {
		log.Fatal(err)
	}
	mcpServer := server.MCPServer()

	var runErr error
	if *httpAddr != "" {
		handler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server { return mcpServer }, nil)
		log.Printf("vault-mcp listening on %s (root %s)", *httpAddr, *root)
		runErr = http.ListenAndServe(*httpAddr, handler)
	} else {
		runErr = mcpServer.Run(context.Background(), &sdkmcp.StdioTransport{})
	}

	if err := server.Close(); err != nil {
		log.Printf("vault-mcp: error closing: %v", err)
	}
	if runErr != nil {
		log.Fatal(runErr)
	}
}
