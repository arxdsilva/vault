import { beforeEach, describe, expect, it } from 'vitest'
import { useSettings } from './settings'

describe('useSettings theme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
    useSettings.setState({ theme: 'system' })
  })

  it('setTheme("dark") sets data-theme and persists to localStorage', () => {
    useSettings.getState().setTheme('dark')

    expect(useSettings.getState().theme).toBe('dark')
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
    expect(localStorage.getItem('vault:theme')).toBe('dark')
  })

  it('setTheme("light") sets data-theme and persists to localStorage', () => {
    useSettings.getState().setTheme('light')

    expect(useSettings.getState().theme).toBe('light')
    expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    expect(localStorage.getItem('vault:theme')).toBe('light')
  })

  it('setTheme("system") clears data-theme so the OS setting takes back over', () => {
    useSettings.getState().setTheme('dark')
    useSettings.getState().setTheme('system')

    expect(useSettings.getState().theme).toBe('system')
    expect(document.documentElement.hasAttribute('data-theme')).toBe(false)
    expect(localStorage.getItem('vault:theme')).toBe('system')
  })
})

describe('useSettings shortcutHelperEnabled', () => {
  beforeEach(() => {
    localStorage.clear()
    useSettings.setState({ shortcutHelperEnabled: true })
  })

  it('defaults to enabled', () => {
    expect(useSettings.getState().shortcutHelperEnabled).toBe(true)
  })

  it('setShortcutHelperEnabled(false) updates state and persists to localStorage', () => {
    useSettings.getState().setShortcutHelperEnabled(false)

    expect(useSettings.getState().shortcutHelperEnabled).toBe(false)
    expect(localStorage.getItem('vault:shortcutHelper')).toBe('false')
  })

  it('setShortcutHelperEnabled(true) updates state and persists to localStorage', () => {
    useSettings.getState().setShortcutHelperEnabled(false)
    useSettings.getState().setShortcutHelperEnabled(true)

    expect(useSettings.getState().shortcutHelperEnabled).toBe(true)
    expect(localStorage.getItem('vault:shortcutHelper')).toBe('true')
  })
})
