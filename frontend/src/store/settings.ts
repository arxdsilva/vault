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

// Gates the "?" keyboard-shortcuts help dialog (spec §5.7) — modeled on
// Gmail's Settings > Keyboard shortcuts toggle. Defaults on; the actual
// shortcuts themselves stay active regardless (this only controls the
// discoverability helper, not navigation itself).
const HELPER_STORAGE_KEY = 'vault:shortcutHelper'
const storedHelperRaw = localStorage.getItem(HELPER_STORAGE_KEY)
const storedHelperEnabled = storedHelperRaw !== 'false'

// Files/folders starting with "." (spec: Unix/Finder hidden-file
// convention) are hidden from directory listings by default, toggleable
// here. Defaults off — "false" is the only string that opts in raw, so
// anything else (including no stored value) stays hidden-by-default.
const SHOW_HIDDEN_STORAGE_KEY = 'vault:showHiddenFiles'
const storedShowHiddenRaw = localStorage.getItem(SHOW_HIDDEN_STORAGE_KEY)
const storedShowHidden = storedShowHiddenRaw === 'true'

interface SettingsState {
  accent: string
  setAccent: (value: string) => void
  theme: Theme
  setTheme: (theme: Theme) => void
  shortcutHelperEnabled: boolean
  setShortcutHelperEnabled: (enabled: boolean) => void
  showHiddenFiles: boolean
  setShowHiddenFiles: (show: boolean) => void
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
  shortcutHelperEnabled: storedHelperEnabled,
  setShortcutHelperEnabled: (enabled) => {
    localStorage.setItem(HELPER_STORAGE_KEY, String(enabled))
    set({ shortcutHelperEnabled: enabled })
  },
  showHiddenFiles: storedShowHidden,
  setShowHiddenFiles: (show) => {
    localStorage.setItem(SHOW_HIDDEN_STORAGE_KEY, String(show))
    set({ showHiddenFiles: show })
  },
}))
