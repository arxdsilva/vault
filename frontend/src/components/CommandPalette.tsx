import { useEffect, useMemo, useState } from 'react'
import { buildFileActions, type FileActionContext } from '../lib/fileActions'
import { vault, type IndexEntry } from '../lib/vault'

export type PaletteMode = 'closed' | 'actions' | 'moveTo'

interface CommandPaletteProps {
  mode: PaletteMode
  onModeChange: (mode: PaletteMode) => void
  ctx: FileActionContext
  onRunAction: (id: string) => void
  onMoveTo: (targetDir: string) => void
}

const MOVE_TO_LIMIT = 20

// Action palette (Cmd/Ctrl+K, spec §5.5), scoped to the current selection.
// "Move to…" reuses the same overlay in a second mode: a fuzzy folder
// picker over the search index, filtered to directories.
export function CommandPalette({ mode, onModeChange, ctx, onRunAction, onMoveTo }: CommandPaletteProps) {
  const [filter, setFilter] = useState('')
  const [selected, setSelected] = useState(0)
  const [folders, setFolders] = useState<IndexEntry[]>([])

  useEffect(() => {
    setFilter('')
    setSelected(0)
  }, [mode])

  const items = useMemo(
    () => buildFileActions(ctx).filter((i) => i.label.toLowerCase().includes(filter.toLowerCase())),
    [ctx, filter],
  )

  useEffect(() => {
    setSelected(0)
  }, [filter])

  useEffect(() => {
    if (mode !== 'moveTo') return
    if (filter.trim() === '') {
      setFolders([])
      return
    }
    const timer = setTimeout(() => {
      vault
        .search(filter, MOVE_TO_LIMIT)
        .then((results) => setFolders(results.filter((r) => r.isDir)))
        .catch(() => {})
    }, 50)
    return () => clearTimeout(timer)
  }, [mode, filter])

  if (mode === 'closed') return null

  const list = mode === 'actions' ? items : folders

  function selectIndex(index: number) {
    if (mode === 'actions') {
      const item = items[index]
      if (!item || item.disabled) return
      if (item.id === 'move-to') {
        onModeChange('moveTo')
        return
      }
      onRunAction(item.id)
      onModeChange('closed')
    } else {
      const folder = folders[index]
      if (!folder) return
      onMoveTo(folder.path)
      onModeChange('closed')
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault()
      onModeChange(mode === 'moveTo' ? 'actions' : 'closed')
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setSelected((s) => Math.min(s + 1, list.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setSelected((s) => Math.max(s - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      selectIndex(selected)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/20 pt-24"
      onClick={() => onModeChange('closed')}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
        className="flex w-full max-w-md flex-col overflow-hidden rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] shadow-lg"
      >
        <input
          autoFocus
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder={mode === 'actions' ? 'Type a command…' : 'Move to…'}
          className="h-11 w-full border-b border-[var(--border)] bg-transparent px-3 text-sm outline-none"
        />
        <div className="max-h-72 overflow-y-auto py-1">
          {mode === 'actions'
            ? items.map((item, i) => (
                <button
                  key={item.id}
                  type="button"
                  disabled={item.disabled}
                  onClick={() => selectIndex(i)}
                  className={`flex h-9 w-full items-center px-3 text-left text-sm transition-colors duration-150 disabled:opacity-40 ${
                    i === selected ? 'bg-[var(--accent)]/15' : 'hover:bg-[var(--surface)]'
                  }`}
                >
                  {item.label}
                </button>
              ))
            : folders.map((f, i) => (
                <button
                  key={f.path}
                  type="button"
                  onClick={() => selectIndex(i)}
                  className={`flex h-9 w-full items-center px-3 text-left text-sm transition-colors duration-150 ${
                    i === selected ? 'bg-[var(--accent)]/15' : 'hover:bg-[var(--surface)]'
                  }`}
                >
                  {f.path}
                </button>
              ))}
          {mode === 'moveTo' && folders.length === 0 && (
            <p className="px-3 py-2 text-sm text-[var(--muted)]">Type to search for a folder.</p>
          )}
        </div>
      </div>
    </div>
  )
}
