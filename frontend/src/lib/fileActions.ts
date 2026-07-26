export interface FileActionEntry {
  path: string
  isDir: boolean
}

export interface FileActionContext {
  entry: FileActionEntry | null
  currentDir: string | null
  pinned: string[]
  clipboardEntry: FileActionEntry | null
}

export interface FileActionItem {
  id: string
  label: string
  disabled: boolean
}

// Shared action list for the command palette (Cmd/Ctrl+K) and the
// right-click context menu (spec §5.5: "All actions also available via
// right-click context menu with the same items").
export function buildFileActions(ctx: FileActionContext): FileActionItem[] {
  const hasEntry = ctx.entry !== null
  const isPinned = ctx.entry !== null && ctx.pinned.includes(ctx.entry.path)

  return [
    { id: 'open', label: 'Open', disabled: !hasEntry },
    { id: 'open-with', label: 'Open With…', disabled: true },
    { id: 'rename', label: 'Rename', disabled: !hasEntry },
    { id: 'move-to', label: 'Move to…', disabled: !hasEntry },
    { id: 'copy', label: 'Copy', disabled: !hasEntry },
    { id: 'paste', label: 'Paste', disabled: ctx.clipboardEntry === null || ctx.currentDir === null },
    { id: 'duplicate', label: 'Duplicate', disabled: !hasEntry },
    { id: 'copy-path', label: 'Copy Path', disabled: !hasEntry },
    { id: 'new-folder', label: 'New Folder', disabled: ctx.currentDir === null },
    { id: 'compress', label: 'Compress', disabled: !hasEntry },
    { id: 'trash', label: 'Move to Trash', disabled: !hasEntry },
    { id: 'pin', label: isPinned ? 'Unpin' : 'Pin', disabled: !hasEntry },
  ]
}
