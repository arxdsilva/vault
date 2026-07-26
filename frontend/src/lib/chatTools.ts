import { previewUrl } from './previewUrl'
import { vault } from './vault'

export interface ToolSpec {
  name: string
  description: string
  inputSchema: {
    type: 'object'
    properties: Record<string, { type: string; description?: string }>
    required?: string[]
  }
}

// The tool surface the embedded chat panel offers an LLM (spec §10). This is
// the TS mirror of internal/mcp/tools.go's tool list — the same 9 operations,
// intentionally duplicated across the language boundary rather than
// codegen'd, since 9 small entries doesn't justify a shared-schema build
// step. Tool *execution* reuses the existing vault.ts calls (and, for
// read_file, the same /preview fileserver route PreviewOverlay already
// fetches from) rather than adding a new backend surface.
export const CHAT_TOOLS: ToolSpec[] = [
  {
    name: 'list_directory',
    description: 'List the entries (files and folders) of a directory.',
    inputSchema: {
      type: 'object',
      properties: { path: { type: 'string', description: 'directory to list' } },
      required: ['path'],
    },
  },
  {
    name: 'search',
    description: 'Fuzzy-search indexed files by name, ranked by match quality, recency, and frequency of use.',
    inputSchema: {
      type: 'object',
      properties: {
        query: { type: 'string', description: 'search query' },
        limit: { type: 'number', description: 'maximum number of results (default 20)' },
      },
      required: ['query'],
    },
  },
  {
    name: 'read_file',
    description: "Read a text file's contents.",
    inputSchema: {
      type: 'object',
      properties: { path: { type: 'string', description: 'path of the file to read' } },
      required: ['path'],
    },
  },
  {
    name: 'move_file',
    description: 'Move a file or folder to a new path.',
    inputSchema: {
      type: 'object',
      properties: {
        src: { type: 'string', description: 'path of the file or folder to move' },
        dst: { type: 'string', description: 'destination path' },
      },
      required: ['src', 'dst'],
    },
  },
  {
    name: 'rename_file',
    description: 'Rename a file or folder in place.',
    inputSchema: {
      type: 'object',
      properties: {
        path: { type: 'string', description: 'path of the file or folder to rename' },
        newPath: { type: 'string', description: 'new path (same parent directory, different name)' },
      },
      required: ['path', 'newPath'],
    },
  },
  {
    name: 'paste_into',
    description:
      "Copy a file or folder into a destination directory, disambiguating the name if one already exists there (like Finder's paste).",
    inputSchema: {
      type: 'object',
      properties: {
        src: { type: 'string', description: 'path of the file or folder to copy' },
        destDir: { type: 'string', description: 'destination directory' },
      },
      required: ['src', 'destDir'],
    },
  },
  {
    name: 'new_folder',
    description: 'Create a new folder inside a parent directory.',
    inputSchema: {
      type: 'object',
      properties: {
        parent: { type: 'string', description: 'parent directory the new folder is created in' },
        name: { type: 'string', description: 'name of the new folder' },
      },
      required: ['parent', 'name'],
    },
  },
  {
    name: 'delete_file',
    description: 'Move a file or folder to the OS trash. Never permanently deletes.',
    inputSchema: {
      type: 'object',
      properties: { path: { type: 'string', description: 'path of the file or folder to trash' } },
      required: ['path'],
    },
  },
  {
    name: 'list_pinned',
    description: "List the user's pinned folders/files shown in Vault's sidebar.",
    inputSchema: { type: 'object', properties: {} },
  },
]

// Dispatches a tool call (by name, JSON input) to the matching vault.ts
// operation and returns a JSON-serializable result.
export async function executeTool(name: string, input: Record<string, unknown>): Promise<unknown> {
  switch (name) {
    case 'list_directory':
      return { entries: await vault.list(input.path as string) }

    case 'search':
      return { results: await vault.search(input.query as string, (input.limit as number) ?? 20) }

    case 'read_file': {
      const path = input.path as string
      const res = await fetch(previewUrl(path))
      if (!res.ok) throw new Error(`could not read ${path}: HTTP ${res.status}`)
      return { content: await res.text() }
    }

    case 'move_file':
      await vault.move(input.src as string, input.dst as string)
      return { ok: true }

    case 'rename_file':
      await vault.rename(input.path as string, input.newPath as string)
      return { ok: true }

    case 'paste_into':
      return { path: await vault.paste(input.src as string, input.destDir as string) }

    case 'new_folder':
      return { path: await vault.newFolder(input.parent as string, input.name as string) }

    case 'delete_file':
      await vault.delete(input.path as string)
      return { ok: true }

    case 'list_pinned':
      return { paths: await vault.listPinned() }

    default:
      throw new Error(`unknown tool: ${name}`)
  }
}
