import { useEffect, useRef, useState } from 'react'
import { Send, X } from 'lucide-react'
import { useChat } from '../store/chat'
import type { ToolCallRecord } from '../lib/llmClient'

interface ChatPanelProps {
  onClose: () => void
}

function ToolCallRow({ call }: { call: ToolCallRecord }) {
  return (
    <details className="rounded-[calc(var(--radius)-2px)] bg-[var(--bg)] px-2 py-1 text-xs text-[var(--muted)]">
      <summary className="cursor-pointer select-none">
        Used <code>{call.name}</code>
      </summary>
      <pre className="mt-1 overflow-x-auto whitespace-pre-wrap">
        {JSON.stringify({ input: call.input, result: call.result }, null, 2)}
      </pre>
    </details>
  )
}

// Docked AI assistant panel (spec §10): a persistent right-side column, not
// a floating overlay — the assistant is meant to feel like a first-class
// collaborator. Tool execution reuses lib/chatTools.ts's executeTool via
// lib/llmClient.ts's agentic loop, driven by store/chat.ts.
export function ChatPanel({ onClose }: ChatPanelProps) {
  const { messages, sending, error, apiKey, send } = useChat()
  const [input, setInput] = useState('')
  const scrollRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [messages, sending])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!input.trim() || sending) return
    send(input)
    setInput('')
  }

  return (
    <div className="flex w-80 shrink-0 flex-col border-l border-[var(--border)] bg-[var(--surface)]">
      <div className="flex h-11 shrink-0 items-center justify-between border-b border-[var(--border)] px-3">
        <span className="text-sm font-medium">Assistant</span>
        <button
          type="button"
          aria-label="Close assistant"
          onClick={onClose}
          className="rounded-[var(--radius)] p-1 text-[var(--muted)] transition-colors duration-150 hover:text-[var(--fg)]"
        >
          <X size={16} aria-hidden="true" />
        </button>
      </div>

      <div
        ref={scrollRef}
        role="log"
        aria-live="polite"
        aria-label="Conversation"
        className="flex flex-1 flex-col gap-2 overflow-y-auto p-3"
      >
        {messages.length === 0 && (
          <p className="text-sm text-[var(--muted)]">
            {apiKey ? 'Ask about your files.' : 'Set an API key in Settings to start chatting.'}
          </p>
        )}
        {messages.map((m, i) => (
          <div key={i} className={`flex flex-col gap-1 ${m.role === 'user' ? 'items-end' : 'items-start'}`}>
            <div
              className={`max-w-[90%] rounded-[var(--radius)] px-3 py-2 text-sm whitespace-pre-wrap ${
                m.role === 'user' ? 'bg-[var(--accent)] text-white' : 'bg-[var(--bg)] text-[var(--fg)]'
              }`}
            >
              {m.text}
            </div>
            {m.toolCalls && m.toolCalls.length > 0 && (
              <div className="flex w-full max-w-[90%] flex-col gap-1">
                {m.toolCalls.map((call, j) => (
                  <ToolCallRow key={j} call={call} />
                ))}
              </div>
            )}
          </div>
        ))}
        {sending && <p className="text-sm text-[var(--muted)]">Thinking…</p>}
        {error && <p className="text-sm text-red-500">{error}</p>}
      </div>

      <form onSubmit={handleSubmit} className="flex shrink-0 gap-2 border-t border-[var(--border)] p-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={apiKey ? 'Message the assistant…' : 'Set an API key in Settings first'}
          aria-label="Message the assistant"
          disabled={!apiKey || sending}
          className="h-9 flex-1 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] px-3 text-sm outline-none transition-colors duration-150 focus:border-[var(--accent)] disabled:opacity-50"
        />
        <button
          type="submit"
          disabled={!apiKey || sending || !input.trim()}
          aria-label="Send"
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[var(--radius)] bg-[var(--accent)] text-white transition-opacity duration-150 disabled:opacity-40"
        >
          <Send size={16} aria-hidden="true" />
        </button>
      </form>
    </div>
  )
}
