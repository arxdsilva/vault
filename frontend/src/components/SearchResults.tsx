import { useEffect, useRef, useState } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { File, Folder, SearchX } from 'lucide-react'
import type { IndexEntry } from '../lib/vault'
import { RenameInput } from './RenameInput'

const ROW_HEIGHT = 40

interface SearchResultsProps {
  results: IndexEntry[]
  statusLabel?: string | null
  onOpen: (entry: IndexEntry) => void
  onReveal: (entry: IndexEntry) => void
  onSelectionChange?: (entry: IndexEntry | null) => void
  onContextMenu?: (e: React.MouseEvent, entry: IndexEntry) => void
  renamingPath?: string | null
  onRenameSubmit?: (entry: IndexEntry, name: string) => void
  onRenameCancel?: () => void
}

function parentOf(path: string): string {
  const i = path.lastIndexOf('/')
  return i <= 0 ? '/' : path.slice(0, i)
}

// Search results (spec §5.2): name, dimmed parent path, file-type icon.
// Enter opens; Cmd/Ctrl+Enter reveals the result's parent folder in-app.
// Virtualized (spec §2.3/§4.3) since the result set can now run into the
// hundreds (core.Search's scopeDir needs a generous limit to guarantee the
// current folder isn't crowded out — see hooks/useSearch.ts).
export function SearchResults({
  results,
  statusLabel,
  onOpen,
  onReveal,
  onSelectionChange,
  onContextMenu,
  renamingPath,
  onRenameSubmit,
  onRenameCancel,
}: SearchResultsProps) {
  const parentRef = useRef<HTMLDivElement | null>(null)
  const [selected, setSelected] = useState(0)

  useEffect(() => {
    setSelected(0)
  }, [results])

  useEffect(() => {
    onSelectionChange?.(results[selected] ?? null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, results])

  const virtualizer = useVirtualizer({
    count: results.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  })

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (results.length === 0) return
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelected((s) => {
          const next = Math.min(s + 1, results.length - 1)
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
        const entry = results[selected]
        if (e.metaKey || e.ctrlKey) onReveal(entry)
        else onOpen(entry)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [results, selected, onOpen, onReveal])

  if (results.length === 0 && !statusLabel) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-2 text-[var(--muted)]">
        <SearchX size={28} strokeWidth={1.5} />
        <p className="text-sm">No matches.</p>
      </div>
    )
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {statusLabel && (
        <div className="shrink-0 border-b border-[var(--border)] px-3 py-1.5 text-xs text-[var(--muted)]">
          {statusLabel}
        </div>
      )}
      <div ref={parentRef} className="min-h-0 flex-1 overflow-y-auto">
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((row) => {
            const entry = results[row.index]
            const Icon = entry.isDir ? Folder : File
            const isRenaming = entry.path === renamingPath
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
                className={`flex cursor-default items-center gap-3 px-3 text-sm transition-colors duration-150 ${
                  row.index === selected ? 'bg-[var(--accent)]/15' : 'hover:bg-[var(--surface)]'
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
                <span className="max-w-[50%] shrink truncate text-[var(--muted)]">{parentOf(entry.path)}</span>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
