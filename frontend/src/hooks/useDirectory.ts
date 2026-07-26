import { useEffect, useState } from 'react'
import { vault, type Entry } from '../lib/vault'
import type { SortKey } from '../store/navigation'

interface DirectoryState {
  entries: Entry[]
  loading: boolean
  error: string | null
}

const comparators: Record<SortKey, (a: Entry, b: Entry) => number> = {
  name: (a, b) => a.name.localeCompare(b.name),
  modified: (a, b) => new Date(b.modTime).getTime() - new Date(a.modTime).getTime(),
  size: (a, b) => b.size - a.size,
  kind: (a, b) => extOf(a.name).localeCompare(extOf(b.name)) || a.name.localeCompare(b.name),
}

function extOf(name: string): string {
  const i = name.lastIndexOf('.')
  return i <= 0 ? '' : name.slice(i + 1).toLowerCase()
}

// Loads dir's entries, keeps the currently watched directory in sync with
// dir (spec §5.3 browsing), and re-sorts live as fsnotify events arrive.
export function useDirectory(dir: string | null, sort: SortKey): DirectoryState {
  const [state, setState] = useState<DirectoryState>({ entries: [], loading: false, error: null })

  useEffect(() => {
    if (dir === null) {
      setState({ entries: [], loading: false, error: null })
      return
    }

    let cancelled = false
    setState((s) => ({ ...s, loading: true, error: null }))

    async function load() {
      try {
        const entries = await vault.list(dir as string)
        if (!cancelled) setState({ entries, loading: false, error: null })
      } catch (err) {
        if (!cancelled) setState({ entries: [], loading: false, error: String(err) })
      }
    }
    load()

    vault.watch(dir).catch(() => {
      // Non-fatal: the listing above already succeeded, so browsing still
      // works — only live updates are unavailable (e.g. permission denied).
    })
    const offChange = vault.onChange((change) => {
      if (change.dir !== dir) return
      load()
    })

    return () => {
      cancelled = true
      offChange()
      vault.unwatch()
    }
  }, [dir])

  const sorted = [...state.entries].sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
    return comparators[sort](a, b)
  })

  return { ...state, entries: sorted }
}
