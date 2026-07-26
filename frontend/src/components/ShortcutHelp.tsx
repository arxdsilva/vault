import { useFocusTrap } from '../hooks/useFocusTrap'

interface ShortcutHelpProps {
  onClose: () => void
}

interface Shortcut {
  keys: string
  label: string
}

interface ShortcutGroup {
  title: string
  shortcuts: Shortcut[]
}

const GROUPS: ShortcutGroup[] = [
  {
    title: 'Navigation',
    shortcuts: [
      { keys: '↑ ↓ ← →', label: 'Move selection' },
      { keys: 'Enter', label: 'Open' },
      { keys: 'Space', label: 'Quick preview' },
      { keys: 'Cmd/Ctrl+↑', label: 'Go to parent folder' },
      { keys: 'Escape', label: 'Close / clear search' },
    ],
  },
  {
    title: 'Search',
    shortcuts: [{ keys: '/', label: 'Focus search' }],
  },
  {
    title: 'File actions',
    shortcuts: [
      { keys: 'C', label: 'Copy' },
      { keys: 'Cmd/Ctrl+V', label: 'Paste' },
      { keys: 'Delete', label: 'Move to Trash' },
      { keys: 'F2', label: 'Rename' },
      { keys: 'Cmd/Ctrl+N', label: 'New folder' },
      { keys: 'Cmd/Ctrl+Z', label: 'Undo' },
    ],
  },
  {
    title: 'Other',
    shortcuts: [
      { keys: 'Cmd/Ctrl+K', label: 'Command palette' },
      { keys: 'Menu / Shift+F10', label: 'Context menu' },
      { keys: '?', label: 'Show this help' },
    ],
  },
]

const HEADING_ID = 'shortcut-help-heading'

// Keyboard shortcuts cheat sheet (spec §5.7/§5.8), modeled on Gmail's "?"
// help dialog. Its visibility is gated by store/settings.ts's
// shortcutHelperEnabled — the shortcuts themselves always work regardless.
export function ShortcutHelp({ onClose }: ShortcutHelpProps) {
  const containerRef = useFocusTrap<HTMLDivElement>()

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20" onClick={onClose}>
      <div
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={HEADING_ID}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => {
          if (e.key === 'Escape') {
            e.preventDefault()
            onClose()
          }
        }}
        className="flex max-h-[80vh] w-full max-w-md flex-col gap-4 overflow-y-auto rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-4 shadow-lg"
      >
        <div className="flex items-center justify-between">
          <h1 id={HEADING_ID} className="text-sm font-medium">
            Keyboard shortcuts
          </h1>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-[calc(var(--radius)-2px)] px-2 py-1 text-xs text-[var(--muted)] hover:bg-[var(--surface)] hover:text-[var(--fg)]"
          >
            Close
          </button>
        </div>

        {GROUPS.map((group) => (
          <div key={group.title} className="flex flex-col gap-1.5">
            <h2 className="text-xs font-medium tracking-wide text-[var(--muted)] uppercase">{group.title}</h2>
            <dl className="flex flex-col gap-1">
              {group.shortcuts.map((s) => (
                <div key={s.label} className="flex items-center justify-between gap-4 text-sm">
                  <dt className="text-[var(--fg)]">{s.label}</dt>
                  <dd className="shrink-0 rounded-[calc(var(--radius)-4px)] border border-[var(--border)] bg-[var(--surface)] px-1.5 py-0.5 font-mono text-xs text-[var(--muted)]">
                    {s.keys}
                  </dd>
                </div>
              ))}
            </dl>
          </div>
        ))}
      </div>
    </div>
  )
}
