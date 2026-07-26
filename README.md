# Vault

A minimal, cross-platform file manager aiming to replace Finder (macOS) and Windows File
Explorer — fast, lightweight, and free of feature bloat.

A file manager that behaves like a command palette, not a filing cabinet: search-first navigation,
zero chrome, keyboard-first, instant everything. Full design spec: [`docs/SPEC.md`](docs/SPEC.md).

## Goals

- Fast startup, low resource usage
- Keyboard-first navigation, full mouse support
- Cross-platform: macOS and Windows at minimum (Linux as a stretch goal)
- Clean, minimal UI — no unnecessary chrome or feature creep
- Feels native on each platform it targets

## Stack

- **Shell:** Wails v3 (Go backend + native webview)
- **Frontend:** React 18 + TypeScript + Tailwind CSS, virtualized lists/grid via TanStack Virtual,
  Zustand for UI state, Shiki for code preview syntax highlighting
- **Search:** Go-side fuzzy matcher (`sahilm/fuzzy`) over an in-memory index of the home
  directory, ranked by match quality → recency → frequency; incremental updates via `fsnotify`
- **Trash:** OS-native trash/recycle-bin per platform adapter — never a permanent delete
- **Thumbnails:** image thumbnails decoded/resized/cached on disk by a bounded Go worker pool

## Architecture

```
main.go                    Wails app entry: window, theming, service registration
internal/core/             platform-neutral: dir listing, file ops, search index, frecency,
                            fsnotify watchers, thumbnail generation — no OS-specific APIs
internal/platform/darwin/  macOS adapter: Trash (~/.Trash)
internal/platform/windows/ Windows adapter: Recycle Bin (SHFileOperationW) — cross-compiles
                            clean but is unverified at runtime (no Windows box in dev/CI here)
internal/service/          thin Wails-bound services (Files, Search, Pinned) exposed to the
                            frontend; call only into internal/core
frontend/                  React + TypeScript + Tailwind UI; talks to internal/service via
                            generated Wails bindings (frontend/bindings/)
```

See [`.cursor/rules/architecture.mdc`](.cursor/rules/architecture.mdc) for the full boundary
rules (core stays UI-agnostic; OS-specific behavior stays in one adapter per OS).

## Keyboard map

| Key | Action |
|-----|--------|
| Type anywhere | Search (top bar is always focused on launch) |
| `Esc` | Clear search, return to Home |
| `↑` `↓` | Move selection in a list/search results |
| `←` `→` `↑` `↓` | Move selection in grid view |
| `Enter` | Open selection |
| `Cmd/Ctrl` + `Enter` (in search) | Reveal result's folder in-app |
| `Cmd/Ctrl` + `↑` | Go to parent folder |
| `Space` | Quick preview (image/video/audio/PDF/text/code) |
| `←` `→` `↑` `↓` (preview open) | Move to previous/next file |
| `Cmd/Ctrl` + `K` | Command palette |
| `F2` | Rename selection |
| `Cmd/Ctrl` + `N` | New folder |
| `Cmd/Ctrl` + `C` (via palette "Copy") / `Cmd/Ctrl` + `V` | Copy / paste |
| `Cmd/Ctrl` + `Z` | Undo (rename/move only) |
| Drag file/folder onto the pinned rail | Pin it; drag within the rail to reorder |

## Non-goals (MVP)

No tabs, split panes, cloud integrations, network drives, bulk metadata editing, plugins, or file
tagging. See `docs/SPEC.md` §7 for the full list.

## Status

All nine build-order steps are implemented and validated. See
[`docs/STATUS.md`](docs/STATUS.md) for the checklist and a list of known limitations (Windows
runtime verification, video thumbnails, "Open With…", `.msi` packaging).

## Development

- Design spec and non-goals: [`docs/SPEC.md`](docs/SPEC.md)
- Coding conventions and workflow live in `.cursor/rules/` (TDD, core/UI/platform-adapter
  boundaries, commit discipline, response style)
- Commits are prepared by AI tooling but always reviewed and executed by a human
  (`.cursor/rules/git-commits.mdc`)

### Setup

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.117  # once, if wails3 isn't on PATH
cd frontend && npm install && cd ..
```

### Run

```sh
wails3 dev                       # dev mode, hot reload
wails3 build && ./bin/vault      # production build, run directly
```

### Package

```sh
wails3 package                   # bin/vault.app (ad-hoc signed)
wails3 task package:dmg          # bin/vault.dmg (macOS only, via hdiutil)
wails3 task windows:package GOOS=windows   # NSIS .exe installer (cross-compiled; see docs/STATUS.md)
```

### Test

```sh
go build ./...
go vet ./...
go test -race -cover ./...        # internal/core, internal/service, internal/platform/darwin
GOOS=windows go build ./... && GOOS=windows go vet ./...   # windows adapter compile check

cd frontend
npx tsc --noEmit
npx vitest run
npm run build
```

`golangci-lint` isn't installed in this environment; run `golangci-lint run` if you have it
available locally.

## License

TBD
