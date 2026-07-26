import { describe, expect, it } from 'vitest'
import { buildFileActions } from './fileActions'

describe('buildFileActions', () => {
  it('disables entry-scoped actions when nothing is selected', () => {
    const items = buildFileActions({ entry: null, currentDir: '/home', pinned: [], clipboardEntry: null })
    const byId = Object.fromEntries(items.map((i) => [i.id, i]))
    expect(byId.open.disabled).toBe(true)
    expect(byId.rename.disabled).toBe(true)
    expect(byId.duplicate.disabled).toBe(true)
    expect(byId.pin.disabled).toBe(true)
  })

  it('enables entry-scoped actions when something is selected', () => {
    const items = buildFileActions({
      entry: { path: '/home/report.txt', isDir: false },
      currentDir: '/home',
      pinned: [],
      clipboardEntry: null,
    })
    const byId = Object.fromEntries(items.map((i) => [i.id, i]))
    expect(byId.open.disabled).toBe(false)
    expect(byId.rename.disabled).toBe(false)
    expect(byId.duplicate.disabled).toBe(false)
  })

  it('labels the pin action "Pin" when unpinned and "Unpin" when pinned', () => {
    const unpinned = buildFileActions({
      entry: { path: '/home/report.txt', isDir: false },
      currentDir: '/home',
      pinned: [],
      clipboardEntry: null,
    })
    expect(unpinned.find((i) => i.id === 'pin')?.label).toBe('Pin')

    const pinned = buildFileActions({
      entry: { path: '/home/report.txt', isDir: false },
      currentDir: '/home',
      pinned: ['/home/report.txt'],
      clipboardEntry: null,
    })
    expect(pinned.find((i) => i.id === 'pin')?.label).toBe('Unpin')
  })

  it('disables New Folder outside of a browsed directory', () => {
    const items = buildFileActions({ entry: null, currentDir: null, pinned: [], clipboardEntry: null })
    expect(items.find((i) => i.id === 'new-folder')?.disabled).toBe(true)
  })

  it('enables Paste only when the clipboard holds something and a directory is open', () => {
    const empty = buildFileActions({ entry: null, currentDir: '/home', pinned: [], clipboardEntry: null })
    expect(empty.find((i) => i.id === 'paste')?.disabled).toBe(true)

    const full = buildFileActions({
      entry: null,
      currentDir: '/home',
      pinned: [],
      clipboardEntry: { path: '/home/report.txt', isDir: false },
    })
    expect(full.find((i) => i.id === 'paste')?.disabled).toBe(false)
  })

  it('enables Properties only when something is selected', () => {
    const nothingSelected = buildFileActions({ entry: null, currentDir: '/home', pinned: [], clipboardEntry: null })
    expect(nothingSelected.find((i) => i.id === 'properties')?.disabled).toBe(true)

    const selected = buildFileActions({
      entry: { path: '/home/report.txt', isDir: false },
      currentDir: '/home',
      pinned: [],
      clipboardEntry: null,
    })
    expect(selected.find((i) => i.id === 'properties')?.disabled).toBe(false)
  })

  it('always disables Open With, since no native app picker is wired up yet', () => {
    const items = buildFileActions({
      entry: { path: '/home/report.txt', isDir: false },
      currentDir: '/home',
      pinned: [],
      clipboardEntry: null,
    })
    expect(items.find((i) => i.id === 'open-with')?.disabled).toBe(true)
  })
})
