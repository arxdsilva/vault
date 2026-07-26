import { describe, expect, it } from 'vitest'
import { formatRelativeTime } from './relativeTime'

const now = new Date('2026-01-15T12:00:00Z')

describe('formatRelativeTime', () => {
  it('reports "just now" for under a minute', () => {
    expect(formatRelativeTime('2026-01-15T11:59:30Z', now)).toBe('just now')
  })

  it('reports minutes ago for under an hour', () => {
    expect(formatRelativeTime('2026-01-15T11:55:00Z', now)).toBe('5m ago')
  })

  it('reports hours ago for under a day', () => {
    expect(formatRelativeTime('2026-01-15T09:00:00Z', now)).toBe('3h ago')
  })

  it('reports days ago for under a week', () => {
    expect(formatRelativeTime('2026-01-13T12:00:00Z', now)).toBe('2d ago')
  })

  it('falls back to a short date beyond a week', () => {
    expect(formatRelativeTime('2025-12-01T12:00:00Z', now)).toBe('Dec 1')
  })
})
