# Performance

Measured against the four targets in `docs/SPEC.md` §8, on an Apple M3 Pro (macOS 26), production
build. Results: `docs/ss-vault-performance.svg` (embedded in the README). Raw numbers below for
reproducibility — re-run these yourself before trusting a stale chart.

| Target (spec §8) | Result | Status |
|---|---|---|
| Cold start to interactive: < 1.5s | 353ms avg (330–417ms, 5 trials) | Pass |
| Search results render: < 100ms @ 200k files | 39ms typical query · 84ms broad query · 101ms broad query scoped to a folder | Mostly pass — see caveat below |
| Scroll at 60fps @ 50k entries | 120fps avg, 62.9fps worst single frame, 0 frames over the 60fps budget | Pass |
| Idle RAM: < 200MB | ~80MB, 40s after launch | Pass |

## Caveat: search latency

The 101ms figure is the backend `core.Search` call only — a broad query (matches every
synthetic entry) scoped to a subfolder, the worst case the ranking/scoping logic handles. It
excludes the frontend's 50ms debounce (`useSearch.ts`) and IPC/render overhead, so real
perceived latency for that specific worst case is higher than 101ms. A query that's actually
narrowed by a few typed characters (the realistic case — nobody searches a single common
substring like "report" and stops typing) is 39ms, comfortably inside the target even after
debounce+render. Not currently optimized further; tracked as a known limitation rather than
silently rounded down.

## How each number was produced

### Search latency

Go benchmark against `internal/core.Search` directly (no UI, no IPC) — `internal/core/search_bench_test.go`,
committed and re-runnable:

```sh
go test ./internal/core/ -run '^$' -bench 'BenchmarkSearch200k' -benchtime=10x -count=5
```

Builds a synthetic 200,000-entry index shaped like a real home directory (5 top-level folders,
100 subfolders each, 10 file extensions) and times three query shapes: a narrow/specific query,
a broad unscoped query, and a broad query scoped to one folder (`Search`'s `scopeDir` — see that
function's doc comment).

### Idle RAM

Built `bin/vault.app` via `wails3 task darwin:package:dmg` (production build, ad-hoc signed),
launched it, and sampled `ps -o rss=` on its PID every 5s for 40s once idle:

```sh
open bin/vault.app
# then, for the running PID:
for i in $(seq 0 5 40); do ps -o rss= -p $PID; sleep 5; done
```

### Cold start

Same production binary, run directly from a terminal (so stdout is visible) rather than via
`open`, five times. Two temporary log lines (removed after measuring, not shipped) bracketed the
interval: one at the very top of `main()`, one on the first frontend→Go call the app makes on
mount (`Pinned.List`, called from `usePinned`'s mount effect) guarded by `sync.Once`. The delta
between those two timestamps is "process spawn → frontend has mounted, run its first effect, and
completed a round trip over the Wails IPC bridge" — a reasonable proxy for "interactive."

### Scroll FPS

The native webview isn't reachable from browser automation tooling, so this measures the same
`FileList` component the app ships, mounted standalone: a temporary Vite entry
(`perf-harness.html` / `src/perf-harness.tsx`, both removed after measuring) rendered `<FileList>`
with 50,000 synthetic entries, fed no Wails bindings at all. Built for production
(`npm run build`, real minified React), served statically, then driven with Playwright against the
system's real Chrome install (`channel: 'chrome'`) — headed, not headless, so frames are actually
vsync'd to the display rather than an unthrottled headless compositor. A 4-second continuous
programmatic scroll (eased sine sweep, not linear) sampled `requestAnimationFrame` deltas; those
480 frame times are what's charted. This machine's display is 120Hz (ProMotion), which is why the
average reads ~120fps rather than capped at 60 — the number that matters against the spec target
is the *worst* single frame (62.9fps) and the dropped-frame count relative to a 60fps budget
(zero).

None of the above is wired into CI — it's a manual, repeatable procedure, not an automated
regression gate. Re-run it after any change to search ranking, the virtualized list, or startup
sequencing.
