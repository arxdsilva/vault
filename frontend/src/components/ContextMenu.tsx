import { useLayoutEffect, useState } from 'react'
import { buildFileActions, type FileActionContext } from '../lib/fileActions'
import { useFocusTrap } from '../hooks/useFocusTrap'

interface ContextMenuProps {
  x: number
  y: number
  ctx: FileActionContext
  onRunAction: (id: string) => void
  onClose: () => void
}

// Space around the window edge the menu should never sit flush against.
const EDGE_MARGIN = 8

// Right-click (or Shift+F10/Menu key — spec: keyboard-first) menu: same
// items as the command palette. role="menu"/"menuitem" with roving focus
// (arrow keys move real DOM focus between items, not just a highlight) and
// a focus trap (see useFocusTrap) so Tab can't leave it while open.
export function ContextMenu({ x, y, ctx, onRunAction, onClose }: ContextMenuProps) {
  const items = buildFileActions(ctx)
  const containerRef = useFocusTrap<HTMLDivElement>()
  // The menu is a DOM node confined to the app window, unlike a native
  // OS context menu — it can never render outside those bounds. So instead
  // it flips to the opposite side of the click when it would overflow,
  // the same way native menus do near a screen edge, keeping every item
  // reachable rather than clipped off-window and permanently unclickable.
  const [pos, setPos] = useState({ left: x, top: y })

  useLayoutEffect(() => {
    const el = containerRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()

    let left = x
    if (left + rect.width > window.innerWidth - EDGE_MARGIN) {
      left = x - rect.width // open to the left of the click instead
    }
    left = Math.min(Math.max(left, EDGE_MARGIN), Math.max(EDGE_MARGIN, window.innerWidth - rect.width - EDGE_MARGIN))

    let top = y
    if (top + rect.height > window.innerHeight - EDGE_MARGIN) {
      top = y - rect.height // open above the click instead
    }
    top = Math.min(Math.max(top, EDGE_MARGIN), Math.max(EDGE_MARGIN, window.innerHeight - rect.height - EDGE_MARGIN))

    setPos({ left, top })
  }, [x, y, items.length])

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
        style={{ position: 'absolute', left: pos.left, top: pos.top }}
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
