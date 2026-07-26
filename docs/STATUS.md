# Status

Tracks progress against the Build Order in `docs/SPEC.md`. Update the checkbox when a step is
functionally complete and validated (build/lint/test green), not merely started.

- [x] 1. Wails v3 + React scaffold, window, theming
- [ ] 2. Go filesystem layer: read dir, metadata, fsnotify watcher, trash, rename/move/copy
- [ ] 3. Virtualized list + breadcrumbs + navigation
- [ ] 4. Search index + frecency ranking + search UI
- [ ] 5. Home screen (Pinned + Recents), pinned rail
- [ ] 6. Command palette + context menu + file operations + undo
- [ ] 7. Quick preview overlay
- [ ] 8. Grid view, thumbnails, polish, empty states, error handling
- [ ] 9. Tests, packaging, README

## Notes

Step 1 done: Wails v3 window (inset title bar, light/dark theme via `prefers-color-scheme`,
indigo accent) + React/TS/Tailwind scaffold with top bar (search input, view toggle, settings
gear). `go vet`, `go build`, and frontend `tsc && vite build` all green. Steps 2–9 not started —
no Go filesystem/search/trash layer yet, no TanStack Virtual or Zustand in `frontend/package.json`.
