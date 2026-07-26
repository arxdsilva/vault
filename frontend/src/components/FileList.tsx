import { useRef, useState, useEffect } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { File, Folder } from 'lucide-react'
import type { Entry } from '../lib/vault'
import { PINNED_DRAG_TYPE } from './PinnedRail'
import { RenameInput } from './RenameInput'

const ROW_HEIGHT = 40

interface FileListProps {
  entries: Entry[]
  onOpen: (entry: Entry) => void
  onSelectionChange?: (entry: Entry | null) => void
  onContextMenu?: (e: React.MouseEvent, entry: Entry) => void
  renamingPath?: string | null
  onRenameSubmit?: (entry: Entry, name: string) => void
  onRenameCancel?: () => void
}

const dateFormatter = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' })

function formatSize(bytes: number, isDir: boolean): string {
  if (isDir) return '—'
  if (bytes < 1024) return `${bytes} B`
  const units = ['KB', 'MB', 'GB', 'TB']
  let value = bytes / 1024
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit += 1
  }
  return `${value.toFixed(value < 10 ? 1 : 0)} ${units[unit]}`
}

// Virtualized file list (spec §4.3, §5.3): only visible rows are rendered so
// scrolling stays smooth even with tens of thousands of entries.
export function FileList({
  entries,
  onOpen,
  onSelectionChange,
  onContextMenu,
  renamingPath,
  onRenameSubmit,
  onRenameCancel,
}: FileListProps) {
  const parentRef = useRef<HTMLDivElement | null>(null)
  const [selected, setSelected] = useState(0)

  useEffect(() => {
    setSelected((s) => Math.min(s, Math.max(entries.length - 1, 0)))
  }, [entries.length])

  useEffect(() => {
    onSelectionChange?.(entries[selected] ?? null)
    // onSelectionChange identity may change every render; only re-fire when
    // the actual selection (index or list) changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, entries])

  const virtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  })

  function handleKeyDown(e: React.KeyboardEvent) {
    if (entries.length === 0) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelected((s) => {
        const next = Math.min(s + 1, entries.length - 1)
        virtualizer.scrollToIndex(next)
        return next
      })
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelected((s) => {
        const next = Math.max(s - 1, 0)
        virtualizer.scrollToIndex(next)
        return next
      })
    } else if (e.key === 'Enter') {
      e.preventDefault()
      onOpen(entries[selected])
    }
  }

  if (entries.length === 0) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-2 text-[var(--muted)]">
        <Folder size={28} strokeWidth={1.5} />
        <p className="text-sm">This folder is empty.</p>
      </div>
    )
  }

  return (
    <div
      ref={parentRef}
      tabIndex={0}
      onKeyDown={handleKeyDown}
      className="min-h-0 flex-1 overflow-y-auto outline-none"
    >
      <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
        {virtualizer.getVirtualItems().map((row) => {
          const entry = entries[row.index]
          const isSelected = row.index === selected
          const isRenaming = entry.path === renamingPath
          const Icon = entry.isDir ? Folder : File
          return (
            <div
              key={entry.path}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: row.size,
                transform: `translateY(${row.start}px)`,
              }}
              onClick={() => setSelected(row.index)}
              onDoubleClick={() => onOpen(entry)}
              onContextMenu={(e) => {
                setSelected(row.index)
                onContextMenu?.(e, entry)
              }}
              draggable
              onDragStart={(e) => e.dataTransfer.setData(PINNED_DRAG_TYPE, entry.path)}
              className={`flex cursor-default items-center gap-3 px-3 text-sm transition-colors duration-150 ${
                isSelected ? 'bg-[var(--accent)]/15' : 'hover:bg-[var(--surface)]'
              }`}
            >
              <Icon size={16} className="shrink-0 text-[var(--muted)]" strokeWidth={1.5} />
              {isRenaming ? (
                <RenameInput
                  initialName={entry.name}
                  onSubmit={(name) => onRenameSubmit?.(entry, name)}
                  onCancel={() => onRenameCancel?.()}
                />
              ) : (
                <span className="min-w-0 flex-1 truncate">{entry.name}</span>
              )}
              <span className="w-40 shrink-0 truncate text-[var(--muted)]">
                {dateFormatter.format(new Date(entry.modTime))}
              </span>
              <span className="w-16 shrink-0 text-right text-[var(--muted)]">
                {formatSize(entry.size, entry.isDir)}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
