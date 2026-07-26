import { useEffect, useRef, useState } from 'react'
import { FolderOpen, LayoutGrid, List, Search, Settings } from 'lucide-react'

type ViewMode = 'list' | 'grid'

function App() {
  const [view, setView] = useState<ViewMode>('list')
  const searchRef = useRef<HTMLInputElement | null>(null)

  // The search bar is the primary way to navigate (spec §5.1) — it should be
  // focused the moment the window appears.
  useEffect(() => {
    searchRef.current?.focus()
  }, [])

  return (
    <div className="flex h-screen flex-col bg-[var(--bg)] text-[var(--fg)]">
      {/* Top bar (56px) — spec §4.1. Left area (before the search input) stays
          clear of the macOS traffic-light buttons via the inset title bar. */}
      <div className="flex h-14 shrink-0 items-center gap-3 border-b border-[var(--border)] pl-20 pr-3">
        <div className="relative flex-1">
          <Search
            size={16}
            className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-[var(--muted)]"
          />
          <input
            ref={searchRef}
            type="text"
            placeholder="Search Vault"
            className="h-9 w-full rounded-[var(--radius)] border border-[var(--border)] bg-[var(--surface)] pl-9 pr-3 text-sm outline-none transition-colors duration-150 focus:border-[var(--accent)]"
          />
        </div>

        <div className="flex items-center gap-1 rounded-[var(--radius)] border border-[var(--border)] p-0.5">
          <button
            type="button"
            aria-pressed={view === 'list'}
            onClick={() => setView('list')}
            className={`rounded-[calc(var(--radius)-2px)] p-1.5 transition-colors duration-150 ${
              view === 'list' ? 'bg-[var(--surface)] text-[var(--fg)]' : 'text-[var(--muted)]'
            }`}
          >
            <List size={16} />
          </button>
          <button
            type="button"
            aria-pressed={view === 'grid'}
            onClick={() => setView('grid')}
            className={`rounded-[calc(var(--radius)-2px)] p-1.5 transition-colors duration-150 ${
              view === 'grid' ? 'bg-[var(--surface)] text-[var(--fg)]' : 'text-[var(--muted)]'
            }`}
          >
            <LayoutGrid size={16} />
          </button>
        </div>

        <button
          type="button"
          aria-label="Settings"
          className="rounded-[var(--radius)] p-2 text-[var(--muted)] transition-colors duration-150 hover:text-[var(--fg)]"
        >
          <Settings size={16} />
        </button>
      </div>

      {/* Main area: pinned rail (48px, spec §4.3) + content. The path strip
          (spec §4.2) mounts here once folder navigation exists (Build Order
          step 3) — it's hidden on this launch/search screen either way. */}
      <div className="flex min-h-0 flex-1">
        <div className="flex w-12 shrink-0 flex-col items-center gap-2 border-r border-[var(--border)] py-3" />

        <div className="flex flex-1 flex-col items-center justify-center gap-2 text-[var(--muted)]">
          <FolderOpen size={28} strokeWidth={1.5} />
          <p className="text-sm">Vault will show your files here.</p>
        </div>
      </div>
    </div>
  )
}

export default App
