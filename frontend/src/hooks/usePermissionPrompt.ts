import { useEffect, useState } from 'react'
import { vault, type PermissionStatus } from '../lib/vault'

// Checks the OS-level filesystem permission once on startup so the app can
// guide the user to the right settings surface itself, instead of letting
// each protected folder trigger its own separate consent prompt on first
// visit (macOS Full Disk Access) or a silent write failure (Windows
// Controlled Folder Access, when the user has turned it on). Copy is
// entirely OS-owned (see core.PermissionStatus) — this hook just renders
// whatever the backend reports. Handling the prompt once (either button)
// persists that choice on the backend, so it doesn't nag on every launch
// even if the permission is still ungranted — a folder that actually fails
// to open still surfaces guidance inline regardless (see useDirectory). A
// no-op on Linux, where the backend always reports granted.
export function usePermissionPrompt() {
  const [status, setStatus] = useState<PermissionStatus | null>(null)

  useEffect(() => {
    vault
      .permissionStatus()
      .then(async (s) => {
        if (s.granted) return
        const dismissed = await vault.permissionPromptDismissed().catch(() => false)
        if (!dismissed) setStatus(s)
      })
      .catch(() => {})
  }, [])

  function openSettings() {
    vault.openPermissionSettings().catch(() => {})
    vault.dismissPermissionPrompt().catch(() => {})
    setStatus(null)
  }

  function dismiss() {
    vault.dismissPermissionPrompt().catch(() => {})
    setStatus(null)
  }

  return { status, openSettings, dismiss }
}
