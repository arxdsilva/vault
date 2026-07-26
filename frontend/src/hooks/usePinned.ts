import { useEffect, useState } from 'react'
import { vault } from '../lib/vault'

// User-configurable pinned locations (spec §5.1, §4.3). Updates are applied
// optimistically and persisted by the Go side.
export function usePinned() {
  const [pinned, setPinned] = useState<string[]>([])

  useEffect(() => {
    vault.listPinned().then(setPinned).catch(() => {})
  }, [])

  function add(path: string) {
    setPinned((prev) => (prev.includes(path) ? prev : [...prev, path]))
    vault.addPinned(path).catch(() => {})
  }

  function remove(path: string) {
    setPinned((prev) => prev.filter((p) => p !== path))
    vault.removePinned(path).catch(() => {})
  }

  function reorder(paths: string[]) {
    setPinned(paths)
    vault.reorderPinned(paths).catch(() => {})
  }

  return { pinned, add, remove, reorder }
}
