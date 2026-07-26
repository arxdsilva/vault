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
- **Frontend:** React 18 + TypeScript + Tailwind CSS, virtualized lists via TanStack Virtual
- **Search:** Go-side fuzzy matcher over an in-memory index, incremental updates via `fsnotify`
- **Trash:** cross-platform trash/recycle-bin library or platform shell APIs — never permanent delete

## Status

Pre-implementation. See [`docs/STATUS.md`](docs/STATUS.md) for build-order progress.

## Development

- Design spec and non-goals: [`docs/SPEC.md`](docs/SPEC.md)
- Coding conventions and workflow live in `.cursor/rules/` (TDD, core/UI/platform-adapter
  boundaries, commit discipline, response style)
- Commits are prepared by AI tooling but always reviewed and executed by a human
  (`.cursor/rules/git-commits.mdc`)

```sh
wails3 dev       # run in dev mode
wails3 package   # build macOS .dmg / Windows .msi
go build ./...
go vet ./...
golangci-lint run
go test -race -cover ./...
```

## License

TBD
