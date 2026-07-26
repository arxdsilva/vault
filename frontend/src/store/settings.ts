import { create } from 'zustand'

export interface AccentPreset {
  name: string
  value: string
}

// Spec §4.4: accent color default indigo, user-changeable among 6 presets.
export const ACCENT_PRESETS: AccentPreset[] = [
  { name: 'Indigo', value: '#6366f1' },
  { name: 'Blue', value: '#3b82f6' },
  { name: 'Green', value: '#22c55e' },
  { name: 'Amber', value: '#f59e0b' },
  { name: 'Pink', value: '#ec4899' },
  { name: 'Purple', value: '#8b5cf6' },
]

const STORAGE_KEY = 'vault:accent'

function applyAccent(value: string) {
  document.documentElement.style.setProperty('--accent', value)
}

const stored = localStorage.getItem(STORAGE_KEY)
if (stored) applyAccent(stored)

interface SettingsState {
  accent: string
  setAccent: (value: string) => void
}

// Light/dark theme itself stays OS-keyed (spec §4.4) and isn't user-
// overridable here; only the accent color is a user preference.
export const useSettings = create<SettingsState>((set) => ({
  accent: stored ?? ACCENT_PRESETS[0].value,
  setAccent: (value) => {
    applyAccent(value)
    localStorage.setItem(STORAGE_KEY, value)
    set({ accent: value })
  },
}))
