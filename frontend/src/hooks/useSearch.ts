import { useEffect, useRef, useState } from 'react'
import { vault, type IndexEntry } from '../lib/vault'

const DEBOUNCE_MS = 50
const RESULT_LIMIT = 50

// Debounced fuzzy search over the indexed home directory (spec §5.2).
export function useSearch(query: string): IndexEntry[] {
  const [results, setResults] = useState<IndexEntry[]>([])
  const latestRequest = useRef(0)

  useEffect(() => {
    if (query.trim() === '') {
      setResults([])
      return
    }

    const requestId = ++latestRequest.current
    const timer = setTimeout(() => {
      vault.search(query, RESULT_LIMIT).then((entries) => {
        if (latestRequest.current === requestId) setResults(entries)
      })
    }, DEBOUNCE_MS)

    return () => clearTimeout(timer)
  }, [query])

  return results
}
