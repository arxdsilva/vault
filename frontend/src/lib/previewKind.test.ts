import { describe, expect, it } from 'vitest'
import { codeLanguageOf, previewKindOf } from './previewKind'

describe('previewKindOf', () => {
  it.each([
    ['photo.png', 'image'],
    ['photo.JPG', 'image'],
    ['clip.mp4', 'video'],
    ['song.mp3', 'audio'],
    ['doc.pdf', 'pdf'],
    ['main.go', 'code'],
    ['index.tsx', 'code'],
    ['notes.txt', 'text'],
    ['Makefile', 'text'],
    ['archive.zip', 'unsupported'],
    ['app.exe', 'unsupported'],
  ])('classifies %s as %s', (name, want) => {
    expect(previewKindOf(name)).toBe(want)
  })
})

describe('codeLanguageOf', () => {
  it('maps known extensions to a Shiki language id', () => {
    expect(codeLanguageOf('main.go')).toBe('go')
    expect(codeLanguageOf('index.tsx')).toBe('tsx')
    expect(codeLanguageOf('script.py')).toBe('python')
  })

  it('falls back to plain text for an unknown extension', () => {
    expect(codeLanguageOf('README')).toBe('text')
  })
})
