import { useRef, useState } from 'react'
import { Folder, HardDrive } from 'lucide-react'
import { vault, type Entry, type Volume } from '../lib/vault'
import { formatRelativeTime } from '../lib/relativeTime'
import { formatSize } from '../lib/formatSize'

export const PINNED_DRAG_TYPE = 'text/vault-path'

const HOVER_DELAY_MS = 350
const HOVER_CARD_ID = 'pinned-hover-card'

const RAIL_WIDTH_KEY = 'vault:railWidth'
const MIN_RAIL_WIDTH = 160
const MAX_RAIL_WIDTH = 400
const DEFAULT_RAIL_WIDTH = 160

function clampRailWidth(width: number): number {
  return Math.min(MAX_RAIL_WIDTH, Math.max(MIN_RAIL_WIDTH, width))
}

function loadRailWidth(): number {
  const stored = Number(localStorage.getItem(RAIL_WIDTH_KEY))
  return Number.isFinite(stored) && stored > 0 ? clampRailWidth(stored) : DEFAULT_RAIL_WIDTH
}

interface HoverCardState {
  path: string
  top: number
  left: number
  entry: Entry | null
  itemCount: number | null
  error: boolean
}

interface PinnedRailProps {
  pinned: string[]
  volumes: Volume[]
  currentDir: string | null
  onNavigate: (path: string) => void
  onAdd: (path: string) => void
  onReorder: (paths: string[]) => void
  onEjectVolume: (path: string) => void
}

function nameOf(path: string): string {
  const parts = path.split('/').filter(Boolean)
  return parts[parts.length - 1] || path
}

function itemsLabel(count: number | null): string {
  if (count === null) return 'Loading items…'
  return `${count} item${count === 1 ? '' : 's'}`
}

