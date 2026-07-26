import { useEffect, useState } from 'react'
import { File, FileVideo, Folder } from 'lucide-react'
import type { Entry } from '../lib/vault'
import { vault } from '../lib/vault'
import { previewKindOf } from '../lib/previewKind'
import { previewUrl } from '../lib/previewUrl'
import { PINNED_DRAG_TYPE } from './PinnedRail'
import { RenameInput } from './RenameInput'

interface GridTileProps {
  entry: Entry
  selected: boolean
  isRenaming: boolean
  onSelect: () => void
  onOpen: () => void
  onContextMenu: (e: React.MouseEvent) => void
  onRenameSubmit: (name: string) => void
  onRenameCancel: () => void
}

// Grid tile (spec §5.3, §5.8): thumbnail + name. Thumbnails are generated
// async on the Go side (worker pool + on-disk cache); until one arrives, or
// for kinds that don't have one (video, folders, ...), a generic icon shows.
export function GridTile({
  entry,
  selected,
  isRenaming,
  onSelect,
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
      onClick={onSelect}
      onDoubleClick={onOpen}
      onContextMenu={onContextMenu}
      draggable
      onDragStart={(e) => e.dataTransfer.setData(PINNED_DRAG_TYPE, entry.path)}
      className={`flex h-28 w-24 shrink-0 cursor-default flex-col items-center gap-1.5 rounded-[var(--radius)] p-2 text-center transition-colors duration-150 ${
        selected ? 'bg-[var(--accent)]/15' : 'hover:bg-[var(--surface)]'
      }`}
    >
      <div className="flex h-14 w-14 items-center justify-center">
        {thumbPath ? (
          <img
            src={previewUrl(thumbPath)}
            alt=""
            className="max-h-14 max-w-14 rounded-sm object-cover"
          />
        ) : (
          <Icon size={32} strokeWidth={1.5} className="text-[var(--muted)]" />
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
