import { useRef, useEffect } from 'react'

interface RenameInputProps {
  initialName: string
  onSubmit: (name: string) => void
  onCancel: () => void
}

// Inline rename field (F2 / command palette "Rename", spec §5.6). Enter or
// blur commits; Escape cancels.
export function RenameInput({ initialName, onSubmit, onCancel }: RenameInputProps) {
  const ref = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    const input = ref.current
    if (!input) return
    input.focus()
    // Pre-select the name, stopping short of the extension, matching Finder.
    const dot = initialName.lastIndexOf('.')
    input.setSelectionRange(0, dot > 0 ? dot : initialName.length)
  }, [initialName])

  function commit() {
    const value = ref.current?.value.trim()
    if (value) onSubmit(value)
    else onCancel()
  }

  return (
    <input
      ref={ref}
      type="text"
      defaultValue={initialName}
      onClick={(e) => e.stopPropagation()}
      onDoubleClick={(e) => e.stopPropagation()}
      onBlur={commit}
      onKeyDown={(e) => {
        e.stopPropagation()
        if (e.key === 'Enter') {
          e.preventDefault()
          commit()
        } else if (e.key === 'Escape') {
          e.preventDefault()
          onCancel()
        }
      }}
      className="h-6 min-w-0 flex-1 rounded-[calc(var(--radius)-2px)] border border-[var(--accent)] bg-[var(--surface)] px-1 text-sm outline-none"
    />
  )
}
