import { useEffect, useRef, useState } from 'react'
import { AlertCircle, LayoutGrid, List, Search, Settings } from 'lucide-react'
import { Breadcrumbs } from './components/Breadcrumbs'
import { CommandPalette, type PaletteMode } from './components/CommandPalette'
import { ContextMenu } from './components/ContextMenu'
import { FileList } from './components/FileList'
import { GridView } from './components/GridView'
import { HomeScreen } from './components/HomeScreen'
import { PinnedRail } from './components/PinnedRail'
import { PreviewOverlay } from './components/PreviewOverlay'
import { SearchResults } from './components/SearchResults'
import { Toast } from './components/Toast'
import { useDirectory } from './hooks/useDirectory'
import { usePinned } from './hooks/usePinned'
import { useSearch } from './hooks/useSearch'
import type { FileActionContext, FileActionEntry } from './lib/fileActions'
import { vault, type Entry, type IndexEntry, type RecentEntry } from './lib/vault'
import { useClipboard } from './store/clipboard'
import { useNavigation, type SortKey } from './store/navigation'
import { useToast } from './store/toast'

type ViewMode = 'list' | 'grid'

function isTypingTarget(target: EventTarget | null): boolean {
  return target instanceof HTMLElement && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA')
}

function App() {
  const [view, setView] = useState<ViewMode>('list')
  const [query, setQuery] = useState('')
  const [selectedEntry, setSelectedEntry] = useState<FileActionEntry | null>(null)
  const [renamingPath, setRenamingPath] = useState<string | null>(null)
  const [paletteMode, setPaletteMode] = useState<PaletteMode>('closed')
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number } | null>(null)
  const [previewPath, setPreviewPath] = useState<string | null>(null)
  const searchRef = useRef<HTMLInputElement | null>(null)
  const { currentDir, navigate, parentOf, sortFor, setSort } = useNavigation()
  const sort = currentDir ? sortFor(currentDir) : 'name'
  const { entries, error } = useDirectory(currentDir, sort)
  const searchResults = useSearch(query)
  const searching = query.trim() !== ''
  const { pinned, add: addPinned, remove: removePinned, reorder: reorderPinned } = usePinned()
  const clipboard = useClipboard()
  const showError = useToast((s) => s.show)

  const actionCtx: FileActionContext = {
    entry: selectedEntry,
    currentDir,
    pinned,
    clipboardEntry: clipboard.entry,
  }

  // Files in the currently visible list, in display order — what Space
  // previews and what arrow keys step through while the preview is open
  // (spec §5.4).
  const previewFiles = (searching ? searchResults : entries).filter((e) => !e.isDir)
  const previewEntry = previewFiles.find((e) => e.path === previewPath) ?? null

  function movePreview(delta: number) {
    if (!previewPath) return
    const index = previewFiles.findIndex((e) => e.path === previewPath)
    const next = previewFiles[index + delta]
    if (next) setPreviewPath(next.path)
  }

  // The search bar is the primary way to navigate (spec §5.1) — it should be
  // focused the moment the window appears.
  useEffect(() => {
    searchRef.current?.focus()
  }, [])

  async function handlePaste() {
    if (!clipboard.entry || !currentDir) return
    try {
      await vault.paste(clipboard.entry.path, currentDir)
    } catch (err) {
      showError(String(err))
    }
  }

  async function handleNewFolder() {
    if (!currentDir) return
    try {
      const path = await vault.newFolder(currentDir, 'Untitled Folder')
      setRenamingPath(path)
    } catch (err) {
      showError(String(err))
    }
  }

  useEffect(() => {
    function handleGlobalKeyDown(e: KeyboardEvent) {
      const mod = e.metaKey || e.ctrlKey

      if (mod && e.key.toLowerCase() === 'k') {
        e.preventDefault()
        setPaletteMode((m) => (m === 'closed' ? 'actions' : 'closed'))
        return
      }
      if (isTypingTarget(e.target)) return

      if (mod && e.key === 'ArrowUp' && currentDir) {
        e.preventDefault()
        navigate(parentOf(currentDir))
        return
      }
      if (mod && e.key.toLowerCase() === 'z') {
        e.preventDefault()
        vault.undo().catch((err) => showError(String(err)))
        return
      }
      if (mod && e.key.toLowerCase() === 'v') {
        e.preventDefault()
        handlePaste()
        return
      }
      if (mod && e.key.toLowerCase() === 'n') {
        e.preventDefault()
        handleNewFolder()
        return
      }
      if (e.key === 'F2' && selectedEntry) {
        e.preventDefault()
        setRenamingPath(selectedEntry.path)
        return
      }
      if (
        e.key === ' ' &&
        !previewPath &&
        paletteMode === 'closed' &&
        !contextMenu &&
        selectedEntry &&
        !selectedEntry.isDir
      ) {
        e.preventDefault()
        setPreviewPath(selectedEntry.path)
      }
    }
    window.addEventListener('keydown', handleGlobalKeyDown)
    return () => window.removeEventListener('keydown', handleGlobalKeyDown)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentDir, navigate, parentOf, selectedEntry, clipboard.entry, previewPath, paletteMode, contextMenu])

  function handleOpen(entry: Entry) {
    vault.touch(entry.path)
    if (entry.isDir) navigate(entry.path)
  }

  function handleOpenRecent(entry: RecentEntry) {
    vault.touch(entry.path)
    // Opening a file (preview / default app) lands in Build Order step 7;
    // for now, reveal it in its folder so it's at least reachable.
    navigate(entry.isDir ? entry.path : parentOf(entry.path))
  }

  function handleOpenResult(entry: IndexEntry) {
    vault.touch(entry.path)
    if (entry.isDir) {
      navigate(entry.path)
    } else {
      navigate(parentOf(entry.path))
    }
    setQuery('')
  }

  function handleRevealResult(entry: IndexEntry) {
    vault.touch(entry.path)
    navigate(parentOf(entry.path))
    setQuery('')
  }

  function handleSearchKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Escape') {
      e.preventDefault()
      setQuery('')
      navigate(null) // Esc clears search and returns to Home (spec §5.2)
    }
  }

  function handleRenameSubmit(entry: { path: string }, name: string) {
    const dir = entry.path.slice(0, entry.path.lastIndexOf('/'))
    vault.rename(entry.path, `${dir}/${name}`).catch((err) => showError(String(err)))
    setRenamingPath(null)
  }

  function handleMoveTo(targetDir: string) {
    if (!selectedEntry) return
    const name = selectedEntry.path.slice(selectedEntry.path.lastIndexOf('/') + 1)
    vault.move(selectedEntry.path, `${targetDir}/${name}`).catch((err) => showError(String(err)))
  }

  function handleRunAction(id: string) {
    const entry = selectedEntry
    switch (id) {
      case 'open':
        if (entry) {
          vault.touch(entry.path)
          if (entry.isDir) navigate(entry.path)
        }
        break
      case 'rename':
        if (entry) setRenamingPath(entry.path)
        break
      case 'copy':
        if (entry) clipboard.copy(entry)
        break
      case 'paste':
        handlePaste()
        break
      case 'duplicate':
        if (entry) vault.duplicate(entry.path).catch((err) => showError(String(err)))
        break
      case 'copy-path':
        if (entry) vault.copyPathToClipboard(entry.path).catch((err) => showError(String(err)))
        break
      case 'new-folder':
        handleNewFolder()
        break
      case 'compress':
        if (entry) vault.compress(entry.path).catch((err) => showError(String(err)))
        break
      case 'trash':
        if (entry) vault.delete(entry.path).catch((err) => showError(String(err)))
        break
      case 'pin':
        if (entry) {
          if (pinned.includes(entry.path)) removePinned(entry.path)
          else addPinned(entry.path)
        }
        break
    }
  }

  function handleContextMenu(e: React.MouseEvent, entry: FileActionEntry) {
    e.preventDefault()
    setSelectedEntry(entry)
    setContextMenu({ x: e.clientX, y: e.clientY })
  }

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
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleSearchKeyDown}
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

      {/* Main area: pinned rail (48px, spec §4.3) + content. */}
      <div className="flex min-h-0 flex-1">
        <PinnedRail pinned={pinned} onNavigate={navigate} onAdd={addPinned} onReorder={reorderPinned} />

        <div className="flex min-w-0 flex-1 flex-col">
          {searching ? (
            <SearchResults
              results={searchResults}
              onOpen={handleOpenResult}
              onReveal={handleRevealResult}
              onSelectionChange={setSelectedEntry}
              onContextMenu={handleContextMenu}
              renamingPath={renamingPath}
              onRenameSubmit={handleRenameSubmit}
              onRenameCancel={() => setRenamingPath(null)}
            />
          ) : currentDir === null ? (
            <HomeScreen pinned={pinned} onNavigate={navigate} onOpenRecent={handleOpenRecent} />
          ) : (
            <>
              <div className="flex items-center">
                <div className="min-w-0 flex-1">
                  <Breadcrumbs dir={currentDir} onNavigate={navigate} />
                </div>
                <select
                  aria-label="Sort by"
                  value={sort}
                  onChange={(e) => setSort(currentDir, e.target.value as SortKey)}
                  className="mr-3 h-6 shrink-0 rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-transparent px-1 text-xs text-[var(--muted)] outline-none"
                >
                  <option value="name">Name</option>
                  <option value="modified">Modified</option>
                  <option value="size">Size</option>
                  <option value="kind">Kind</option>
                </select>
              </div>

              {error ? (
                <div className="flex flex-1 flex-col items-center justify-center gap-2 text-[var(--muted)]">
                  <AlertCircle size={28} strokeWidth={1.5} />
                  <p className="text-sm">Couldn't open this folder: {error}</p>
                </div>
              ) : view === 'grid' ? (
                <GridView
                  entries={entries}
                  onOpen={handleOpen}
                  onSelectionChange={setSelectedEntry}
                  onContextMenu={handleContextMenu}
                  renamingPath={renamingPath}
                  onRenameSubmit={handleRenameSubmit}
                  onRenameCancel={() => setRenamingPath(null)}
                />
              ) : (
                <FileList
                  entries={entries}
                  onOpen={handleOpen}
                  onSelectionChange={setSelectedEntry}
                  onContextMenu={handleContextMenu}
                  renamingPath={renamingPath}
                  onRenameSubmit={handleRenameSubmit}
                  onRenameCancel={() => setRenamingPath(null)}
                />
              )}
            </>
          )}
        </div>
      </div>

      <CommandPalette
        mode={paletteMode}
        onModeChange={setPaletteMode}
        ctx={actionCtx}
        onRunAction={handleRunAction}
        onMoveTo={handleMoveTo}
      />

      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          ctx={actionCtx}
          onRunAction={(id) => {
            if (id === 'move-to') setPaletteMode('moveTo')
            else handleRunAction(id)
          }}
          onClose={() => setContextMenu(null)}
        />
      )}

      {previewEntry && (
        <PreviewOverlay
          entry={previewEntry}
          onClose={() => setPreviewPath(null)}
          onPrev={() => movePreview(-1)}
          onNext={() => movePreview(1)}
        />
      )}

      <Toast />
    </div>
  )
}

export default App
