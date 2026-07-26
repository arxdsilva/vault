import { useEffect, useState } from 'react'
import { File, Folder, SearchX } from 'lucide-react'
import type { IndexEntry } from '../lib/vault'
import { RenameInput } from './RenameInput'

interface SearchResultsProps {
  results: IndexEntry[]
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
export function SearchResults({
  results,
  onOpen,
  onReveal,
  onSelectionChange,
  onContextMenu,
  renamingPath,
  onRenameSubmit,
  onRenameCancel,
}: SearchResultsProps) {
  const [selected, setSelected] = useState(0)

  useEffect(() => {
    setSelected(0)
  }, [results])

  useEffect(() => {
    onSelectionChange?.(results[selected] ?? null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, results])

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (results.length === 0) return
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        setSelected((s) => Math.min(s + 1, results.length - 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setSelected((s) => Math.max(s - 1, 0))
      } else if (e.key === 'Enter') {
        e.preventDefault()
        const entry = results[selected]
        if (e.metaKey || e.ctrlKey) onReveal(entry)
        else onOpen(entry)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [results, selected, onOpen, onReveal])

  if (results.length === 0) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-2 text-[var(--muted)]">
        <SearchX size={28} strokeWidth={1.5} />
        <p className="text-sm">No matches.</p>
      </div>
    )
  }

  return (
    <div className="min-h-0 flex-1 overflow-y-auto">
      {results.map((entry, i) => {
        const Icon = entry.isDir ? Folder : File
        const isRenaming = entry.path === renamingPath
        return (
          <div
            key={entry.path}
            onClick={() => setSelected(i)}
            onDoubleClick={() => onOpen(entry)}
            onContextMenu={(e) => {
              setSelected(i)
              onContextMenu?.(e, entry)
            }}
            className={`flex h-10 cursor-default items-center gap-3 px-3 text-sm transition-colors duration-150 ${
              i === selected ? 'bg-[var(--accent)]/15' : 'hover:bg-[var(--surface)]'
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
            <span className="max-w-[50%] shrink truncate text-[var(--muted)]">
              {parentOf(entry.path)}
            </span>
          </div>
        )
      })}
    </div>
  )
}
