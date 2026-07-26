import { useEffect, useRef } from 'react'

const FOCUSABLE_SELECTOR =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

// Standard modal/menu focus pattern, shared by every overlay (CommandPalette,
// ContextMenu, SettingsPanel, PreviewOverlay) so it's implemented once
// rather than four times: moves focus into the container on mount (first
// focusable descendant, or the container itself — give it tabIndex={-1}),
// traps Tab/Shift+Tab from leaving the container while active, and restores
// focus to whatever had it before on unmount. Doesn't handle Escape/outside
// -click closing — callers keep their own onClose semantics for that.
export function useFocusTrap<T extends HTMLElement>(active = true) {
  const containerRef = useRef<T | null>(null)

  useEffect(() => {
    if (!active) return
    const previouslyFocused = document.activeElement as HTMLElement | null

    const container = containerRef.current
    if (container) {
      const first = container.querySelector<HTMLElement>(FOCUSABLE_SELECTOR)
      ;(first ?? container).focus()
    }

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key !== 'Tab' || !containerRef.current) return
      const focusable = Array.from(containerRef.current.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR))
      if (focusable.length === 0) {
        e.preventDefault()
        return
      }
      const first = focusable[0]
      const last = focusable[focusable.length - 1]
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault()
        last.focus()
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault()
        first.focus()
      }
    }
    document.addEventListener('keydown', handleKeyDown, true)

    return () => {
      document.removeEventListener('keydown', handleKeyDown, true)
      previouslyFocused?.focus()
    }
  }, [active])

  return containerRef
}
