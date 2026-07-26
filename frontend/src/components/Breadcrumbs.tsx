import { ChevronRight, Home } from 'lucide-react'
import { segmentsOf } from '../store/navigation'

interface BreadcrumbsProps {
  dir: string
  onNavigate: (path: string) => void
}

// Path strip (36px, spec §4.2). Hidden on the launch/search screen — only
// mounted once a folder is being browsed.
export function Breadcrumbs({ dir, onNavigate }: BreadcrumbsProps) {
  const segments = segmentsOf(dir)

  return (
    <div className="flex h-9 shrink-0 items-center gap-1 overflow-x-auto border-b border-[var(--border)] px-3 text-sm text-[var(--muted)]">
      <button
        type="button"
        onClick={() => onNavigate('/')}
        className="flex items-center gap-1 rounded-[calc(var(--radius)-2px)] px-1.5 py-1 transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)]"
      >
        <Home size={14} />
      </button>
      {segments.map((segment, i) => (
        <div key={segment.path} className="flex shrink-0 items-center gap-1">
          <ChevronRight size={14} className="shrink-0" />
          <button
            type="button"
            onClick={() => onNavigate(segment.path)}
            className={`rounded-[calc(var(--radius)-2px)] px-1.5 py-1 whitespace-nowrap transition-colors duration-150 hover:bg-[var(--surface)] hover:text-[var(--fg)] ${
              i === segments.length - 1 ? 'text-[var(--fg)]' : ''
            }`}
          >
            {segment.label}
          </button>
        </div>
      ))}
    </div>
  )
}
