import { useEffect, useRef, useState } from 'react'
import { allIndices, clampIndex, clampSelection, rangeIndices, toggleIndex } from '../lib/multiSelect'

interface ClickModifiers {
  metaKey: boolean
  ctrlKey: boolean
  shiftKey: boolean
}

// Checkbox-style multi-mark over a list of `count` items, shared by
// FileList and GridView (Gmail-style: browsing never disturbs a build-up
// selection, only explicit marking does).
//
// `focus` is a pure cursor: plain click and plain arrow-key movement only
// move it, they never mark anything. `marked` is the explicit set built via
// "x", the checkbox, Shift+range, or Cmd/Ctrl+click/toggle — what bulk
// actions operate on. When `marked` is empty, callers (FileList/GridView's
// onSelectionChange) fall back to treating the focused item as the
// effective single-item selection, so single-file actions (Delete, Copy,
// rename, Space-preview) keep working without requiring an explicit mark
// first — only building an actual multi-item selection requires marking.
//
// clearSignal is an opaque value (e.g. an incrementing counter) that, when
// it changes, empties `marked` — how a parent (App.tsx's selection action
// bar "Clear" button / Escape) clears this hook's own state instead of just
// its own mirrored copy of it.
export function useMultiSelect(count: number, clearSignal?: unknown) {
  const [marked, setMarked] = useState<Set<number>>(new Set())
  const [focus, setFocus] = useState(0)
  const anchorRef = useRef(0)
  const isFirstClearSignal = useRef(true)

  useEffect(() => {
    // Skip the mount-time run — clearSignal's initial value is a change
    // from "no value yet", not a real clear request.
    if (isFirstClearSignal.current) {
      isFirstClearSignal.current = false
      return
    }
    setMarked(new Set())
  }, [clearSignal])

  useEffect(() => {
    const nextFocus = clampIndex(focus, count)
    setFocus(nextFocus)
    setMarked((s) => clampSelection(s, count))
    if (anchorRef.current >= count) anchorRef.current = clampIndex(anchorRef.current, count)
    // Only reacts to the list's length changing (a new directory, an item
    // appearing/disappearing) — `focus` is read for its current value, not
    // watched, since re-running on every focus change would fight the
    // click/keyboard handlers below that already own `focus`.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [count])

  // Moves the cursor without marking anything — plain click and plain
  // arrow-key movement.
  function focusOnly(index: number) {
    anchorRef.current = index
    setFocus(index)
  }

  // Marks exactly one item, replacing any existing mark — used when
  // right-clicking a row that isn't already part of the marked set.
  function markSingle(index: number) {
    anchorRef.current = index
    setFocus(index)
    setMarked(new Set([index]))
  }

  function toggle(index: number) {
    anchorRef.current = index
    setFocus(index)
    setMarked((s) => toggleIndex(s, index))
  }

  function extendTo(index: number) {
    setFocus(index)
    setMarked(new Set(rangeIndices(anchorRef.current, index)))
  }

  // Arrow-key navigation: plain move only repositions the cursor, Shift+move
  // extends the marked range from the last plain/toggle anchor (matches
  // Shift-click). Returns the resulting focus index so callers can e.g.
  // scroll it into view.
  function moveFocus(delta: number, extend: boolean): number {
    const next = clampIndex(focus + delta, count)
    if (extend) extendTo(next)
    else focusOnly(next)
    return next
  }

  function handleClick(index: number, e: ClickModifiers) {
    if (e.shiftKey) extendTo(index)
    else if (e.metaKey || e.ctrlKey) toggle(index)
    else focusOnly(index)
  }

  // Right-click: keep an existing mark if the clicked row is already part
  // of it (so the context menu acts on the whole marked set), otherwise
  // mark just this row — matches Finder.
  function selectForContextMenu(index: number): void {
    if (!marked.has(index)) markSingle(index)
  }

  function selectAll() {
    setMarked(allIndices(count))
  }

  return { marked, focus, handleClick, moveFocus, selectForContextMenu, selectAll, toggle }
}
