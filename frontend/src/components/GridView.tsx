import { useEffect, useRef, useState } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { Folder } from 'lucide-react'
import type { Entry } from '../lib/vault'
import { entriesAt } from '../lib/multiSelect'
import { useMultiSelect } from '../hooks/useMultiSelect'
import { GridTile } from './GridTile'

interface GridViewProps {
  entries: Entry[]
  onOpen: (entry: Entry) => void
  onSelectionChange?: (entries: Entry[]) => void
  onContextMenu?: (e: React.MouseEvent, entries: Entry[]) => void
  renamingPath?: string | null
  onRenameSubmit?: (entry: Entry, name: string) => void
  onRenameCancel?: () => void
  // Changing this value (e.g. an incrementing counter) clears the selection —
  // see useMultiSelect's clearSignal.
  clearSelectionSignal?: unknown
  // Reports the true marked count, separate from onSelectionChange's
  // fallback-inclusive count — App.tsx's selection action bar should only
  // appear for an explicit mark, not just because some tile has focus.
  onMarkedCountChange?: (count: number) => void
}

const TILE_SIZE = 104 // px per tile, including gap

function optionId(index: number): string {
  return `grid-tile-${index}`
}

// Virtualized grid view (spec §5.3, §5.8): only visible rows of tiles are
// rendered, same instant-scroll principle as the list view. Exposed as a
// listbox (role="listbox"/"option", aria-selected) — see FileList.tsx/
// SearchResults.tsx for the same pattern and its virtualization caveat.
// Marking shares useMultiSelect with FileList (see its doc comment for the
// focus-vs-mark split); here a "move" is a 2D step (columns computed from
// the container width) rather than +/-1.
export function GridView({
  entries,
  onOpen,
  onSelectionChange,
  onContextMenu,
  renamingPath,
  onRenameSubmit,
  onRenameCancel,
  clearSelectionSignal,
  onMarkedCountChange,
}: GridViewProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [width, setWidth] = useState(0)
  const { marked, focus, handleClick, moveFocus, selectForContextMenu, selectAll, toggle } = useMultiSelect(
    entries.length,
    clearSelectionSignal,
  )

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver((observed) => setWidth(observed[0].contentRect.width))
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    const effective = marked.size > 0 ? entriesAt(marked, entries) : entries[focus] ? [entries[focus]] : []
    onSelectionChange?.(effective)
    onMarkedCountChange?.(marked.size)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [marked, focus, entries])

  const columns = Math.max(1, Math.floor(width / TILE_SIZE))
  const rowCount = Math.ceil(entries.length / columns)

  const virtualizer = useVirtualizer({
    count: rowCount,
    getScrollElement: () => containerRef.current,
    estimateSize: () => TILE_SIZE,
    overscan: 3,
  })

  function move(delta: number, extend: boolean) {
    const next = moveFocus(delta, extend)
    virtualizer.scrollToIndex(Math.floor(next / columns))
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (entries.length === 0) return
    const mod = e.metaKey || e.ctrlKey
    if (mod && e.key.toLowerCase() === 'a') {
      e.preventDefault()
      selectAll()
    } else if (e.key === 'ArrowRight') {
      e.preventDefault()
      move(1, e.shiftKey)
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      move(-1, e.shiftKey)
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      move(columns, e.shiftKey)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      move(-columns, e.shiftKey)
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (entries[focus]) onOpen(entries[focus])
    } else if (e.key.toLowerCase() === 'x' && !mod && !e.altKey && focus >= 0) {
      // Gmail-style toggle: "x" checks/unchecks the focused tile without
      // moving focus, building a selection one keystroke at a time.
      e.preventDefault()
      toggle(focus)
    }
  }

  if (entries.length === 0) {
    return (
      <div role="status" className="flex flex-1 flex-col items-center justify-center gap-2 text-[var(--muted)]">
        <Folder size={28} strokeWidth={1.5} aria-hidden="true" />
        <p className="text-sm">This folder is empty.</p>
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      role="listbox"
      aria-label="Files"
      aria-multiselectable="true"
      tabIndex={0}
      aria-activedescendant={entries[focus] ? optionId(focus) : undefined}
      onKeyDown={handleKeyDown}
      className="min-h-0 flex-1 overflow-y-auto outline-none"
    >
      {width > 0 && (
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((row) => (
            <div
              key={row.index}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: row.size,
                transform: `translateY(${row.start}px)`,
              }}
              className="flex gap-1 px-2"
            >
              {Array.from({ length: columns }, (_, col) => {
                const index = row.index * columns + col
                const entry = entries[index]
                if (!entry) return <div key={col} className="h-28 w-24 shrink-0" />
                const isMarked = marked.has(index)
                return (
                  <GridTile
                    key={entry.path}
                    entry={entry}
                    id={optionId(index)}
                    marked={isMarked}
                    focused={index === focus}
                    setSize={entries.length}
                    posInSet={index + 1}
                    isRenaming={entry.path === renamingPath}
                    onSelect={(e) => handleClick(index, e)}
                    onToggleMark={() => toggle(index)}
                    onOpen={() => onOpen(entry)}
                    onContextMenu={(e) => {
                      const effective = isMarked ? entriesAt(marked, entries) : [entry]
                      selectForContextMenu(index)
                      onContextMenu?.(e, effective)
                    }}
                    onRenameSubmit={(name) => onRenameSubmit?.(entry, name)}
                    onRenameCancel={() => onRenameCancel?.()}
                  />
                )
              })}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
