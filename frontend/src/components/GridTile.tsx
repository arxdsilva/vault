import { useEffect, useState } from 'react'
import { File, FileVideo, Folder, X } from 'lucide-react'
import type { Entry } from '../lib/vault'
import { vault } from '../lib/vault'
import { previewKindOf } from '../lib/previewKind'
import { previewUrl } from '../lib/previewUrl'
import { PINNED_DRAG_TYPE } from './PinnedRail'
import { RenameInput } from './RenameInput'

interface GridTileProps {
  entry: Entry
  id: string
  marked: boolean
  focused: boolean
  setSize: number
  posInSet: number
  isRenaming: boolean
  onSelect: (e: React.MouseEvent) => void
  onToggleMark: (e: React.MouseEvent) => void
  onOpen: () => void
  onContextMenu: (e: React.MouseEvent) => void
  onRenameSubmit: (name: string) => void
  onRenameCancel: () => void
}

// Grid tile (spec §5.3, §5.8): thumbnail + name. Thumbnails are generated
// async on the Go side (worker pool + on-disk cache); until one arrives, or
// for kinds that don't have one (video, folders, ...), a generic icon shows.
//
// `marked` (checkbox, top-left corner) and `focused` (cursor highlight) are
// deliberately distinct — see useMultiSelect's doc comment.
export function GridTile({
  entry,
  id,
  marked,
  focused,
  setSize,
  posInSet,
  isRenaming,
  onSelect,
  onToggleMark,
  onOpen,
  onContextMenu,
  onRenameSubmit,
  onRenameCancel,
}: GridTileProps) {
  const kind = previewKindOf(entry.name)
  const [thumbPath, setThumbPath] = useState<string | null>(null)

  useEffect(() => {
    setThumbPath(null)
    if (entry.isDir || kind !== 'image') return
    let cancelled = false
    vault
      .thumbnail(entry.path)
      .then((path) => {
        if (!cancelled) setThumbPath(path)
      })
      .catch(() => {
        // No thumbnail available; the icon fallback below covers it.
      })
    return () => {
      cancelled = true
    }
  }, [entry.path, entry.isDir, kind])

  const Icon = entry.isDir ? Folder : kind === 'video' ? FileVideo : File

  return (
    <div
      id={id}
      role="option"
      aria-selected={marked}
      aria-setsize={setSize}
      aria-posinset={posInSet}
      onClick={onSelect}
      onDoubleClick={onOpen}
      onContextMenu={onContextMenu}
      draggable
      onDragStart={(e) => e.dataTransfer.setData(PINNED_DRAG_TYPE, entry.path)}
      className={`relative flex h-28 w-24 shrink-0 cursor-default flex-col items-center gap-1.5 rounded-[var(--radius)] p-2 text-center transition-colors duration-150 ${
        marked ? 'bg-[var(--accent)]/15' : focused ? 'bg-[var(--surface)]' : 'hover:bg-[var(--surface)]'
      }`}
    >
      <button
        type="button"
        role="checkbox"
        aria-checked={marked}
        aria-label={marked ? `Unmark ${entry.name}` : `Mark ${entry.name}`}
        onClick={(e) => {
          e.stopPropagation()
          onToggleMark(e)
        }}
        className={`absolute top-1 left-1 z-10 flex h-4 w-4 items-center justify-center rounded-[3px] border transition-colors duration-150 ${
          marked
            ? 'border-[var(--accent)] bg-[var(--accent)] text-white'
            : 'border-[var(--border)] bg-[var(--bg)] hover:border-[var(--muted)]'
        }`}
      >
        {marked && <X size={11} strokeWidth={3} aria-hidden="true" />}
      </button>

      <div className="flex h-14 w-14 items-center justify-center">
        {thumbPath ? (
          <img
            src={previewUrl(thumbPath)}
            alt=""
            className="max-h-14 max-w-14 rounded-sm object-cover"
          />
        ) : (
          <Icon size={32} strokeWidth={1.5} className="text-[var(--muted)]" aria-hidden="true" />
        )}
      </div>
      {isRenaming ? (
        <RenameInput initialName={entry.name} onSubmit={onRenameSubmit} onCancel={onRenameCancel} />
      ) : (
        <span className="w-full truncate text-xs">{entry.name}</span>
      )}
    </div>
  )
}
