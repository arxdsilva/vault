import { create } from 'zustand'

export interface ClipboardEntry {
  path: string
  isDir: boolean
}

interface ClipboardState {
  entry: ClipboardEntry | null
  copy: (entry: ClipboardEntry) => void
  clear: () => void
}

// Internal copy/paste clipboard (spec §5.5 "Copy" + Cmd/Ctrl+V), distinct
// from the OS text clipboard used by "Copy Path".
export const useClipboard = create<ClipboardState>((set) => ({
  entry: null,
  copy: (entry) => set({ entry }),
  clear: () => set({ entry: null }),
}))
