import { describe, expect, it } from 'vitest'
import { ancestorChain, buildSearchStages } from './searchStages'
import type { IndexEntry } from './vault'

function entry(path: string): IndexEntry {
  return { name: path.split('/').pop() ?? path, path, isDir: false }
}

describe('ancestorChain', () => {
  it('returns just [dir] when dir is already home', () => {
    expect(ancestorChain('/Users/alex', '/Users/alex', 4)).toEqual(['/Users/alex'])
  })

  it('walks up from dir to home', () => {
    expect(ancestorChain('/Users/alex/Projects/ClientX', '/Users/alex', 4)).toEqual([
      '/Users/alex/Projects/ClientX',
      '/Users/alex/Projects',
      '/Users/alex',
    ])
  })

  it('stops at maxLevels even if home is further up', () => {
    expect(ancestorChain('/a/b/c/d/e', '/', 2)).toEqual(['/a/b/c/d/e', '/a/b/c/d'])
  })

  it('stops at the filesystem root if reached before home', () => {
    expect(ancestorChain('/a/b', '/somewhere/else', 10)).toEqual(['/a/b', '/a', '/'])
  })
})

describe('buildSearchStages', () => {
  it('returns a single unscoped stage when there is no current directory', () => {
    const entries = [entry('/Users/alex/a.txt'), entry('/Users/alex/Projects/b.txt')]
    const stages = buildSearchStages(entries, null, '/Users/alex', 4)
    expect(stages).toEqual([{ label: null, results: entries }])
  })

  it('stages the current folder first, then expands to parents', () => {
    const inScope = entry('/Users/alex/Projects/ClientX/report.txt')
    const inParent = entry('/Users/alex/Projects/other.txt')
    const inHome = entry('/Users/alex/elsewhere.txt')
    const entries = [inScope, inParent, inHome]

    const stages = buildSearchStages(entries, '/Users/alex/Projects/ClientX', '/Users/alex', 4)

    expect(stages).toHaveLength(3)
    expect(stages[0]).toEqual({ label: 'Searching in ClientX', results: [inScope] })
    expect(stages[1]).toEqual({ label: 'Expanding to Projects', results: [inScope, inParent] })
    expect(stages[2]).toEqual({ label: 'Expanding to alex', results: [inScope, inParent, inHome] })
  })

  it('never lists the same result twice across stages', () => {
    const inScope = entry('/Users/alex/Projects/ClientX/report.txt')
    const stages = buildSearchStages([inScope], '/Users/alex/Projects/ClientX', '/Users/alex', 4)

    for (const stage of stages) {
      const paths = stage.results.map((r) => r.path)
      expect(new Set(paths).size).toBe(paths.length)
    }
  })

  it('adds a final "everywhere" stage when maxLevels is capped before reaching home', () => {
    const inScope = entry('/a/b/c/report.txt')
    const outside = entry('/elsewhere/report.txt')
    const stages = buildSearchStages([inScope, outside], '/a/b/c', '/', 1)

    expect(stages).toHaveLength(2)
    expect(stages[0]).toEqual({ label: 'Searching in c', results: [inScope] })
    expect(stages[1]).toEqual({ label: 'Expanding to everywhere', results: [inScope, outside] })
  })

  it('does not add an "everywhere" stage when the ancestor chain already covers every result', () => {
    const inScope = entry('/Users/alex/Projects/ClientX/report.txt')
    const stages = buildSearchStages([inScope], '/Users/alex/Projects/ClientX', '/Users/alex', 4)

    expect(stages[stages.length - 1].label).not.toBe('Expanding to everywhere')
  })
})