// Rail of pinned locations and mounted volumes (spec §4.3): icon + name so
// entries are distinguishable at a glance, plus a rich hover card (full
// path, item count/size, last modified) shown on hover *or* keyboard focus
// so it's not a mouse-only affordance. Pinned locations are reorderable and
// extendable by dragging a folder onto the rail; volumes (external/USB
// drives, network shares) appear the moment they're mounted and can be
// ejected from the rail.
export function PinnedRail({
  pinned,
  volumes,
  currentDir,
  onNavigate,
  onAdd,
  onReorder,
  onEjectVolume,
}: PinnedRailProps) {
  const [dragOverRail, setDragOverRail] = useState(false)
  const [draggingIndex, setDraggingIndex] = useState<number | null>(null)
  const [hoverCard, setHoverCard] = useState<HoverCardState | null>(null)
  const [railWidth, setRailWidth] = useState(loadRailWidth)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hoverRequest = useRef(0)
  const railWidthRef = useRef(railWidth)

  function startResize(e: React.MouseEvent) {
    e.preventDefault()
    const startX = e.clientX
    const startWidth = railWidth
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'

    function onMouseMove(ev: MouseEvent) {
      const next = clampRailWidth(startWidth + (ev.clientX - startX))
      railWidthRef.current = next
      setRailWidth(next)
    }
    function onMouseUp() {
      document.removeEventListener('mousemove', onMouseMove)
      document.removeEventListener('mouseup', onMouseUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
      localStorage.setItem(RAIL_WIDTH_KEY, String(railWidthRef.current))
    }
    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', onMouseUp)
  }

  function clearHoverTimer() {
    if (hoverTimer.current !== null) {
      clearTimeout(hoverTimer.current)
      hoverTimer.current = null
    }
  }

  function showHoverCard(target: HTMLButtonElement, path: string, delayMs: number) {
    clearHoverTimer()
    if (draggingIndex !== null) return
    const rect = target.getBoundingClientRect()
    const requestId = ++hoverRequest.current
    hoverTimer.current = setTimeout(() => {
      setHoverCard({ path, top: rect.top, left: rect.right + 8, entry: null, itemCount: null, error: false })
      Promise.all([vault.stat(path), vault.list(path).catch(() => null)]).then(
        ([entry, children]) => {
          if (hoverRequest.current !== requestId) return
          setHoverCard({
            path,
            top: rect.top,
            left: rect.right + 8,
            entry,
            itemCount: entry.isDir && children ? children.length : null,
            error: false,
          })
        },
        () => {
          if (hoverRequest.current !== requestId) return
          setHoverCard({ path, top: rect.top, left: rect.right + 8, entry: null, itemCount: null, error: true })
        },
      )
    }, delayMs)
  }

  function hideHoverCard() {
    clearHoverTimer()
    hoverRequest.current += 1 // invalidate any in-flight fetch for the card we're leaving
    setHoverCard(null)
  }

  function dropPath(e: React.DragEvent): string {
    return e.dataTransfer.getData(PINNED_DRAG_TYPE)
  }

  function handleDropOnRail(e: React.DragEvent) {
    e.preventDefault()
    setDragOverRail(false)
    if (draggingIndex !== null) return // reorder is handled by the item's own onDrop
    const path = dropPath(e)
    if (path && !pinned.includes(path)) onAdd(path)
  }

  function handleDropOnItem(e: React.DragEvent, index: number) {
    e.preventDefault()
    e.stopPropagation()
    if (draggingIndex !== null) {
      const next = [...pinned]
      const [moved] = next.splice(draggingIndex, 1)
      next.splice(index, 0, moved)
      onReorder(next)
      setDraggingIndex(null)
      return
    }
    const path = dropPath(e)
    if (path && !pinned.includes(path)) onAdd(path)
  }

  return (
    <div className="relative flex shrink-0" style={{ width: railWidth }}>
      <div
        onDragOver={(e) => {
          e.preventDefault()
          setDragOverRail(true)
        }}
        onDragLeave={() => setDragOverRail(false)}
        onDrop={handleDropOnRail}
        className={`flex w-full flex-col items-stretch gap-1 overflow-y-auto border-r border-[var(--border)] px-2 py-3 transition-colors duration-150 ${
          dragOverRail ? 'bg-[var(--accent)]/10' : ''
        }`}
      >
        <div role="group" aria-label="Pinned locations" className="flex flex-col gap-1">
          {pinned.map((path, i) => (
            <button
              key={path}
              type="button"
              aria-label={nameOf(path)}
              aria-current={path === currentDir ? 'true' : undefined}
              aria-describedby={hoverCard?.path === path ? HOVER_CARD_ID : undefined}
              draggable
              onDragStart={(e) => {
                setDraggingIndex(i)
                hideHoverCard()
                e.dataTransfer.setData(PINNED_DRAG_TYPE, path)
              }}
              onDragEnd={() => setDraggingIndex(null)}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => handleDropOnItem(e, i)}
              onMouseEnter={(e) => showHoverCard(e.currentTarget, path, HOVER_DELAY_MS)}
              onMouseLeave={hideHoverCard}
              onFocus={(e) => showHoverCard(e.currentTarget, path, 0)}
              onBlur={hideHoverCard}
              onClick={() => onNavigate(path)}
              className={`flex h-9 w-full shrink-0 items-center gap-2 rounded-[var(--radius)] px-2 transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)] ${
                path === currentDir ? 'bg-[var(--surface)] text-[var(--fg)]' : 'text-[var(--muted)]'
              }`}
            >
              <Folder size={18} strokeWidth={1.5} className="shrink-0" aria-hidden="true" />
              <span className="min-w-0 flex-1 truncate text-left text-sm">{nameOf(path)}</span>
            </button>
          ))}
        </div>

        {volumes.length > 0 && (
          <>
            <div className="my-1 border-t border-[var(--border)]" />
            <p className="px-2 text-[10px] font-medium uppercase tracking-wide text-[var(--muted)]">Volumes</p>
            <div role="group" aria-label="Volumes" className="flex flex-col gap-1">
              {volumes.map((volume) => (
                <div key={volume.path} className="group flex h-9 w-full shrink-0 items-center gap-1">
                  <button
                    type="button"
                    aria-label={volume.name}
                    aria-current={volume.path === currentDir ? 'true' : undefined}
                    aria-describedby={hoverCard?.path === volume.path ? HOVER_CARD_ID : undefined}
                    onMouseEnter={(e) => showHoverCard(e.currentTarget, volume.path, HOVER_DELAY_MS)}
                    onMouseLeave={hideHoverCard}
                    onFocus={(e) => showHoverCard(e.currentTarget, volume.path, 0)}
                    onBlur={hideHoverCard}
                    onClick={() => onNavigate(volume.path)}
                    className={`flex h-9 min-w-0 flex-1 items-center gap-2 rounded-[var(--radius)] px-2 transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)] ${
                      volume.path === currentDir ? 'bg-[var(--surface)] text-[var(--fg)]' : 'text-[var(--muted)]'
                    }`}
                  >
                    <HardDrive size={18} strokeWidth={1.5} className="shrink-0" aria-hidden="true" />
                    <span className="min-w-0 flex-1 truncate text-left text-sm">{volume.name}</span>
                  </button>
                  {volume.ejectable && (
                    <button
                      type="button"
                      aria-label={`Eject ${volume.name}`}
                      onClick={() => onEjectVolume(volume.path)}
                      className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-[var(--muted)] opacity-0 transition-opacity duration-150 hover:bg-[var(--border)] hover:text-[var(--fg)] focus-visible:opacity-100 group-hover:opacity-100"
                    >
                      <span aria-hidden="true">⏏</span>
                    </button>
                  )}
                </div>
              ))}
            </div>
          </>
        )}

        {hoverCard && (
          <div
            id={HOVER_CARD_ID}
            role="status"
            className="fixed z-50 flex w-64 flex-col gap-1 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-3 text-xs shadow-lg"
            style={{ top: hoverCard.top, left: hoverCard.left }}
          >
            <p className="truncate text-sm font-medium text-[var(--fg)]">{nameOf(hoverCard.path)}</p>
            <p className="text-[var(--muted)] break-all">{hoverCard.path}</p>
            {hoverCard.error ? (
              <p className="text-[var(--muted)]">This location no longer exists.</p>
            ) : (
              <div className="flex items-center gap-1.5 text-[var(--muted)]">
                <span>
                  {hoverCard.entry === null
                    ? 'Loading…'
                    : hoverCard.entry.isDir
                      ? itemsLabel(hoverCard.itemCount)
                      : formatSize(hoverCard.entry.size, false)}
                </span>
                {hoverCard.entry && (
                  <>
                    <span>·</span>
                    <span>{formatRelativeTime(hoverCard.entry.modTime)}</span>
                  </>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      <div
        onMouseDown={startResize}
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize side panel"
        className="absolute right-0 top-0 h-full w-1.5 -mr-0.5 cursor-col-resize hover:bg-[var(--accent)]/50 active:bg-[var(--accent)]/70"
      />
    </div>
  )
}
