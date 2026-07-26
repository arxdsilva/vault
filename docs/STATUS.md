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
  alongside each entry, kept in sync via `Indexer.Visit` and `Indexer.handle` (including on
  `fsnotify.Write`, not just create/remove). `Search` appends a content-match tier (plain
  substring, not fuzzy) after all name matches, deduplicated against them. No API/binding changes
  — `core.Search`'s signature and `SearchResult`/`IndexEntry` shapes are unchanged, so the MCP
  server's `search` tool and the frontend's search both pick this up automatically. `go test -race
  -cover ./internal/core/...` green (88.9% coverage).

### Search bug fix: folder-local matches silently crowded out (spec §5.2)

Root-caused a real, reported bug ("search files I know are in Downloads and get nothing"):
`sahilm/fuzzy` scores *shorter* names higher for an identical prefix match, and `Search` hard-cuts
the ranked candidate list to `limit` (was 50). A folder with many similarly-prefixed short names
(screenshots, `file (1).pdf`, `file (2).pdf`, ...) — Downloads being the textbook case — can
deterministically push a longer, more specific match below the cutoff *every time*, not just
occasionally (reproduced directly in `internal/core/search_test.go` before fixing it).

Fix, now that `wails3` is available in this environment (it wasn't earlier in the session; the
`internal/mcp` read-only design decisions from that period stand, but this one required changing a
bound method):
- `core.Search` gained a `scopeDir` parameter: matches under `scopeDir` get their own allocation of
  up to `limit`, independent of how many higher-scoring matches exist elsewhere in the tree. This
  is a genuine guarantee, not a cosmetic reorder — it operates on the full, untruncated candidate
  list before any cutoff is applied.
- `service.Search.Query` (Wails-bound) gained a matching `dir` parameter; bindings regenerated via
  `wails3 generate bindings` (confirmed working: `wails3 build` also passed end-to-end with the
  real production binding-generation flags). `internal/mcp`'s `search` tool gained an equivalent
  optional `dir` input.
- `frontend/src/hooks/useSearch.ts` now passes the current folder as `dir` and raised
  `RESULT_LIMIT` from 50 to 200 (the scoping guarantee doesn't help when the *noise* is in the same
  folder as the target, which is the more common real case — a bigger budget shrinks that window).
  `SearchResults.tsx` is now virtualized (`@tanstack/react-virtual`, matching `FileList.tsx`) so
  the larger limit stays cheap to render.
- **Residual limitation, stated plainly:** if a single folder has more than ~200 files that all
  score equally against the query (e.g. hundreds of near-identical screenshot names), the specific
  one a user wants can still be excluded — scoping guarantees fairness *between* folders, not
  unbounded recall *within* one. A real fix for that would mean not hard-truncating by score at
  all (pagination, or a fundamentally different ranking), which is a larger change than this pass
  covers. `internal/core/search_test.go` has a regression test documenting this residual case
  explicitly rather than pretending it's fully solved.

### Search redesign: visit-driven indexing, not eager whole-home walk (spec §5.2)

Root-caused a second, worse instance of "search returns nothing" (reported live: searching a file
known to be in `~/Downloads` returned zero results with no error). Debug logging added to
`service.Search.Query` and `useSearch` (`log.Printf`/`console.debug`, prefixed `search:`/
`[search]` — this codebase had no logging anywhere before, and it's the fastest way to confirm the
index/query wiring end-to-end if this class of bug recurs) showed every query being dropped with
"index not ready yet". Two compounding causes, found in order:

1. `core.BuildIndex` walked the whole home directory synchronously before the indexer was
   considered ready, including into `~/Library` (96GB on the reproducing machine: caches, Mail,
   app support, CoreSimulator, ...), which never finished in any reasonable time.
2. Once that was fixed, the walk still failed outright: macOS TCC denies access to
   privacy-protected folders with `EPERM` ("operation not permitted"), not the standard `EACCES`
   that Go's `os.IsPermission` recognizes — so a single protected folder anywhere in the tree
   aborted the *entire* build instead of being skipped, contradicting `BuildIndex`'s own doc
   comment ("unreadable subtrees are skipped rather than failing the whole walk").

Rather than keep patching an eager whole-home walk against an open-ended set of huge/protected
folders, the indexing model changed: **`core.Indexer` only ever indexes directories that have
actually been visited**, not the whole home tree up front.
- `core.BuildIndex` is gone. `Indexer.Visit(dir)` replaces it: a non-recursive `os.ReadDir` of one
  directory, upserted into the index and content-indexed, with a single fsnotify watch added for
  that directory. A subdirectory shown in a visited folder gets an index *entry* (findable by
  name) but its own contents aren't indexed until it's separately visited — `handle`'s Create
  branch no longer recurses a watch onto new subdirectories the way `watchTree` used to.
  `StartIndexer` now just validates `root` exists and returns immediately; the index starts empty.
- `service.Search` gained a Wails-bound `Visit(dir) error` method. `frontend/src/hooks/
  useDirectory.ts` calls `vault.visit(dir)` alongside the existing `vault.watch(dir)` every time
  the browsed folder changes, so search coverage tracks browsing automatically.
- `internal/mcp`'s `list_directory` tool calls `indexer.Visit` too (best-effort, doesn't fail the
  listing) — an MCP agent's `search` tool only covers what it has already listed in the session;
  the tool description says so explicitly so an agent knows to `list_directory` first if `search`
  comes back empty. This was a deliberate tradeoff (discussed and confirmed, not assumed): MCP
  clients have no "browsing session" the way the GUI does, but per-visit indexing was chosen
  uniformly over keeping the old eager-whole-root behavior just for MCP, to avoid two different
  indexing strategies in one codebase.
- The `~/Library`-specific exclusion from the first fix is gone — moot now, since nothing is ever
  walked unless a human or an agent deliberately visits it, and a non-recursive `ReadDir` of even a
  huge directory is cheap regardless of its name. The EPERM-vs-`os.IsPermission` fix is *also*
  effectively moot for the same reason (nothing walks a subtree it wasn't asked to), but the
  underlying `os.IsPermission` gap is real and would resurface if any future code re-adds a
  recursive walk — worth remembering if that happens.
- **Consequence, stated plainly:** search no longer covers "anywhere in your home directory" the
  moment the app launches — only folders you've actually opened (GUI) or listed (MCP) during the
  session. This was an explicit, discussed tradeoff against `docs/SPEC.md`'s original "index of the
  user's home directory" framing, chosen because eager indexing was fundamentally unworkable
  against real home directories (huge folders, OS privacy gates) rather than a preference.

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
- **The search index watcher uses one fsnotify watch per visited directory** (see
  `internal/core/indexer.go`'s `Visit`), not a single native recursive stream (macOS FSEvents /
  Windows `ReadDirectoryChangesW` both support that natively). Since indexing is visit-driven (see
  the redesign note above) rather than a whole-tree walk, watch count now tracks how many folders
  have actually been browsed/listed, not the size of the home directory — the 200k-file/<100ms
  search target in spec §8 is scoped accordingly and hasn't been separately benchmarked.
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
