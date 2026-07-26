import { describe, expect, it } from 'vitest'
import {
  allIndices,
  clampIndex,
  clampSelection,
  entriesAt,
  rangeIndices,
  sortedIndices,
  toggleIndex,
} from './multiSelect'

describe('rangeIndices', () => {
  it('builds an inclusive ascending range', () => {
    expect(rangeIndices(2, 5)).toEqual([2, 3, 4, 5])
  })

  it('works when the anchor is after the target (selecting upward)', () => {
    expect(rangeIndices(5, 2)).toEqual([2, 3, 4, 5])
  })

  it('returns a single index when both ends match', () => {
    expect(rangeIndices(3, 3)).toEqual([3])
  })
})

describe('clampSelection', () => {
  it('drops indices at or beyond the new count', () => {
    const result = clampSelection(new Set([0, 2, 4, 5]), 4)
    expect(result).toEqual(new Set([0, 2]))
  })

  it('leaves an already-valid selection untouched', () => {
    const result = clampSelection(new Set([0, 1]), 4)
    expect(result).toEqual(new Set([0, 1]))
  })

  it('returns an empty set when count is 0', () => {
    expect(clampSelection(new Set([0, 1]), 0)).toEqual(new Set())
  })
})

describe('clampIndex', () => {
  it('clamps to the last valid index', () => {
    expect(clampIndex(10, 4)).toBe(3)
  })

  it('clamps negative indices to 0', () => {
    expect(clampIndex(-1, 4)).toBe(0)
  })

  it('returns 0 for an empty list regardless of input', () => {
    expect(clampIndex(5, 0)).toBe(0)
  })

  it('leaves an in-range index untouched', () => {
    expect(clampIndex(2, 4)).toBe(2)
  })
})

describe('toggleIndex', () => {
  it('adds an index not yet selected', () => {
    expect(toggleIndex(new Set([1]), 2)).toEqual(new Set([1, 2]))
  })

  it('removes an index already selected', () => {
    expect(toggleIndex(new Set([1, 2]), 2)).toEqual(new Set([1]))
  })

  it('does not mutate the input set', () => {
    const input = new Set([1])
    toggleIndex(input, 2)
    expect(input).toEqual(new Set([1]))
  })
})

describe('allIndices', () => {
  it('builds a contiguous set from 0 to count - 1', () => {
    expect(allIndices(3)).toEqual(new Set([0, 1, 2]))
  })

  it('returns an empty set for count 0', () => {
    expect(allIndices(0)).toEqual(new Set())
  })
})

describe('sortedIndices', () => {
  it('sorts numerically, not lexicographically', () => {
    expect(sortedIndices(new Set([10, 2, 1]))).toEqual([1, 2, 10])
  })

  it('returns an empty array for an empty set', () => {
    expect(sortedIndices(new Set())).toEqual([])
  })
})

describe('entriesAt', () => {
  const list = ['a', 'b', 'c']

  it('maps selected indices to items in ascending order', () => {
    expect(entriesAt(new Set([2, 0]), list)).toEqual(['a', 'c'])
  })

  it('drops indices that are out of range for the list', () => {
    // Simulates the list having shrunk out from under a stale selection.
    expect(entriesAt(new Set([0, 5, -1]), list)).toEqual(['a'])
  })

  it('returns an empty array when nothing is selected', () => {
    expect(entriesAt(new Set(), list)).toEqual([])
  })

  it('returns an empty array when every selected index is stale', () => {
    expect(entriesAt(new Set([9, 10]), list)).toEqual([])
  })
})
