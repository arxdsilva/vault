import { Archive, Copy, Files, FolderInput, Link2, Pin, PinOff, Trash2, X } from 'lucide-react'

interface SelectionActionBarProps {
  count: number
  isPinned: boolean
  onRunAction: (id: string) => void
  onClear: () => void
}

const ACTIONS = [
  { id: 'move-to', label: 'Move to…', Icon: FolderInput },
  { id: 'copy', label: 'Copy', Icon: Copy },
  { id: 'duplicate', label: 'Duplicate', Icon: Files },
  { id: 'copy-path', label: 'Copy Path', Icon: Link2 },
  { id: 'compress', label: 'Compress', Icon: Archive },
]

// Gmail-style selection toolbar: once one or more files are selected
// (mouse, Shift/Cmd-click, or the "x" keyboard toggle in FileList/GridView),
// this floats below the top bar with the bulk actions valid for a
// multi-item selection (see buildFileActions in lib/fileActions.ts —
// Open/Rename/Properties are deliberately excluded here since they only
// make sense for exactly one item and stay in the context menu instead).
export function SelectionActionBar({ count, isPinned, onRunAction, onClear }: SelectionActionBarProps) {
  if (count === 0) return null

  return (
    <div
      role="toolbar"
      aria-label="Selection actions"
      className="fixed top-16 left-1/2 z-40 flex -translate-x-1/2 items-center gap-1 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-1 pl-3 shadow-lg"
    >
      <span className="mr-1 text-xs font-medium whitespace-nowrap text-[var(--muted)]">{count} selected</span>

      {ACTIONS.map(({ id, label, Icon }) => (
        <button
          key={id}
          type="button"
          aria-label={label}
          title={label}
          onClick={() => onRunAction(id)}
          className="rounded-[calc(var(--radius)-2px)] p-1.5 text-[var(--fg)] transition-colors duration-150 hover:bg-[var(--surface)]"
        >
          <Icon size={15} aria-hidden="true" />
        </button>
      ))}

      <button
        type="button"
        aria-label={isPinned ? 'Unpin' : 'Pin'}
        title={isPinned ? 'Unpin' : 'Pin'}
        onClick={() => onRunAction('pin')}
        className="rounded-[calc(var(--radius)-2px)] p-1.5 text-[var(--fg)] transition-colors duration-150 hover:bg-[var(--surface)]"
      >
        {isPinned ? <PinOff size={15} aria-hidden="true" /> : <Pin size={15} aria-hidden="true" />}
      </button>

      <button
        type="button"
        aria-label="Move to Trash"
        title="Move to Trash"
        onClick={() => onRunAction('trash')}
        className="rounded-[calc(var(--radius)-2px)] p-1.5 text-[var(--fg)] transition-colors duration-150 hover:bg-[var(--surface)]"
      >
        <Trash2 size={15} aria-hidden="true" />
      </button>

      <div className="mx-1 h-5 w-px bg-[var(--border)]" />

      <button
        type="button"
        aria-label="Clear selection"
        title="Clear selection (Escape)"
        onClick={onClear}
        className="rounded-[calc(var(--radius)-2px)] p-1.5 text-[var(--muted)] transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)]"
      >
        <X size={15} aria-hidden="true" />
      </button>
    </div>
  )
}
