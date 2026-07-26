import { useEffect, useState } from 'react'
import { File, Folder } from 'lucide-react'
import { vault, type Entry } from '../lib/vault'
import { formatDateTime } from '../lib/formatDate'
import { formatSize } from '../lib/formatSize'
import { extOf } from '../lib/previewKind'
import { useFocusTrap } from '../hooks/useFocusTrap'

interface PropertiesDialogProps {
  path: string
  onClose: () => void
}

const HEADING_ID = 'properties-heading'

function nameOf(path: string): string {
  const parts = path.split('/').filter(Boolean)
  return parts[parts.length - 1] || path
}

function parentOf(path: string): string {
  const parts = path.split('/').filter(Boolean)
  return '/' + parts.slice(0, -1).join('/')
}

function kindLabel(name: string, isDir: boolean): string {
  if (isDir) return 'Folder'
  const ext = extOf(name)
  return ext ? `${ext.toUpperCase()} File` : 'File'
}

function itemsLabel(count: number): string {
  return `${count} item${count === 1 ? '' : 's'}`
}

// Right-click > Properties (spec §5.5): a lightweight Get-Info-style panel
// — the file/folder's own metadata isn't otherwise surfaced anywhere but a
// PinnedRail hover card, and that only covers pinned locations. Fetches its
// own data (mirroring that hover card) rather than relying on the caller to
// have full Entry data already, since the context menu only carries
// {path, isDir}.
export function PropertiesDialog({ path, onClose }: PropertiesDialogProps) {
  const containerRef = useFocusTrap<HTMLDivElement>()
  const [entry, setEntry] = useState<Entry | null>(null)
  const [itemCount, setItemCount] = useState<number | null>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    let cancelled = false
    Promise.all([vault.stat(path), vault.list(path).catch(() => null)]).then(
      ([statEntry, children]) => {
        if (cancelled) return
        setEntry(statEntry)
        setItemCount(statEntry.isDir && children ? children.length : null)
      },
      () => {
        if (!cancelled) setError(true)
      },
    )
    return () => {
      cancelled = true
    }
  }, [path])

  return (
    <div className="fixed inset-0 z-50" onClick={onClose}>
      <div
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={HEADING_ID}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => {
          if (e.key === 'Escape') {
            e.preventDefault()
            onClose()
          }
        }}
        className="absolute top-1/2 left-1/2 flex w-80 -translate-x-1/2 -translate-y-1/2 flex-col gap-3 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-4 shadow-lg"
      >
        <div className="flex items-center gap-2">
          {entry?.isDir ? (
            <Folder size={20} strokeWidth={1.5} className="shrink-0 text-[var(--muted)]" aria-hidden="true" />
          ) : (
            <File size={20} strokeWidth={1.5} className="shrink-0 text-[var(--muted)]" aria-hidden="true" />
          )}
          <h1 id={HEADING_ID} className="truncate text-sm font-medium text-[var(--fg)]">
            {nameOf(path)}
          </h1>
        </div>

        {error ? (
          <p className="text-xs text-[var(--muted)]">This item no longer exists.</p>
        ) : entry === null ? (
          <p className="text-xs text-[var(--muted)]">Loading…</p>
        ) : (
          <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1.5 text-xs">
            <dt className="text-[var(--muted)]">Kind</dt>
            <dd className="truncate text-[var(--fg)]">{kindLabel(entry.name, entry.isDir)}</dd>

            <dt className="text-[var(--muted)]">Location</dt>
            <dd className="truncate text-[var(--fg)]" title={parentOf(path)}>
              {parentOf(path)}
            </dd>

            <dt className="text-[var(--muted)]">Size</dt>
            <dd className="text-[var(--fg)]">
              {entry.isDir ? (itemCount === null ? 'Loading…' : itemsLabel(itemCount)) : formatSize(entry.size, false)}
            </dd>

            <dt className="text-[var(--muted)]">Modified</dt>
            <dd className="text-[var(--fg)]">{formatDateTime(entry.modTime)}</dd>
          </dl>
        )}

        <button
          type="button"
          onClick={onClose}
          className="self-end rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-[var(--surface)] px-3 py-1 text-xs text-[var(--fg)] transition-colors duration-150 hover:bg-[var(--border)]"
        >
          Close
        </button>
      </div>
    </div>
  )
}
