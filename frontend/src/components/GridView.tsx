import { useEffect, useRef, useState } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { Folder } from 'lucide-react'
import type { Entry } from '../lib/vault'
import { GridTile } from './GridTile'

interface GridViewProps {
  entries: Entry[]
  onOpen: (entry: Entry) => void
  onSelectionChange?: (entry: Entry | null) => void
  onContextMenu?: (e: React.MouseEvent, entry: Entry) => void
  renamingPath?: string | null
  onRenameSubmit?: (entry: Entry, name: string) => void
  onRenameCancel?: () => void
}

const TILE_SIZE = 104 // px per tile, including gap

function optionId(index: number): string {
  return `grid-tile-${index}`
}

// Virtualized grid view (spec §5.3, §5.8): only visible rows of tiles are
// rendered, same instant-scroll principle as the list view. Exposed as a
// listbox (role="listbox"/"option", aria-selected) — see FileList.tsx/
// SearchResults.tsx for the same pattern and its virtualization caveat.
export function GridView({
  entries,
  onOpen,
  onSelectionChange,
  onContextMenu,
  renamingPath,
  onRenameSubmit,
  onRenameCancel,
}: GridViewProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [width, setWidth] = useState(0)
  const [selected, setSelected] = useState(0)

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver((observed) => setWidth(observed[0].contentRect.width))
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  useEffect(() => {
    setSelected((s) => Math.min(s, Math.max(entries.length - 1, 0)))
  }, [entries.length])

  useEffect(() => {
    onSelectionChange?.(entries[selected] ?? null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, entries])

  const columns = Math.max(1, Math.floor(width / TILE_SIZE))
  const rowCount = Math.ceil(entries.length / columns)

  const virtualizer = useVirtualizer({
    count: rowCount,
    getScrollElement: () => containerRef.current,
    estimateSize: () => TILE_SIZE,
    overscan: 3,
  })

  function move(delta: number) {
    setSelected((s) => {
      const next = Math.min(Math.max(s + delta, 0), entries.length - 1)
      virtualizer.scrollToIndex(Math.floor(next / columns))
      return next
    })
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (entries.length === 0) return
    if (e.key === 'ArrowRight') {
      e.preventDefault()
      move(1)
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault()
      move(-1)
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      move(columns)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      move(-columns)
    } else if (e.key === 'Enter') {
      e.preventDefault()
      onOpen(entries[selected])
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
      tabIndex={0}
      aria-activedescendant={entries[selected] ? optionId(selected) : undefined}
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
                return (
                  <GridTile
                    key={entry.path}
                    entry={entry}
                    id={optionId(index)}
                    selected={index === selected}
                    setSize={entries.length}
                    posInSet={index + 1}
                    isRenaming={entry.path === renamingPath}
                    onSelect={() => setSelected(index)}
                    onOpen={() => onOpen(entry)}
                    onContextMenu={(e) => {
                      setSelected(index)
                      onContextMenu?.(e, entry)
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
