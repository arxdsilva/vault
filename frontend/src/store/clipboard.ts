import { create } from 'zustand'

export interface ClipboardEntry {
  path: string
  isDir: boolean
}

interface ClipboardState {
  entries: ClipboardEntry[]
  copy: (entries: ClipboardEntry[]) => void
  clear: () => void
}

// Internal copy/paste clipboard (spec §5.5 "Copy" + Cmd/Ctrl+V), distinct
// from the OS text clipboard used by "Copy Path". Holds every entry from
// the last Copy (spec: multi-select) so Paste can place them all at once.
export const useClipboard = create<ClipboardState>((set) => ({
  entries: [],
  copy: (entries) => set({ entries }),
  clear: () => set({ entries: [] }),
}))
