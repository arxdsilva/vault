import { ShieldAlert } from 'lucide-react'
import { useFocusTrap } from '../hooks/useFocusTrap'

interface PermissionPromptProps {
  onOpenSettings: () => void
  onDismiss: () => void
}

const HEADING_ID = 'permission-prompt-heading'

// Full Disk Access explainer for macOS: without it, Desktop/Documents/
// Downloads/removable & network volumes each trigger their own separate
// system consent prompt the first time they're visited. Granting it once
// here avoids that one-by-one nagging. Dismissable rather than blocking —
// the app remains usable without it, just noisier.
export function PermissionPrompt({ onOpenSettings, onDismiss }: PermissionPromptProps) {
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
              Grant Full Disk Access
            </h1>
            <p className="text-xs text-[var(--muted)]">
              Without it, macOS asks for permission separately the first time you open Desktop, Documents,
              Downloads, or an external drive. Granting Full Disk Access once avoids the repeated prompts.
            </p>
            <p className="text-xs text-[var(--muted)]">
              After enabling Vault in the list that opens, quit and reopen the app.
            </p>
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
            Open System Settings
          </button>
        </div>
      </div>
    </div>
  )
}
