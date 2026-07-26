import { describe, expect, it, beforeEach } from 'vitest'
import { segmentsOf, useNavigation } from './navigation'

describe('segmentsOf', () => {
  it('splits an absolute path into breadcrumb segments with cumulative paths', () => {
    expect(segmentsOf('/Users/alex/Projects')).toEqual([
      { label: 'Users', path: '/Users' },
      { label: 'alex', path: '/Users/alex' },
      { label: 'Projects', path: '/Users/alex/Projects' },
    ])
  })

  it('returns an empty array for the root path', () => {
    expect(segmentsOf('/')).toEqual([])
  })

  it('ignores a trailing slash', () => {
    expect(segmentsOf('/Users/alex/')).toEqual([
      { label: 'Users', path: '/Users' },
      { label: 'alex', path: '/Users/alex' },
    ])
  })
})

describe('useNavigation', () => {
  beforeEach(() => {
    useNavigation.setState({ currentDir: null, sortByDir: {} })
  })

  it('starts on the launch screen (currentDir = null)', () => {
    expect(useNavigation.getState().currentDir).toBeNull()
  })

  it('navigate sets the current directory', () => {
    useNavigation.getState().navigate('/Users/alex')
    expect(useNavigation.getState().currentDir).toBe('/Users/alex')
  })

  it('navigate(null) returns to the launch screen', () => {
    useNavigation.getState().navigate('/Users/alex')
    useNavigation.getState().navigate(null)
    expect(useNavigation.getState().currentDir).toBeNull()
  })

  describe('parentOf', () => {
    it('returns the parent of a nested directory', () => {
      expect(useNavigation.getState().parentOf('/Users/alex/Projects')).toBe('/Users/alex')
    })

    it('returns root for a top-level directory', () => {
      expect(useNavigation.getState().parentOf('/Users')).toBe('/')
    })
  })

  describe('sort preference', () => {
    it('defaults to sorting by name', () => {
      expect(useNavigation.getState().sortFor('/Users/alex')).toBe('name')
    })

    it('remembers a sort choice per directory', () => {
      useNavigation.getState().setSort('/Users/alex', 'size')
      expect(useNavigation.getState().sortFor('/Users/alex')).toBe('size')
      expect(useNavigation.getState().sortFor('/Users/other')).toBe('name')
    })
  })
})
