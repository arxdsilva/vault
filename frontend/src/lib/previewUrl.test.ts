import { describe, expect, it } from 'vitest'
import { previewUrl } from './previewUrl'

describe('previewUrl', () => {
  it('builds a /preview URL from an absolute path', () => {
    expect(previewUrl('/Users/alex/photo.png')).toBe('/preview/Users/alex/photo.png')
  })

  it('percent-encodes each path segment without escaping the separators', () => {
    expect(previewUrl('/Users/alex/my report.txt')).toBe('/preview/Users/alex/my%20report.txt')
  })

  it('encodes special characters within a segment', () => {
    expect(previewUrl('/Users/alex/100% done.txt')).toBe('/preview/Users/alex/100%25%20done.txt')
  })
})
