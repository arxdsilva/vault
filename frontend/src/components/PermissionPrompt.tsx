import { ShieldAlert } from 'lucide-react'
import { useFocusTrap } from '../hooks/useFocusTrap'
import type { PermissionStatus } from '../lib/vault'

interface PermissionPromptProps {
  status: PermissionStatus
  onOpenSettings: () => void
  onDismiss: () => void
}

const HEADING_ID = 'permission-prompt-heading'

// Explains an ungranted OS-level filesystem permission (macOS Full Disk
// Access, or Windows Controlled Folder Access when the user has turned it
// on) and links straight to the right settings surface. All copy comes
// from status — each platform adapter owns its own terminology (see
// core.PermissionStatus) — so this component has no OS-specific text.
// Dismissable rather than blocking — the app remains usable without it,
// just noisier (or, on Windows, with silently failed writes).
export function PermissionPrompt({ status, onOpenSettings, onDismiss }: PermissionPromptProps) {
  const containerRef = useFocusTrap<HTMLDivElement>()

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/20" onClick={onDismiss}>
      <div
        ref={containerRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={HEADING_ID}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => {
          if (e.key === 'Escape') {
            e.preventDefault()
            onDismiss()
          }
        }}
        className="flex w-full max-w-sm flex-col gap-3 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-4 shadow-lg"
      >
        <div className="flex items-start gap-3">
          <ShieldAlert size={20} className="mt-0.5 shrink-0 text-[var(--accent)]" aria-hidden="true" />
          <div className="flex flex-col gap-1">
            <h1 id={HEADING_ID} className="text-sm font-medium">
              {status.title}
            </h1>
            {status.message.split('\n\n').map((paragraph) => (
              <p key={paragraph} className="text-xs text-[var(--muted)]">
                {paragraph}
              </p>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onDismiss}
            className="rounded-[calc(var(--radius)-2px)] px-2 py-1 text-xs text-[var(--muted)] hover:bg-[var(--surface)] hover:text-[var(--fg)]"
          >
            Not now
          </button>
          <button
            type="button"
            onClick={onOpenSettings}
            className="rounded-[calc(var(--radius)-2px)] bg-[var(--accent)] px-3 py-1 text-xs font-medium text-white"
          >
            {status.settingsLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
