import { create } from 'zustand'

export type SortKey = 'name' | 'modified' | 'size' | 'kind'

interface NavigationState {
  // null = launch/search screen (spec §5.1); a path = browsing that folder.
  currentDir: string | null
  sortByDir: Record<string, SortKey>
  navigate: (dir: string | null) => void
  parentOf: (dir: string) => string | null
  sortFor: (dir: string) => SortKey
  setSort: (dir: string, key: SortKey) => void
}

// Splits a POSIX-ish path into breadcrumb segments. Good enough for macOS/
// Linux; Windows drive letters ("C:\\Users\\...") get a platform-specific
// pass when that adapter lands (see docs/SPEC.md §4.2).
export function segmentsOf(dir: string): { label: string; path: string }[] {
  const parts = dir.split('/').filter(Boolean)
  let path = ''
  return parts.map((label) => {
    path += '/' + label
    return { label, path }
  })
}

export const useNavigation = create<NavigationState>((set, get) => ({
  currentDir: null,
  sortByDir: {},

  navigate: (dir) => set({ currentDir: dir }),

  parentOf: (dir) => {
    const parent = dir.replace(/\/+$/, '').split('/').slice(0, -1).join('/')
    return parent === '' ? '/' : parent
  },

  sortFor: (dir) => get().sortByDir[dir] ?? 'name',

  setSort: (dir, key) =>
    set((state) => ({ sortByDir: { ...state.sortByDir, [dir]: key } })),
}))
