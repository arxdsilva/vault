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

// Spec §4.4: light/dark defaults to the OS setting ("system") but is
// user-overridable to force one or the other regardless of OS.
export type Theme = 'system' | 'light' | 'dark'

const THEME_STORAGE_KEY = 'vault:theme'

function isTheme(value: string | null): value is Theme {
  return value === 'system' || value === 'light' || value === 'dark'
}

// "system" removes the override so index.css's prefers-color-scheme media
// query takes back over; "light"/"dark" set data-theme, which index.css
// gives priority over the OS setting.
function applyTheme(theme: Theme) {
  if (theme === 'system') {
    document.documentElement.removeAttribute('data-theme')
  } else {
    document.documentElement.setAttribute('data-theme', theme)
  }
}

const storedThemeRaw = localStorage.getItem(THEME_STORAGE_KEY)
const storedTheme: Theme = isTheme(storedThemeRaw) ? storedThemeRaw : 'system'
applyTheme(storedTheme)

interface SettingsState {
  accent: string
  setAccent: (value: string) => void
  theme: Theme
  setTheme: (theme: Theme) => void
}

export const useSettings = create<SettingsState>((set) => ({
  accent: stored ?? ACCENT_PRESETS[0].value,
  setAccent: (value) => {
    applyAccent(value)
    localStorage.setItem(STORAGE_KEY, value)
    set({ accent: value })
  },
  theme: storedTheme,
  setTheme: (theme) => {
    applyTheme(theme)
    localStorage.setItem(THEME_STORAGE_KEY, theme)
    set({ theme })
  },
}))
