import { useEffect, useRef, useState } from 'react'
import { vault, type IndexEntry } from '../lib/vault'
import { buildSearchStages } from '../lib/searchStages'

const DEBOUNCE_MS = 50
const RESULT_LIMIT = 50
const STAGE_DELAY_MS = 180
const MAX_ANCESTOR_LEVELS = 4

export interface SearchState {
  results: IndexEntry[]
  label: string | null
}

const EMPTY_STATE: SearchState = { results: [], label: null }

// Debounced fuzzy search over the indexed home directory (spec §5.2). When
// browsing a folder, results are staged: current-folder matches appear
// first, then the search progressively widens to parent folders — "expand
// slowly ... show what it's doing" — until the whole fetched result set has
// been revealed. With no folder open (e.g. the Home screen), there's no
// context to scope by, so results appear all at once as before.
export function useSearch(query: string, currentDir: string | null): SearchState {
  const [state, setState] = useState<SearchState>(EMPTY_STATE)
  const latestRequest = useRef(0)

  useEffect(() => {
    if (query.trim() === '') {
      setState(EMPTY_STATE)
      return
    }

    const requestId = ++latestRequest.current
    const timers: ReturnType<typeof setTimeout>[] = []

    const debounce = setTimeout(() => {
      Promise.all([vault.search(query, RESULT_LIMIT), currentDir ? vault.homeDir() : Promise.resolve('')]).then(
        ([entries, home]) => {
          if (latestRequest.current !== requestId) return

          const stages = buildSearchStages(entries, currentDir, home, MAX_ANCESTOR_LEVELS)
          stages.forEach((stage, i) => {
            timers.push(
              setTimeout(() => {
                if (latestRequest.current === requestId) setState(stage)
              }, i * STAGE_DELAY_MS),
            )
          })
        },
      )
    }, DEBOUNCE_MS)

    return () => {
      clearTimeout(debounce)
      timers.forEach(clearTimeout)
    }
  }, [query, currentDir])

  return state
}
