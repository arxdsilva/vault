# Build Spec: "Vault" — A Minimal, Modern File Manager

You are an AI coding agent. Build a cross-platform desktop file manager (macOS + Windows) that replaces Finder/Explorer for everyday use. Optimize for simplicity, speed, and a modern feel. Follow this spec closely; where details are unspecified, choose the simplest option that preserves the design principles.

## 1. Product Vision

A file manager that behaves like a command palette, not a filing cabinet. Users find files by searching and recency, not by clicking through folder trees. One window, minimal chrome, keyboard-first, instant everything.

## 2. Design Principles (non-negotiable)

1. **Search-first.** The search bar is the primary navigation. Browsing is secondary.
2. **Zero clutter.** No toolbars, no status bars, no folder trees. Actions appear contextually.
3. **Instant.** Every interaction under 100ms perceived latency. Virtualize all lists. Never block the UI on filesystem I/O.
4. **Keyboard-first.** Every action must be reachable without a mouse.
5. **Calm aesthetics.** Generous whitespace, a single accent color, subtle 150–200ms transitions, native light/dark mode following the OS.

## 3. Tech Stack

- **Shell:** Wails v3 (Go backend + native webview) — small binary, native performance, fast iteration.
- **Frontend:** React 18 + TypeScript + Tailwind CSS.
- **List virtualization:** TanStack Virtual.
- **Search index:** Go-side fuzzy matcher (e.g., `github.com/sahilm/fuzzy` or `github.com/junegunn/fzf/src/algo`) over an in-memory index of the user's home directory; incremental updates via `github.com/fsnotify/fsnotify` watcher running in its own goroutine.
- **Concurrency:** goroutines + worker pools for indexing, thumbnail generation, and file operations; communicate progress to the frontend via Wails events.
- **Trash support:** a cross-platform Go trash library (e.g., `github.com/hymkor/trash-go`) or platform shell APIs — never permanent delete.
- **State:** Zustand or React context — keep it simple.
- **Icons:** Lucide.

## 4. Layout

Single window, three zones:

1. **Top bar (56px):** search input, full width, always focused on launch. Right side: view toggle (list/grid) and settings gear only.
2. **Path strip (36px):** clickable breadcrumbs (`Home › Projects › ClientX`). Hidden on the launch/search screen, visible when browsing a folder.
3. **Content area:** virtualized file list (default) or grid. Left edge: a slim 48px rail of pinned locations (icons only, tooltip on hover, user-configurable via drag-drop).

No sidebar tree. No menu bar beyond the OS-required minimum.

## 5. Core Screens & Behaviors

### 5.1 Launch state ("Home")
- Search bar focused, empty.
- Below: two sections — **Pinned** (user-chosen folders/files, grid of large tiles) and **Recent** (last 20 opened/modified files, list with relative timestamps like "2h ago").

### 5.2 Search
- Typing filters instantly (debounce ≤ 50ms) across the indexed home directory.
- Fuzzy matching, ranked by: match quality → recency of access → frequency of access (frecency).
- Results show name, parent path (dimmed), and file-type icon. Enter opens; Cmd/Ctrl+Enter reveals in its folder inside the app.
- Esc clears search and returns to Home.

### 5.3 Browsing
- Clicking a folder (or Enter on it) navigates into it; breadcrumbs update.
- Sort: name / modified / size / kind, via a small inline control; remember per-folder.
- List rows: icon, name, modified date, size. Grid tiles: thumbnail + name.
- Thumbnails for images/video generated async in Go worker goroutines, cached on disk.

### 5.4 Quick preview
- Spacebar on selection opens an overlay preview: images, video (with playback), PDF, plain text, and code (syntax-highlighted via Shiki), audio.
- Arrow keys move between files while preview stays open. Esc closes.

### 5.5 Command palette
- Cmd/Ctrl+K opens an action palette scoped to the current selection: Open, Open With…, Rename, Move to… (with fuzzy folder picker), Copy, Duplicate, Copy Path, New Folder, Compress, Move to Trash, Pin.
- All actions also available via right-click context menu with the same items.

