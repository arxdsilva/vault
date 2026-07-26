# Status

Tracks progress against the Build Order in `docs/SPEC.md`. Update the checkbox when a step is
functionally complete and validated (build/lint/test green), not merely started.

- [x] 1. Wails v3 + React scaffold, window, theming
- [x] 2. Go filesystem layer: read dir, metadata, fsnotify watcher, trash, rename/move/copy
- [x] 3. Virtualized list + breadcrumbs + navigation
- [x] 4. Search index + frecency ranking + search UI
- [x] 5. Home screen (Pinned + Recents), pinned rail
- [x] 6. Command palette + context menu + file operations + undo
- [x] 7. Quick preview overlay
- [x] 8. Grid view, thumbnails, polish, empty states, error handling
- [x] 9. Tests, packaging, README
- [x] 10. AI integration: MCP server + embedded chat panel

## Notes

All nine original build-order steps are functionally complete and validated: `go build`/`go vet`/
`go test -race -cover` green (87-89% coverage on `internal/core`, `internal/service`,
`internal/platform/darwin`), frontend `tsc --noEmit`/`vite build`/`vitest run` green (37 tests),
and a real `wails3 package` produced a working signed `bin/vault.app` plus `bin/vault.dmg` (via
`task package:dmg`, `hdiutil`-based) on this macOS machine.

### AI Integration (step 10)

- **MCP server (`cmd/vault-mcp`, `internal/mcp`)** built on the official
  `github.com/modelcontextprotocol/go-sdk`. 9 tools backed directly by `internal/core` (no Wails
  dependency). `go build`/`go vet`/`go test -race -cover` green (80.9% coverage on `internal/mcp`);
  manually smoke-tested end-to-end over stdio with a real MCP client against a scratch directory
  (`list_directory` + `read_file` round-tripped correctly). The `--http` streamable-HTTP transport
  is implemented but not exercised against a real ChatGPT/HTTP MCP client — that ecosystem doesn't
  yet have a common way to point a general MCP-HTTP client at an arbitrary local server the way
  Claude Desktop/Cursor do for stdio.
- **Embedded chat panel** (`frontend/src/components/ChatPanel.tsx`, `lib/chatTools.ts`,
  `lib/llmClient.ts`, `store/chat.ts`) supports Claude (Anthropic Messages API) and ChatGPT (OpenAI
  Chat Completions API), both with the same 9-tool agentic loop. `tsc --noEmit`/`vitest run`/
  `vite build` green (66 tests). Not verified against a real provider round-trip in this
  environment — that needs the user's own API key; the tool-dispatch/UI logic is covered by mocked
  tests, and a real native-window visual check (`make run`) hasn't been performed either.
- Added `internal/platform/trasher_{darwin,windows}.go` (`platform.NewTrasher()`) so the GUI app
  and `vault-mcp` share the same OS trash adapter instead of duplicating build-tagged code.
- Frontend test infra: added `jsdom` as a dev dependency and set `test.environment: "jsdom"` /
  `test.execArgv: ["--no-experimental-webstorage"]` in `vite.config.ts` — Node 22+'s built-in
  (experimental) `localStorage` global otherwise shadows jsdom's working implementation with a
  stub that throws, which broke `store/chat.ts`'s (and would have broken `store/settings.ts`'s)
  `localStorage` persistence under test.

### Content search (spec §5.2 follow-up)

- `internal/core/textindex.go` classifies small text/code files for content indexing (mirrors
  `frontend/src/lib/previewKind.ts`'s extension set); `Index` now stores lowercased content
  alongside each entry, kept in sync via `BuildIndex` and `Indexer.handle` (including on
  `fsnotify.Write`, not just create/remove). `Search` appends a content-match tier (plain
  substring, not fuzzy) after all name matches, deduplicated against them. No API/binding changes
  — `core.Search`'s signature and `SearchResult`/`IndexEntry` shapes are unchanged, so the MCP
  server's `search` tool and the frontend's search both pick this up automatically. `go test -race
  -cover ./internal/core/...` green (88.9% coverage).

### Known limitations / follow-ups

- **Windows adapter is unverified at runtime.** `internal/platform/windows` (Recycle Bin via
  `SHFileOperationW`) only cross-compiles clean (`GOOS=windows go build/vet ./...`) — there's no
  Windows machine in this environment to actually run it. Same caveat applies to the `/preview`
  fileserver route's `RootPath: "/"` with Windows drive-letter paths (`C:\...`); it may need
  adapter-specific path handling.
- **Video thumbnails aren't generated** — the grid view falls back to a generic icon for video
  files. Real frame extraction needs a video decoder (e.g. shelling out to `ffmpeg`), which this
  project doesn't depend on. Image thumbnails are real (decoded, resized, disk-cached via a
  worker pool in `internal/core/thumbnail.go` + `thumbnailer.go`).
- **"Open With…" is disabled** in the command palette/context menu (see
  `frontend/src/lib/fileActions.ts`) — no native "choose an application" picker is wired up.
- **Windows packaging doesn't produce a literal `.msi`.** The scaffolded `windows:package` task
  (NSIS) produces a `.exe` installer; an MSIX task producing `.msix` also exists. A WiX-based
  `.msi` isn't set up. macOS packaging does produce a real `.dmg` (`task package:dmg`).
- **The recursive search index watcher uses one fsnotify watch per directory** (see
  `internal/core/indexer.go`), not a single native recursive stream (macOS FSEvents / Windows
  `ReadDirectoryChangesW` both support that natively). Fine for typical home directories; a
  single-stream backend would use far fewer OS resources at very large file/directory counts —
  the 200k-file/<100ms search target in spec §8 hasn't been benchmarked at that scale.
- **No true progress/cancel UI for long copy/move operations** (spec §5.6 mentions this); errors
  from file operations do surface via a toast (`frontend/src/store/toast.ts`), but there's no
  progress bar or cancel button for large copies.
- The e2e smoke test (`internal/service/smoke_test.go`, launch → search → open → preview) drives
  the real service stack (including the actual Wails `fileserver` service) directly rather than
  through a browser/webview — there's no Wails UI-automation harness set up in this project.
- **The embedded chat panel's API key is stored in `localStorage` in plaintext** — no OS keychain
  integration. Wails v3 has no built-in keychain wrapper; per-OS secure storage is a real addition,
  not done here.
- **MCP tools omit `duplicate`, `compress`, `undo`, `thumbnail`, `watch`, and mutating pins** —
  deliberately out of scope for the initial 9-tool set (list/search/read/move/rename/paste-into/
  new-folder/trash/list-pinned); additive follow-ups if needed.
- **Content search keeps every indexed text file's full content in RAM** (capped at ~200KB/file
  via `core.DefaultReadFileMaxBytes`, and only for text-classified extensions), and content
  matching is an O(indexed text files) substring scan per query rather than a real inverted index.
  Fine for typical home directories; not benchmarked at the 200k-file/<100ms search target in spec
  §8 with a large proportion of text/code files.
