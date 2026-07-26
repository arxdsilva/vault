const MINUTE = 60
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR
const WEEK = 7 * DAY

const monthDayFormatter = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' })

// Formats a past ISO timestamp relative to now, e.g. "2h ago" (spec §5.1
// Recent list). Falls back to a short date beyond a week.
export function formatRelativeTime(iso: string, now: Date = new Date()): string {
  const seconds = Math.max(0, (now.getTime() - new Date(iso).getTime()) / 1000)

  if (seconds < MINUTE) return 'just now'
  if (seconds < HOUR) return `${Math.floor(seconds / MINUTE)}m ago`
  if (seconds < DAY) return `${Math.floor(seconds / HOUR)}h ago`
  if (seconds < WEEK) return `${Math.floor(seconds / DAY)}d ago`
  return monthDayFormatter.format(new Date(iso))
}
