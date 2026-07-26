import { useEffect, useState } from 'react'
import { File, Folder } from 'lucide-react'
import { vault, type RecentEntry } from '../lib/vault'
import { formatRelativeTime } from '../lib/relativeTime'

interface HomeScreenProps {
  pinned: string[]
  onNavigate: (path: string) => void
  onOpenRecent: (entry: RecentEntry) => void
}

function nameOf(path: string): string {
  const parts = path.split('/').filter(Boolean)
  return parts[parts.length - 1] || path
}

const RECENT_LIMIT = 20

// Launch state ("Home", spec §5.1): Pinned tiles + Recent list, shown when
// no search query is active and no folder is being browsed.
export function HomeScreen({ pinned, onNavigate, onOpenRecent }: HomeScreenProps) {
  const [recent, setRecent] = useState<RecentEntry[]>([])

  useEffect(() => {
    vault.recent(RECENT_LIMIT).then(setRecent).catch(() => {})
  }, [])

  return (
    <div className="flex flex-1 flex-col gap-6 overflow-y-auto p-6">
      <section>
        <h2 className="mb-3 text-xs font-medium tracking-wide text-[var(--muted)] uppercase">
          Pinned
        </h2>
        {pinned.length === 0 ? (
          <p className="text-sm text-[var(--muted)]">Drag a folder onto the rail to pin it.</p>
        ) : (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(96px,1fr))] gap-4">
            {pinned.map((path) => (
              <button
                key={path}
                type="button"
                onClick={() => onNavigate(path)}
                className="flex flex-col items-center gap-2 rounded-[var(--radius)] p-3 text-center transition-colors duration-150 hover:bg-[var(--surface)]"
              >
                <Folder size={32} strokeWidth={1.5} className="text-[var(--accent)]" />
                <span className="w-full truncate text-xs">{nameOf(path)}</span>
              </button>
            ))}
          </div>
        )}
      </section>

      <section className="min-h-0 flex-1">
        <h2 className="mb-3 text-xs font-medium tracking-wide text-[var(--muted)] uppercase">
          Recent
        </h2>
        {recent.length === 0 ? (
          <p className="text-sm text-[var(--muted)]">Nothing opened yet.</p>
        ) : (
          <div className="flex flex-col">
            {recent.map((entry) => {
              const Icon = entry.isDir ? Folder : File
              return (
                <button
                  key={entry.path}
                  type="button"
                  onClick={() => onOpenRecent(entry)}
                  className="flex h-10 items-center gap-3 rounded-[var(--radius)] px-2 text-left text-sm transition-colors duration-150 hover:bg-[var(--surface)]"
                >
                  <Icon size={16} className="shrink-0 text-[var(--muted)]" strokeWidth={1.5} />
                  <span className="min-w-0 flex-1 truncate">{entry.name}</span>
                  <span className="shrink-0 text-xs text-[var(--muted)]">
                    {formatRelativeTime(entry.lastAccess)}
                  </span>
                </button>
              )
            })}
          </div>
        )}
      </section>
    </div>
  )
}
