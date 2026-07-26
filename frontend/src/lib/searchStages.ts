import type { IndexEntry } from './vault'

export interface SearchStage {
  label: string | null // null => unscoped (no folder open, e.g. the Home screen)
  results: IndexEntry[] // cumulative results revealed as of this stage
}

function nameOf(path: string): string {
  const parts = path.split('/').filter(Boolean)
  return parts[parts.length - 1] || path
}

function isUnder(path: string, dir: string): boolean {
  return path === dir || path.startsWith(dir.endsWith('/') ? dir : `${dir}/`)
}

function parentOf(dir: string): string {
  const trimmed = dir.replace(/\/+$/, '')
  const parent = trimmed.split('/').slice(0, -1).join('/')
  return parent === '' ? '/' : parent
}

// Ancestor chain from dir up to (and including) home, capped at maxLevels
// entries so a very deeply nested folder doesn't produce an excessive
// number of expansion stages.
export function ancestorChain(dir: string, home: string, maxLevels: number): string[] {
  const chain: string[] = []
  let current = dir
  while (chain.length < maxLevels) {
    chain.push(current)
    if (current === home) break
    const parent = parentOf(current)
    if (parent === current) break
    current = parent
  }
  return chain
}

// Builds the sequence of progressive-reveal stages for a folder-scoped
// search (spec §5.2): the current folder's matches first, then each parent
// folder in turn, ending with everything else. Each stage holds the
// *cumulative* results revealed so far, so the caller can just render
// stage.results directly at every step. When currentDir is null (no folder
// open — the Home/launch screen), there's no context to scope by, so the
// whole result set comes back as a single unscoped stage.
export function buildSearchStages(
  entries: IndexEntry[],
  currentDir: string | null,
  home: string,
  maxLevels: number,
): SearchStage[] {
  if (!currentDir) {
    return [{ label: null, results: entries }]
  }

  const chain = ancestorChain(currentDir, home, maxLevels)
  const revealed = new Set<string>()
  const stages: SearchStage[] = []
  let cumulative: IndexEntry[] = []

  chain.forEach((dir, i) => {
    const additions = entries.filter((e) => !revealed.has(e.path) && isUnder(e.path, dir))
    additions.forEach((e) => revealed.add(e.path))
    cumulative = [...cumulative, ...additions]
    stages.push({
      label: i === 0 ? `Searching in ${nameOf(dir)}` : `Expanding to ${nameOf(dir)}`,
      results: cumulative,
    })
  })

  const rest = entries.filter((e) => !revealed.has(e.path))
  if (rest.length > 0) {
    cumulative = [...cumulative, ...rest]
    stages.push({ label: 'Expanding to everywhere', results: cumulative })
  }

  return stages
}
