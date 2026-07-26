import { Check } from 'lucide-react'
import { defaultModelFor, type Provider } from '../lib/llmClient'
import { ACCENT_PRESETS } from '../store/settings'

interface SettingsPanelProps {
  accent: string
  onSelectAccent: (value: string) => void
  provider: Provider
  onSelectProvider: (provider: Provider) => void
  apiKey: string
  onApiKeyChange: (apiKey: string) => void
  model: string
  onModelChange: (model: string) => void
  onClose: () => void
}

// Minimal settings popover anchored under the gear button (spec §4.4):
// accent color presets, plus the AI Assistant's provider/API key/model
// (spec §10) — one settings entry point rather than a second surface.
export function SettingsPanel({
  accent,
  onSelectAccent,
  provider,
  onSelectProvider,
  apiKey,
  onApiKeyChange,
  model,
  onModelChange,
  onClose,
}: SettingsPanelProps) {
  return (
    <div className="fixed inset-0 z-50" onClick={onClose}>
      <div
        onClick={(e) => e.stopPropagation()}
        className="absolute top-14 right-3 flex w-72 flex-col gap-4 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-3 shadow-lg"
      >
        <div className="flex flex-col gap-2">
          <h2 className="text-xs font-medium tracking-wide text-[var(--muted)] uppercase">Accent color</h2>
          <div className="flex flex-wrap gap-2">
            {ACCENT_PRESETS.map((preset) => (
              <button
                key={preset.value}
                type="button"
                title={preset.name}
                aria-label={preset.name}
                aria-pressed={accent === preset.value}
                onClick={() => onSelectAccent(preset.value)}
                style={{ backgroundColor: preset.value }}
                className="flex h-7 w-7 items-center justify-center rounded-full transition-transform duration-150 hover:scale-110"
              >
                {accent === preset.value && <Check size={14} className="text-white" strokeWidth={3} />}
              </button>
            ))}
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h2 className="text-xs font-medium tracking-wide text-[var(--muted)] uppercase">AI Assistant</h2>

          <label className="flex flex-col gap-1 text-xs text-[var(--muted)]">
            Provider
            <select
              value={provider}
              onChange={(e) => onSelectProvider(e.target.value as Provider)}
              className="h-8 rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-[var(--surface)] px-2 text-sm text-[var(--fg)] outline-none"
            >
              <option value="anthropic">Claude</option>
              <option value="openai">ChatGPT</option>
            </select>
          </label>

          <label className="flex flex-col gap-1 text-xs text-[var(--muted)]">
            API key
            <input
              type="password"
              value={apiKey}
              onChange={(e) => onApiKeyChange(e.target.value)}
              placeholder="sk-…"
              autoComplete="off"
              className="h-8 rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-[var(--surface)] px-2 text-sm text-[var(--fg)] outline-none focus:border-[var(--accent)]"
            />
          </label>

          <label className="flex flex-col gap-1 text-xs text-[var(--muted)]">
            Model
            <input
              type="text"
              value={model}
              onChange={(e) => onModelChange(e.target.value)}
              placeholder={defaultModelFor(provider)}
              className="h-8 rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-[var(--surface)] px-2 text-sm text-[var(--fg)] outline-none focus:border-[var(--accent)]"
            />
          </label>

          <p className="text-xs text-[var(--muted)]">
            Stored locally on this device. Never sent anywhere except {provider === 'anthropic' ? 'Anthropic' : 'OpenAI'}'s API.
          </p>
        </div>
      </div>
    </div>
  )
}
