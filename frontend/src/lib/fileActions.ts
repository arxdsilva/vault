export interface FileActionEntry {
  path: string
  isDir: boolean
}

export interface FileActionContext {
  entries: FileActionEntry[]
  currentDir: string | null
  pinned: string[]
  clipboardEntries: FileActionEntry[]
}

export interface FileActionItem {
  id: string
  label: string
  disabled: boolean
}

// Shared action list for the command palette (Cmd/Ctrl+K) and the
// right-click context menu (spec §5.5: "All actions also available via
// right-click context menu with the same items"). Most actions operate on
// however many entries are selected (spec: multi-select); a few — Open,
// Rename, Properties — only make sense for exactly one.
export function buildFileActions(ctx: FileActionContext): FileActionItem[] {
  const hasSelection = ctx.entries.length > 0
  const isSingle = ctx.entries.length === 1
  const isPinned = hasSelection && ctx.entries.every((e) => ctx.pinned.includes(e.path))

  return [
    { id: 'open', label: 'Open', disabled: !isSingle },
    { id: 'open-with', label: 'Open With…', disabled: true },
    { id: 'rename', label: 'Rename', disabled: !isSingle },
    { id: 'move-to', label: 'Move to…', disabled: !hasSelection },
    { id: 'copy', label: 'Copy', disabled: !hasSelection },
    { id: 'paste', label: 'Paste', disabled: ctx.clipboardEntries.length === 0 || ctx.currentDir === null },
    { id: 'duplicate', label: 'Duplicate', disabled: !hasSelection },
    { id: 'copy-path', label: 'Copy Path', disabled: !hasSelection },
    { id: 'properties', label: 'Properties', disabled: !isSingle },
    { id: 'new-folder', label: 'New Folder', disabled: ctx.currentDir === null },
    { id: 'compress', label: 'Compress', disabled: !hasSelection },
    { id: 'trash', label: 'Move to Trash', disabled: !hasSelection },
    { id: 'pin', label: isPinned ? 'Unpin' : 'Pin', disabled: !hasSelection },
  ]
}
