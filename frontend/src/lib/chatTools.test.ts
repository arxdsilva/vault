import { afterEach, describe, expect, it, vi } from 'vitest'
import { CHAT_TOOLS, executeTool } from './chatTools'

vi.mock('./vault', () => ({
  vault: {
    list: vi.fn(async () => [{ name: 'a.txt', path: '/home/a.txt', isDir: false, size: 1, modTime: '' }]),
    search: vi.fn(async () => [{ name: 'a.txt', path: '/home/a.txt', isDir: false }]),
    move: vi.fn(async () => {}),
    rename: vi.fn(async () => {}),
    paste: vi.fn(async () => '/home/dest/a.txt'),
    newFolder: vi.fn(async () => '/home/Untitled'),
    delete: vi.fn(async () => {}),
    listPinned: vi.fn(async () => ['/home']),
  },
}))

import { vault } from './vault'

describe('CHAT_TOOLS', () => {
  it('names exactly the 9 supported tools', () => {
    expect(CHAT_TOOLS.map((t) => t.name)).toEqual([
      'list_directory',
      'search',
      'read_file',
      'move_file',
      'rename_file',
      'paste_into',
      'new_folder',
      'delete_file',
      'list_pinned',
    ])
  })
})

describe('executeTool', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.clearAllMocks()
  })

  it('list_directory dispatches to vault.list', async () => {
    const result = await executeTool('list_directory', { path: '/home' })
    expect(vault.list).toHaveBeenCalledWith('/home')
    expect(result).toEqual({ entries: [{ name: 'a.txt', path: '/home/a.txt', isDir: false, size: 1, modTime: '' }] })
  })

  it('search dispatches to vault.search, defaulting limit to 20', async () => {
    await executeTool('search', { query: 'report' })
    expect(vault.search).toHaveBeenCalledWith('report', 20)
  })

  it('search passes through an explicit limit', async () => {
    await executeTool('search', { query: 'report', limit: 5 })
    expect(vault.search).toHaveBeenCalledWith('report', 5)
  })

  it('read_file fetches the /preview route and returns its text', async () => {
    const fetchMock = vi.fn(async () => new Response('hello', { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await executeTool('read_file', { path: '/home/note.txt' })

    expect(fetchMock).toHaveBeenCalledWith('/preview/home/note.txt')
    expect(result).toEqual({ content: 'hello' })
  })

  it('read_file throws when the fetch fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('nope', { status: 404 })),
    )

    await expect(executeTool('read_file', { path: '/home/missing.txt' })).rejects.toThrow('HTTP 404')
  })

  it('move_file dispatches to vault.move', async () => {
    const result = await executeTool('move_file', { src: '/a', dst: '/b' })
    expect(vault.move).toHaveBeenCalledWith('/a', '/b')
    expect(result).toEqual({ ok: true })
  })

  it('rename_file dispatches to vault.rename', async () => {
    await executeTool('rename_file', { path: '/a', newPath: '/b' })
    expect(vault.rename).toHaveBeenCalledWith('/a', '/b')
  })

  it('paste_into dispatches to vault.paste and returns the resulting path', async () => {
    const result = await executeTool('paste_into', { src: '/a', destDir: '/dest' })
    expect(vault.paste).toHaveBeenCalledWith('/a', '/dest')
    expect(result).toEqual({ path: '/home/dest/a.txt' })
  })

  it('new_folder dispatches to vault.newFolder and returns the created path', async () => {
    const result = await executeTool('new_folder', { parent: '/home', name: 'Untitled' })
    expect(vault.newFolder).toHaveBeenCalledWith('/home', 'Untitled')
    expect(result).toEqual({ path: '/home/Untitled' })
  })

  it('delete_file dispatches to vault.delete', async () => {
    const result = await executeTool('delete_file', { path: '/home/a.txt' })
    expect(vault.delete).toHaveBeenCalledWith('/home/a.txt')
    expect(result).toEqual({ ok: true })
  })

  it('list_pinned dispatches to vault.listPinned', async () => {
    const result = await executeTool('list_pinned', {})
    expect(result).toEqual({ paths: ['/home'] })
  })

  it('an unknown tool name throws', async () => {
    await expect(executeTool('not_a_real_tool', {})).rejects.toThrow('unknown tool')
  })
})
