import { buildFileActions, type FileActionContext } from '../lib/fileActions'

interface ContextMenuProps {
  x: number
  y: number
  ctx: FileActionContext
  onRunAction: (id: string) => void
  onClose: () => void
}

// Right-click menu (spec §5.5): same items as the command palette.
export function ContextMenu({ x, y, ctx, onRunAction, onClose }: ContextMenuProps) {
  const items = buildFileActions(ctx)

  return (
    <div
      className="fixed inset-0 z-50"
      onClick={onClose}
      onContextMenu={(e) => {
        e.preventDefault()
        onClose()
      }}
    >
      <div
        style={{ position: 'absolute', left: x, top: y }}
        onClick={(e) => e.stopPropagation()}
        className="flex min-w-[180px] flex-col overflow-hidden rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] py-1 shadow-lg"
      >
        {items.map((item) => (
          <button
            key={item.id}
            type="button"
            disabled={item.disabled}
            onClick={() => {
              onRunAction(item.id)
              onClose()
            }}
            className="flex h-8 w-full items-center px-3 text-left text-sm transition-colors duration-150 hover:bg-[var(--surface)] disabled:opacity-40 disabled:hover:bg-transparent"
          >
            {item.label}
          </button>
        ))}
      </div>
    </div>
  )
}
