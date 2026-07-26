import { describe, expect, it } from 'vitest'
import { filterHiddenEntries, isHiddenName } from './hiddenFiles'

describe('isHiddenName', () => {
  it('treats a leading dot as hidden', () => {
    expect(isHiddenName('.localized')).toBe(true)
    expect(isHiddenName('.git')).toBe(true)
  })

  it('treats a name without a leading dot as visible', () => {
    expect(isHiddenName('report.txt')).toBe(false)
    expect(isHiddenName('Photos')).toBe(false)
  })
})

describe('filterHiddenEntries', () => {
  const entries = [{ name: 'report.txt' }, { name: '.localized' }, { name: 'Photos' }, { name: '.git' }]

  it('drops dotfile entries by default', () => {
    const result = filterHiddenEntries(entries, false)
    expect(result.map((e) => e.name)).toEqual(['report.txt', 'Photos'])
  })

  it('keeps every entry when showHidden is true', () => {
    const result = filterHiddenEntries(entries, true)
    expect(result).toEqual(entries)
  })

  it('returns an empty array unchanged', () => {
    expect(filterHiddenEntries([], false)).toEqual([])
  })
})
