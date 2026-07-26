import { useState } from 'react'
import { Folder } from 'lucide-react'

export const PINNED_DRAG_TYPE = 'text/vault-path'

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

// Slim 48px rail of pinned locations (spec §4.3): icons only, tooltip on
// hover, reorderable and extendable by dragging a folder onto it.
export function PinnedRail({ pinned, onNavigate, onAdd, onReorder }: PinnedRailProps) {
  const [dragOverRail, setDragOverRail] = useState(false)
  const [draggingIndex, setDraggingIndex] = useState<number | null>(null)

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
          title={nameOf(path)}
          draggable
          onDragStart={(e) => {
            setDraggingIndex(i)
            e.dataTransfer.setData(PINNED_DRAG_TYPE, path)
          }}
          onDragEnd={() => setDraggingIndex(null)}
          onDragOver={(e) => e.preventDefault()}
          onDrop={(e) => handleDropOnItem(e, i)}
          onClick={() => onNavigate(path)}
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[var(--radius)] text-[var(--muted)] transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)]"
        >
          <Folder size={18} strokeWidth={1.5} />
        </button>
      ))}
    </div>
  )
}