### 5.6 File operations
- Rename inline (F2 or via palette).
- Drag-and-drop move/copy (Alt/Option for copy), including drag out to other apps and drop in from OS.
- Multi-select: Shift/Cmd/Ctrl click, and Cmd/Ctrl+A.
- Deletions go to the OS trash — never permanent delete in the MVP.
- Show a non-blocking progress toast for long copies/moves, with cancel.
- Undo (Cmd/Ctrl+Z) for rename and move.

### 5.7 Keyboard map (minimum)
- Arrow keys: navigate. Enter: open. Cmd/Ctrl+↑: go to parent. Space: preview.
- Cmd/Ctrl+K: palette. Cmd/Ctrl+F or just typing: search. F2: rename. Cmd/Ctrl+N: new folder.

## 6. Visual Design

- Typeface: Inter (or system UI stack). Base size 14px, generous 1.5 line height in lists (row height ~40px).
- Light and dark themes keyed to OS setting; accent color default indigo, user-changeable among 6 presets.
- Rounded 8px corners, soft shadows only on overlays, no borders where spacing suffices.
- Motion: fade/slide 150ms ease-out for overlays and navigation; no bouncy or decorative animation.
- Empty states: friendly single line + icon, never blank white.

## 7. Non-Goals (MVP)

- No tabs, no split panes, no cloud integrations, no network drives, no bulk metadata editing, no general third-party plugin system, no file tagging. Keep it out; note them as future ideas in the README. ("No plugin system" does not exclude the specific AI-integration surface described in §10 — that's a defined, first-party feature, not an open plugin API.)

## 8. Performance Targets

- Cold start to interactive: < 1.5s.
- Search results render: < 100ms after keystroke on an index of 200k files.
- Scroll at 60fps in folders with 50k entries (virtualized).
- Idle RAM: < 200MB.

## 9. Quality & Delivery

- TypeScript strict mode; Go code passes `go vet` and `golangci-lint` with a sensible default config; race detector clean (`go test -race`).
- Unit tests for the frecency ranker and path utilities; at least one e2e smoke test (launch → search → open → preview).
- Handle permission errors gracefully (macOS full-disk access prompt with a clear explainer screen).
- Deliverables: runnable dev setup (`wails3 dev`), build instructions for macOS `.dmg` and Windows `.msi` (`wails3 package`), and a README covering architecture, keyboard map, and non-goals.

## 10. AI Integration

Vault is meant to be an AI-first replacement for Finder/Explorer, not just a faster one: an AI
assistant should be able to browse, search, and operate on the user's files the same way a human
does through the GUI. Two surfaces share the same 9 underlying operations (list directory, search,
read file, move, rename, paste-into, new folder, trash, list pinned):

- **MCP server (`cmd/vault-mcp`).** A standalone Go binary exposing Vault's file-system core
  (`internal/core`, via `internal/mcp`) as an MCP (Model Context Protocol) server, so Claude
  Desktop, Cursor, or any other MCP client can drive Vault. Serves over stdio by default (what
  Claude Desktop and Cursor expect) or `--http` for a local streamable-HTTP transport, which keeps
  the door open for ChatGPT or other HTTP-based MCP clients as that ecosystem matures. See
  README.md for client setup.
- **Embedded chat panel.** A docked, bring-your-own-key assistant inside Vault itself
  (`frontend/src/components/ChatPanel.tsx`), supporting both Claude and ChatGPT. Tool execution
  reuses the same Wails-bound calls the rest of the UI already uses (`frontend/src/lib/vault.ts`)
  rather than adding a second backend surface.
- **Safety default:** both surfaces confine file operations to the user's home directory by
  default (path-traversal-checked) rather than the whole filesystem, since an AI agent is a
  different trust boundary than a human clicking through the GUI. The MCP server's `--root` flag
  overrides this for power users.
- **Known limitation:** the chat panel's API key is stored in `localStorage` in plaintext — no OS
  keychain integration yet (see docs/STATUS.md).

## 11. Build Order

1. Wails v3 + React scaffold, window, theming.
2. Go filesystem layer: read dir, metadata, fsnotify watcher, trash, rename/move/copy.
3. Virtualized list + breadcrumbs + navigation.
4. Search index + frecency ranking + search UI.
5. Home screen (Pinned + Recents), pinned rail.
6. Command palette + context menu + file operations + undo.
7. Quick preview overlay.
8. Grid view, thumbnails, polish, empty states, error handling.
9. Tests, packaging, README.
10. AI integration: MCP server + embedded chat panel.
