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

## Notes

All nine build-order steps are functionally complete and validated: `go build`/`go vet`/
`go test -race -cover` green (87-89% coverage on `internal/core`, `internal/service`,
`internal/platform/darwin`), frontend `tsc --noEmit`/`vite build`/`vitest run` green (37 tests),
and a real `wails3 package` produced a working signed `bin/vault.app` plus `bin/vault.dmg` (via
`task package:dmg`, `hdiutil`-based) on this macOS machine.

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
