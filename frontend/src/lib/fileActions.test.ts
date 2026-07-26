import { describe, expect, it } from 'vitest'
import { buildFileActions } from './fileActions'

describe('buildFileActions', () => {
  it('disables entry-scoped actions when nothing is selected', () => {
    const items = buildFileActions({ entries: [], currentDir: '/home', pinned: [], clipboardEntries: [] })
    const byId = Object.fromEntries(items.map((i) => [i.id, i]))
    expect(byId.open.disabled).toBe(true)
    expect(byId.rename.disabled).toBe(true)
    expect(byId.duplicate.disabled).toBe(true)
    expect(byId.pin.disabled).toBe(true)
    expect(byId.properties.disabled).toBe(true)
  })

  it('enables entry-scoped actions when a single item is selected', () => {
    const items = buildFileActions({
      entries: [{ path: '/home/report.txt', isDir: false }],
      currentDir: '/home',
      pinned: [],
      clipboardEntries: [],
    })
    const byId = Object.fromEntries(items.map((i) => [i.id, i]))
    expect(byId.open.disabled).toBe(false)
    expect(byId.rename.disabled).toBe(false)
    expect(byId.duplicate.disabled).toBe(false)
    expect(byId.properties.disabled).toBe(false)
  })

  it('enables multi-select-safe actions but disables single-only ones for 2+ selected', () => {
    const items = buildFileActions({
      entries: [
        { path: '/home/report.txt', isDir: false },
        { path: '/home/notes.txt', isDir: false },
      ],
      currentDir: '/home',
      pinned: [],
      clipboardEntries: [],
    })
    const byId = Object.fromEntries(items.map((i) => [i.id, i]))
    // Only make sense for exactly one item.
    expect(byId.open.disabled).toBe(true)
    expect(byId.rename.disabled).toBe(true)
    expect(byId.properties.disabled).toBe(true)
    // Apply fine across the whole selection.
    expect(byId['move-to'].disabled).toBe(false)
    expect(byId.copy.disabled).toBe(false)
    expect(byId.duplicate.disabled).toBe(false)
    expect(byId['copy-path'].disabled).toBe(false)
    expect(byId.compress.disabled).toBe(false)
    expect(byId.trash.disabled).toBe(false)
    expect(byId.pin.disabled).toBe(false)
  })

  it('labels the pin action "Pin" when unpinned and "Unpin" when pinned', () => {
    const unpinned = buildFileActions({
      entries: [{ path: '/home/report.txt', isDir: false }],
      currentDir: '/home',
      pinned: [],
      clipboardEntries: [],
    })
    expect(unpinned.find((i) => i.id === 'pin')?.label).toBe('Pin')

    const pinned = buildFileActions({
      entries: [{ path: '/home/report.txt', isDir: false }],
      currentDir: '/home',
      pinned: ['/home/report.txt'],
      clipboardEntries: [],
    })
    expect(pinned.find((i) => i.id === 'pin')?.label).toBe('Unpin')
  })

  it('labels the pin action "Pin" for a mixed selection (not every item pinned)', () => {
    const mixed = buildFileActions({
      entries: [
        { path: '/home/report.txt', isDir: false },
        { path: '/home/notes.txt', isDir: false },
      ],
      currentDir: '/home',
      pinned: ['/home/report.txt'],
      clipboardEntries: [],
    })
    expect(mixed.find((i) => i.id === 'pin')?.label).toBe('Pin')
  })

  it('labels the pin action "Unpin" only when every selected item is already pinned', () => {
    const allPinned = buildFileActions({
      entries: [
        { path: '/home/report.txt', isDir: false },
        { path: '/home/notes.txt', isDir: false },
      ],
      currentDir: '/home',
      pinned: ['/home/report.txt', '/home/notes.txt'],
      clipboardEntries: [],
    })
    expect(allPinned.find((i) => i.id === 'pin')?.label).toBe('Unpin')
  })

  it('disables New Folder outside of a browsed directory', () => {
    const items = buildFileActions({ entries: [], currentDir: null, pinned: [], clipboardEntries: [] })
    expect(items.find((i) => i.id === 'new-folder')?.disabled).toBe(true)
  })

  it('enables Paste only when the clipboard holds something and a directory is open', () => {
    const empty = buildFileActions({ entries: [], currentDir: '/home', pinned: [], clipboardEntries: [] })
    expect(empty.find((i) => i.id === 'paste')?.disabled).toBe(true)

    const full = buildFileActions({
      entries: [],
      currentDir: '/home',
      pinned: [],
      clipboardEntries: [{ path: '/home/report.txt', isDir: false }],
    })
    expect(full.find((i) => i.id === 'paste')?.disabled).toBe(false)
  })

  it('always disables Open With, since no native app picker is wired up yet', () => {
    const items = buildFileActions({
      entries: [{ path: '/home/report.txt', isDir: false }],
      currentDir: '/home',
      pinned: [],
      clipboardEntries: [],
    })
    expect(items.find((i) => i.id === 'open-with')?.disabled).toBe(true)
  })
})
