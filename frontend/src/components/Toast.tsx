import { useEffect } from 'react'
import { useToast } from '../store/toast'

const AUTO_DISMISS_MS = 4000

// Non-blocking error toast (spec §5.6 "non-blocking...toast", adapted here
// for surfacing failed operations rather than copy/move progress).
export function Toast() {
  const { message, dismiss } = useToast()

  useEffect(() => {
    if (!message) return
    const timer = setTimeout(dismiss, AUTO_DISMISS_MS)
    return () => clearTimeout(timer)
  }, [message, dismiss])

  if (!message) return null

  return (
    <div
      role="alert"
      className="fixed bottom-6 left-1/2 z-50 max-w-md -translate-x-1/2 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--surface)] px-4 py-2 text-sm text-[var(--fg)] shadow-lg"
    >
      {message}
    </div>
  )
}
