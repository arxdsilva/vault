import { useEffect, useRef, useState } from 'react'
import { allIndices, clampIndex, clampSelection, rangeIndices, toggleIndex } from '../lib/multiSelect'

interface ClickModifiers {
  metaKey: boolean
  ctrlKey: boolean
  shiftKey: boolean
}

// Finder/Explorer-style multi-select over a list of `count` items, shared by
// FileList and GridView. Tracks which indices are selected, plus a single
// `focus` index (the aria-activedescendant / Enter-to-open / arrow-key
// target) which is always one of the selected indices except right after a
// Cmd/Ctrl-click deselects the item focus was on.
export function useMultiSelect(count: number) {
  // Index 0 starts selected (once there's anything to select) so arrow keys/
  // Enter work immediately without an initial click — matches the previous
  // single-select behavior both FileList and GridView had.
  const [selected, setSelected] = useState<Set<number>>(new Set([0]))
  const [focus, setFocus] = useState(0)
  const anchorRef = useRef(0)

  useEffect(() => {
    const nextFocus = clampIndex(focus, count)
    setFocus(nextFocus)
    setSelected((s) => {
      const clamped = clampSelection(s, count)
      // Falling out of range entirely (e.g. the selected tail of a folder
      // got shorter) clamps down to the nearest valid item instead of
      // leaving nothing selected/focusable.
      return clamped.size > 0 || count === 0 ? clamped : new Set([nextFocus])
    })
    if (anchorRef.current >= count) anchorRef.current = clampIndex(anchorRef.current, count)
    // Only reacts to the list's length changing (a new directory, an item
    // appearing/disappearing) — `focus` is read for its current value, not
    // watched, since re-running on every focus change would fight the
    // click/keyboard handlers below that already own `focus`.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [count])

  function selectSingle(index: number) {
    anchorRef.current = index
    setFocus(index)
    setSelected(new Set([index]))
  }

  function toggle(index: number) {
    anchorRef.current = index
    setFocus(index)
    setSelected((s) => toggleIndex(s, index))
  }

  function extendTo(index: number) {
    setFocus(index)
    setSelected(new Set(rangeIndices(anchorRef.current, index)))
  }

  // Arrow-key navigation: plain move collapses to the new focus, Shift+move
  // extends the range from the last plain/toggle anchor (matches Shift-click).
  // Returns the resulting focus index so callers can e.g. scroll it into view.
  function moveFocus(delta: number, extend: boolean): number {
    const next = clampIndex(focus + delta, count)
    if (extend) extendTo(next)
    else selectSingle(next)
    return next
  }

  function handleClick(index: number, e: ClickModifiers) {
    if (e.shiftKey) extendTo(index)
    else if (e.metaKey || e.ctrlKey) toggle(index)
    else selectSingle(index)
  }

  // Right-click: keep an existing multi-selection if the clicked row is
  // already part of it (so the context menu acts on the whole selection),
  // otherwise replace the selection with just this row — matches Finder.
  function selectForContextMenu(index: number): void {
    if (!selected.has(index)) selectSingle(index)
  }

  function selectAll() {
    setSelected(allIndices(count))
  }

  return { selected, focus, handleClick, moveFocus, selectSingle, selectForContextMenu, selectAll }
}
