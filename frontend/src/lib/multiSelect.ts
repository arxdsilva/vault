// Pure index-set math backing multi-select in FileList/GridView (Finder/
// Explorer-style: click selects one, Cmd/Ctrl-click toggles, Shift-click or
// Shift-Arrow extends a range from the last plain/toggle selection).

export function rangeIndices(a: number, b: number): number[] {
  const lo = Math.min(a, b)
  const hi = Math.max(a, b)
  const out: number[] = []
  for (let i = lo; i <= hi; i++) out.push(i)
  return out
}

// Drops indices that fell out of range after the underlying list shrank
// (e.g. a file was deleted out from under an active multi-selection).
export function clampSelection(selected: ReadonlySet<number>, count: number): Set<number> {
  const next = new Set<number>()
  selected.forEach((i) => {
    if (i >= 0 && i < count) next.add(i)
  })
  return next
}

export function clampIndex(index: number, count: number): number {
  return Math.min(Math.max(index, 0), Math.max(count - 1, 0))
}

export function toggleIndex(selected: ReadonlySet<number>, index: number): Set<number> {
  const next = new Set(selected)
  if (next.has(index)) next.delete(index)
  else next.add(index)
  return next
}

export function allIndices(count: number): Set<number> {
  const next = new Set<number>()
  for (let i = 0; i < count; i++) next.add(i)
  return next
}

export function sortedIndices(selected: ReadonlySet<number>): number[] {
  return Array.from(selected).sort((a, b) => a - b)
}

// Maps selected indices to items in ascending order, dropping any index
// that's out of range for `list`. Selection state and the entries array it
// indexes into are updated by separate effects (clamping vs. re-deriving
// the selected items), so there's a render in between where `selected` can
// still hold indices from before the list just shrank — without this guard
// that produces `undefined` list entries instead of a shorter, valid list.
export function entriesAt<T>(selected: ReadonlySet<number>, list: readonly T[]): T[] {
  const out: T[] = []
  sortedIndices(selected).forEach((i) => {
    const item = list[i]
    if (item !== undefined) out.push(item)
  })
  return out
}
