import { buildFileActions, type FileActionContext } from '../lib/fileActions'
import { useFocusTrap } from '../hooks/useFocusTrap'

interface ContextMenuProps {
  x: number
  y: number
  ctx: FileActionContext
  onRunAction: (id: string) => void
  onClose: () => void
}

// Right-click (or Shift+F10/Menu key — spec: keyboard-first) menu: same
// items as the command palette. role="menu"/"menuitem" with roving focus
// (arrow keys move real DOM focus between items, not just a highlight) and
// a focus trap (see useFocusTrap) so Tab can't leave it while open.
export function ContextMenu({ x, y, ctx, onRunAction, onClose }: ContextMenuProps) {
  const items = buildFileActions(ctx)
  const containerRef = useFocusTrap<HTMLDivElement>()

  function menuItems(): HTMLElement[] {
    return Array.from(
      containerRef.current?.querySelectorAll<HTMLElement>('[role="menuitem"]:not([disabled])') ?? [],
    )
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault()
      onClose()
      return
    }
    const focusables = menuItems()
    if (focusables.length === 0) return
    const currentIndex = focusables.indexOf(document.activeElement as HTMLElement)
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      focusables[(currentIndex + 1 + focusables.length) % focusables.length].focus()
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      focusables[(currentIndex - 1 + focusables.length) % focusables.length].focus()
    } else if (e.key === 'Home') {
      e.preventDefault()
      focusables[0].focus()
    } else if (e.key === 'End') {
      e.preventDefault()
      focusables[focusables.length - 1].focus()
    }
  }

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
        ref={containerRef}
        role="menu"
        aria-label="Actions"
        tabIndex={-1}
        style={{ position: 'absolute', left: x, top: y }}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
        className="flex min-w-[180px] flex-col overflow-hidden rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] py-1 shadow-lg"
      >
        {items.map((item) => (
          <button
            key={item.id}
            role="menuitem"
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
