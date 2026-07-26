// A name is "hidden" by the same convention every Unix tool and Finder use:
// it starts with a dot. Windows has its own hidden-attribute bit, but this
// app only sees names/paths from the backend, not file attributes, so the
// dot convention is the only signal available cross-platform.
export function isHiddenName(name: string): boolean {
  return name.startsWith('.')
}

// Filters entries whose name starts with "." unless showHidden is set
// (spec: hidden files/folders are hidden by default, toggleable in
// Settings).
export function filterHiddenEntries<T extends { name: string }>(entries: T[], showHidden: boolean): T[] {
  if (showHidden) return entries
  return entries.filter((e) => !isHiddenName(e.name))
}
