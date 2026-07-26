import { useEffect, useState } from 'react'
import { vault } from '../lib/vault'

// Checks macOS Full Disk Access once on startup so the app can guide the
// user to System Settings itself instead of letting each protected folder
// (Desktop, Documents, Downloads, …) trigger its own separate consent
// prompt on first visit. Dismissing only hides the prompt for this
// session — it re-checks (and re-prompts if still ungranted) on next launch.
// A no-op on Windows/Linux, where the backend always reports granted.
export function usePermissionPrompt() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    vault
      .permissionStatus()
      .then((status) => setVisible(!status.granted))
      .catch(() => {})
  }, [])

  function openSettings() {
    vault.openPermissionSettings().catch(() => {})
    setVisible(false)
  }

  function dismiss() {
    setVisible(false)
  }

  return { visible, openSettings, dismiss }
}
