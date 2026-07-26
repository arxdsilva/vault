import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { codeLanguageOf, previewKindOf } from '../lib/previewKind'
import { previewUrl } from '../lib/previewUrl'

interface PreviewEntry {
  name: string
  path: string
}

interface PreviewOverlayProps {
  entry: PreviewEntry
  onClose: () => void
  onPrev: () => void
  onNext: () => void
}

const MAX_TEXT_BYTES = 2 * 1024 * 1024

type TextState =
  | { status: 'loading' }
  | { status: 'ready'; html: string }
  | { status: 'plain'; content: string }
  | { status: 'too-large' }
  | { status: 'error'; message: string }

// Quick preview overlay (Space, spec §5.4): images/video/audio/PDF stream
// straight from the Go fileserver; text/code is fetched and, for code,
// syntax-highlighted via Shiki. Arrow keys move between files; Esc closes.
export function PreviewOverlay({ entry, onClose, onPrev, onNext }: PreviewOverlayProps) {
  const kind = previewKindOf(entry.name)
  const url = previewUrl(entry.path)
  const [text, setText] = useState<TextState>({ status: 'loading' })

  useEffect(() => {
    if (kind !== 'text' && kind !== 'code') return
    let cancelled = false
    setText({ status: 'loading' })

    fetch(url)
      .then(async (res) => {
        const length = Number(res.headers.get('content-length') ?? '0')
        if (length > MAX_TEXT_BYTES) {
          if (!cancelled) setText({ status: 'too-large' })
          return
        }
        const content = await res.text()
        if (cancelled) return
        if (kind === 'code') {
          const { codeToHtml } = await import('shiki')
          const html = await codeToHtml(content, { lang: codeLanguageOf(entry.name), theme: 'github-dark' })
          if (!cancelled) setText({ status: 'ready', html })
        } else {
          setText({ status: 'plain', content })
        }
      })
      .catch((err) => {
        if (!cancelled) setText({ status: 'error', message: String(err) })
      })

    return () => {
      cancelled = true
    }
  }, [kind, url, entry.name])

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        e.preventDefault()
        e.stopPropagation()
        onClose()
      } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
        e.preventDefault()
        e.stopPropagation()
        onPrev()
      } else if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
        e.preventDefault()
        e.stopPropagation()
        onNext()
      }
    }
    // Capture phase: take priority over the underlying list's own arrow-key
    // navigation and the app-level Escape handler while the overlay is open.
    window.addEventListener('keydown', handleKeyDown, true)
    return () => window.removeEventListener('keydown', handleKeyDown, true)
  }, [onClose, onPrev, onNext])

  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-black/90" onClick={onClose}>
      <div className="flex h-12 shrink-0 items-center justify-between px-4 text-sm text-white/80">
        <span className="truncate">{entry.name}</span>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close preview"
          className="rounded p-1 hover:bg-white/10"
        >
          <X size={18} />
        </button>
      </div>

      <div
        onClick={(e) => e.stopPropagation()}
        className="flex min-h-0 flex-1 items-center justify-center overflow-auto p-6"
      >
        {kind === 'image' && (
          <img src={url} alt={entry.name} className="max-h-full max-w-full object-contain" />
        )}
        {kind === 'video' && <video src={url} controls autoPlay className="max-h-full max-w-full" />}
        {kind === 'audio' && <audio src={url} controls className="w-full max-w-xl" />}
        {kind === 'pdf' && <iframe src={url} title={entry.name} className="h-full w-full bg-white" />}
        {(kind === 'text' || kind === 'code') && (
          <div className="h-full w-full max-w-4xl overflow-auto rounded-[var(--radius)] bg-[#0d1117] p-4 text-sm text-white/90">
            {text.status === 'loading' && <p>Loading…</p>}
            {text.status === 'too-large' && <p>File is too large to preview.</p>}
            {text.status === 'error' && <p>Couldn't load this file: {text.message}</p>}
            {text.status === 'plain' && <pre className="whitespace-pre-wrap">{text.content}</pre>}
            {text.status === 'ready' && <div dangerouslySetInnerHTML={{ __html: text.html }} />}
          </div>
        )}
        {kind === 'unsupported' && <p className="text-white/70">No preview available for this file.</p>}
      </div>
    </div>
  )
}
