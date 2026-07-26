import { create } from 'zustand'
import { sendMessage, type ChatMessage, type Provider } from '../lib/llmClient'

const STORAGE_KEY = 'vault:chat'

interface StoredSettings {
  provider: Provider
  apiKey: string
  model: string
}

const DEFAULT_STORED: StoredSettings = { provider: 'anthropic', apiKey: '', model: '' }

function loadStored(): StoredSettings {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return DEFAULT_STORED
  try {
    return { ...DEFAULT_STORED, ...JSON.parse(raw) }
  } catch {
    return DEFAULT_STORED
  }
}

function saveStored(s: StoredSettings) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(s))
}

interface ChatState {
  provider: Provider
  apiKey: string
  model: string
  messages: ChatMessage[]
  sending: boolean
  error: string | null
  setProvider: (provider: Provider) => void
  setApiKey: (apiKey: string) => void
  setModel: (model: string) => void
  send: (text: string) => Promise<void>
  clear: () => void
}

const stored = loadStored()

// Embedded AI assistant (spec §10): a bring-your-own-key chat panel that can
// call the same 9 tools the MCP server exposes (see lib/chatTools.ts).
// Provider/apiKey/model persist to localStorage the same manual way
// store/settings.ts persists the accent color; conversation messages are
// session-only and never written to disk.
export const useChat = create<ChatState>((set, get) => ({
  provider: stored.provider,
  apiKey: stored.apiKey,
  model: stored.model,
  messages: [],
  sending: false,
  error: null,

  setProvider: (provider) => {
    saveStored({ provider, apiKey: get().apiKey, model: get().model })
    set({ provider })
  },

  setApiKey: (apiKey) => {
    saveStored({ provider: get().provider, apiKey, model: get().model })
    set({ apiKey })
  },

  setModel: (model) => {
    saveStored({ provider: get().provider, apiKey: get().apiKey, model })
    set({ model })
  },

  send: async (text) => {
    const { provider, apiKey, model, messages } = get()
    if (!apiKey.trim() || !text.trim()) return

    const userMessage: ChatMessage = { role: 'user', text }
    set({ messages: [...messages, userMessage], sending: true, error: null })

    try {
      const reply = await sendMessage({ provider, apiKey, model, history: messages, userText: text })
      set((s) => ({ messages: [...s.messages, reply], sending: false }))
    } catch (err) {
      set({ sending: false, error: String(err) })
    }
  },

  clear: () => set({ messages: [], error: null }),
}))
