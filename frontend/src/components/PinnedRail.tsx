import { useRef, useState } from 'react'
import { Folder } from 'lucide-react'
import { vault, type Entry } from '../lib/vault'
import { formatRelativeTime } from '../lib/relativeTime'
import { formatSize } from '../lib/formatSize'

export const PINNED_DRAG_TYPE = 'text/vault-path'

const HOVER_DELAY_MS = 350

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
  onNavigate: (path: string) => void
  onAdd: (path: string) => void
  onReorder: (paths: string[]) => void
}

function nameOf(path: string): string {
  const parts = path.split('/').filter(Boolean)
  return parts[parts.length - 1] || path
}

function itemsLabel(count: number | null): string {
  if (count === null) return 'Loading items…'
  return `${count} item${count === 1 ? '' : 's'}`
}

// Slim 48px rail of pinned locations (spec §4.3): icons only, a rich hover
// card (full path, item count/size, last modified), reorderable and
// extendable by dragging a folder onto it.
export function PinnedRail({ pinned, onNavigate, onAdd, onReorder }: PinnedRailProps) {
  const [dragOverRail, setDragOverRail] = useState(false)
  const [draggingIndex, setDraggingIndex] = useState<number | null>(null)
  const [hoverCard, setHoverCard] = useState<HoverCardState | null>(null)
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const hoverRequest = useRef(0)

  function clearHoverTimer() {
    if (hoverTimer.current !== null) {
      clearTimeout(hoverTimer.current)
      hoverTimer.current = null
    }
  }

  function handleMouseEnter(e: React.MouseEvent<HTMLButtonElement>, path: string) {
    clearHoverTimer()
    if (draggingIndex !== null) return
    const rect = e.currentTarget.getBoundingClientRect()
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
    }, HOVER_DELAY_MS)
  }

  function handleMouseLeave() {
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
    <div
      onDragOver={(e) => {
        e.preventDefault()
        setDragOverRail(true)
      }}
      onDragLeave={() => setDragOverRail(false)}
      onDrop={handleDropOnRail}
      className={`flex w-12 shrink-0 flex-col items-center gap-2 border-r border-[var(--border)] py-3 transition-colors duration-150 ${
        dragOverRail ? 'bg-[var(--accent)]/10' : ''
      }`}
    >
      {pinned.map((path, i) => (
        <button
          key={path}
          type="button"
          aria-label={nameOf(path)}
          draggable
          onDragStart={(e) => {
            setDraggingIndex(i)
            handleMouseLeave()
            e.dataTransfer.setData(PINNED_DRAG_TYPE, path)
          }}
          onDragEnd={() => setDraggingIndex(null)}
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => handleDropOnItem(e, i)}
          onMouseEnter={(e) => handleMouseEnter(e, path)}
          onMouseLeave={handleMouseLeave}
          onClick={() => onNavigate(path)}
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[var(--radius)] text-[var(--muted)] transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)]"
        >
          <Folder size={18} strokeWidth={1.5} />
        </button>
      ))}

      {hoverCard && (
        <div
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
  )
}
