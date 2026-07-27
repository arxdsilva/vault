import { useEffect, useState } from 'react'
import { vault, type PermissionStatus } from '../lib/vault'

// Checks the OS-level filesystem permission once on startup so the app can
// guide the user to the right settings surface itself, instead of letting
// each protected folder trigger its own separate consent prompt on first
// visit (macOS Full Disk Access) or a silent write failure (Windows
// Controlled Folder Access, when the user has turned it on). Copy is
// entirely OS-owned (see core.PermissionStatus) — this hook just renders
// whatever the backend reports. Dismissing only hides the prompt for this
// session — it re-checks (and re-prompts if still ungranted) on next
// launch. A no-op on Linux, where the backend always reports granted.
export function usePermissionPrompt() {
  const [status, setStatus] = useState<PermissionStatus | null>(null)

  useEffect(() => {
    vault
      .permissionStatus()
      .then((s) => setStatus(s.granted ? null : s))
      .catch(() => {})
  }, [])

  function openSettings() {
    vault.openPermissionSettings().catch(() => {})
    setStatus(null)
  }

  function dismiss() {
    setStatus(null)
  }

  return { status, openSettings, dismiss }
}
