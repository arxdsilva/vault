import { useRef, useEffect } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { File, Folder, X } from 'lucide-react'
import type { Entry } from '../lib/vault'
import { formatDateTime } from '../lib/formatDate'
import { formatSize } from '../lib/formatSize'
import { entriesAt } from '../lib/multiSelect'
import { useMultiSelect } from '../hooks/useMultiSelect'
import { PINNED_DRAG_TYPE } from './PinnedRail'
import { RenameInput } from './RenameInput'

const ROW_HEIGHT = 40

interface FileListProps {
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
  // appear for an explicit mark, not just because some row has focus.
  onMarkedCountChange?: (count: number) => void
}

function optionId(index: number): string {
  return `file-row-${index}`
}

// Virtualized file list (spec §4.3, §5.3): only visible rows are rendered so
// scrolling stays smooth even with tens of thousands of entries. Exposed as
// a listbox (role="listbox"/"option", aria-selected) — see SearchResults.tsx
// for the same pattern and its virtualization caveat.
//
// Marking (Gmail-style, via useMultiSelect) is deliberately separate from
// the keyboard/click cursor: plain click and plain arrow keys only move
// `focus` (highlighted, but not marked); the checkbox, "x", Shift-range, and
// Cmd/Ctrl-click/toggle/A build up `marked`, what bulk actions operate on.
// Browsing around never disturbs a build-up selection. When nothing is
// marked, the focused row is reported as the effective single-item
// selection so single-file shortcuts (Delete, Copy, rename, …) keep working
// without an explicit mark first.
export function FileList({
  entries,
  onOpen,
  onSelectionChange,
  onContextMenu,
  renamingPath,
  onRenameSubmit,
  onRenameCancel,
  clearSelectionSignal,
  onMarkedCountChange,
}: FileListProps) {
  const parentRef = useRef<HTMLDivElement | null>(null)
  const { marked, focus, handleClick, moveFocus, selectForContextMenu, selectAll, toggle } = useMultiSelect(
    entries.length,
    clearSelectionSignal,
  )

  useEffect(() => {
    const effective = marked.size > 0 ? entriesAt(marked, entries) : entries[focus] ? [entries[focus]] : []
    onSelectionChange?.(effective)
    onMarkedCountChange?.(marked.size)
    // onSelectionChange identity may change every render; only re-fire when
    // the actual selection (marks, cursor, or list) changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [marked, focus, entries])

  const virtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  })

  function handleKeyDown(e: React.KeyboardEvent) {
    if (entries.length === 0) return
    const mod = e.metaKey || e.ctrlKey
    if (mod && e.key.toLowerCase() === 'a') {
      e.preventDefault()
      selectAll()
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      virtualizer.scrollToIndex(moveFocus(1, e.shiftKey))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      virtualizer.scrollToIndex(moveFocus(-1, e.shiftKey))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (entries[focus]) onOpen(entries[focus])
    } else if (e.key.toLowerCase() === 'x' && !mod && !e.altKey && focus >= 0) {
      // Gmail-style toggle: "x" checks/unchecks the focused row without
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
      ref={parentRef}
      role="listbox"
      aria-label="Files"
      aria-multiselectable="true"
      tabIndex={0}
      aria-activedescendant={entries[focus] ? optionId(focus) : undefined}
      onKeyDown={handleKeyDown}
      className="min-h-0 flex-1 overflow-y-auto outline-none"
    >
      <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
        {virtualizer.getVirtualItems().map((row) => {
          const entry = entries[row.index]
          const isMarked = marked.has(row.index)
          const isFocused = row.index === focus
          const isRenaming = entry.path === renamingPath
          const Icon = entry.isDir ? Folder : File
          return (
            <div
              key={entry.path}
              id={optionId(row.index)}
              role="option"
              aria-selected={isMarked}
              aria-setsize={entries.length}
              aria-posinset={row.index + 1}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                height: row.size,
                transform: `translateY(${row.start}px)`,
              }}
              onClick={(e) => handleClick(row.index, e)}
              onDoubleClick={() => onOpen(entry)}
              onContextMenu={(e) => {
                const effective = isMarked ? entriesAt(marked, entries) : [entry]
                selectForContextMenu(row.index)
                onContextMenu?.(e, effective)
              }}
              draggable
              onDragStart={(e) => e.dataTransfer.setData(PINNED_DRAG_TYPE, entry.path)}
              className={`flex cursor-default items-center gap-3 px-3 text-sm transition-colors duration-150 ${
                isMarked ? 'bg-[var(--accent)]/15' : isFocused ? 'bg-[var(--surface)]' : 'hover:bg-[var(--surface)]'
              }`}
            >
              <button
                type="button"
                role="checkbox"
                aria-checked={isMarked}
                aria-label={isMarked ? `Unmark ${entry.name}` : `Mark ${entry.name}`}
                onClick={(e) => {
                  e.stopPropagation()
                  toggle(row.index)
                }}
                className={`flex h-4 w-4 shrink-0 items-center justify-center rounded-[3px] border transition-colors duration-150 ${
                  isMarked
                    ? 'border-[var(--accent)] bg-[var(--accent)] text-white'
                    : 'border-[var(--border)] hover:border-[var(--muted)]'
                }`}
              >
                {isMarked && <X size={11} strokeWidth={3} aria-hidden="true" />}
              </button>

              <Icon size={16} className="shrink-0 text-[var(--muted)]" strokeWidth={1.5} aria-hidden="true" />
              {isRenaming ? (
                <RenameInput
                  initialName={entry.name}
                  onSubmit={(name) => onRenameSubmit?.(entry, name)}
                  onCancel={() => onRenameCancel?.()}
                />
              ) : (
                <span className="min-w-0 flex-1 truncate">{entry.name}</span>
              )}
              <span className="w-40 shrink-0 truncate text-[var(--muted)]">{formatDateTime(entry.modTime)}</span>
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
