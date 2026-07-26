import { Check, Monitor, Moon, Sun } from 'lucide-react'
import { defaultModelFor, type Provider } from '../lib/llmClient'
import { useFocusTrap } from '../hooks/useFocusTrap'
import { ACCENT_PRESETS, type Theme } from '../store/settings'

interface SettingsPanelProps {
  accent: string
  onSelectAccent: (value: string) => void
  theme: Theme
  onSelectTheme: (theme: Theme) => void
  shortcutHelperEnabled: boolean
  onSelectShortcutHelperEnabled: (enabled: boolean) => void
  onShowShortcuts: () => void
  provider: Provider
  onSelectProvider: (provider: Provider) => void
  apiKey: string
  onApiKeyChange: (apiKey: string) => void
  model: string
  onModelChange: (model: string) => void
  onClose: () => void
}

const THEME_OPTIONS: { value: Theme; label: string; Icon: typeof Sun }[] = [
  { value: 'system', label: 'System', Icon: Monitor },
  { value: 'light', label: 'Light', Icon: Sun },
  { value: 'dark', label: 'Dark', Icon: Moon },
]

const HEADING_ID = 'settings-heading'

// Minimal settings popover anchored under the gear button (spec §4.4):
// appearance (theme + accent color presets), plus the AI Assistant's
// provider/API key/model (spec §10) — one settings entry point rather than
// a second surface. A dialog (role="dialog", focus trap, Escape to close,
// focus restored to the gear button on close — see useFocusTrap) rather
// than a plain popover div.
export function SettingsPanel({
  accent,
  onSelectAccent,
  theme,
  onSelectTheme,
  shortcutHelperEnabled,
  onSelectShortcutHelperEnabled,
  onShowShortcuts,
  provider,
  onSelectProvider,
  apiKey,
  onApiKeyChange,
  model,
  onModelChange,
  onClose,
}: SettingsPanelProps) {
  const containerRef = useFocusTrap<HTMLDivElement>()

  return (
    <div className="fixed inset-0 z-50" onClick={onClose}>
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
            onClose()
          }
        }}
        className="absolute top-14 right-3 flex w-72 flex-col gap-4 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-3 shadow-lg"
      >
        <h1 id={HEADING_ID} className="sr-only">
          Settings
        </h1>

        <div className="flex flex-col gap-2">
          <h2 className="text-xs font-medium tracking-wide text-[var(--muted)] uppercase">Appearance</h2>
          <div className="flex gap-1 rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-[var(--surface)] p-1">
            {THEME_OPTIONS.map(({ value, label, Icon }) => (
              <button
                key={value}
                type="button"
                aria-label={label}
                aria-pressed={theme === value}
                onClick={() => onSelectTheme(value)}
                className={`flex flex-1 items-center justify-center gap-1.5 rounded-[calc(var(--radius)-4px)] py-1.5 text-xs transition-colors duration-150 ${
                  theme === value
                    ? 'bg-[var(--bg)] text-[var(--fg)] shadow-sm'
                    : 'text-[var(--muted)] hover:text-[var(--fg)]'
                }`}
              >
                <Icon size={13} strokeWidth={1.75} aria-hidden="true" />
                {label}
              </button>
            ))}
          </div>
        </div>

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
                {accent === preset.value && <Check size={14} className="text-white" strokeWidth={3} aria-hidden="true" />}
              </button>
            ))}
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h2 className="text-xs font-medium tracking-wide text-[var(--muted)] uppercase">Shortcuts</h2>
          <div className="flex items-center justify-between gap-2 text-xs text-[var(--muted)]">
            <span>Shortcut helper (press ?)</span>
            <button
              type="button"
              role="switch"
              aria-checked={shortcutHelperEnabled}
              aria-label="Keyboard shortcut helper"
              onClick={() => onSelectShortcutHelperEnabled(!shortcutHelperEnabled)}
              className={`relative h-5 w-9 shrink-0 rounded-full transition-colors duration-150 ${
                shortcutHelperEnabled ? 'bg-[var(--accent)]' : 'bg-[var(--border)]'
              }`}
            >
              <span
                aria-hidden="true"
                className={`absolute top-0.5 h-4 w-4 rounded-full bg-white transition-transform duration-150 ${
                  shortcutHelperEnabled ? 'translate-x-[18px]' : 'translate-x-0.5'
                }`}
              />
            </button>
          </div>
          <button
            type="button"
            onClick={onShowShortcuts}
            className="self-start rounded-[calc(var(--radius)-2px)] border border-[var(--border)] bg-[var(--surface)] px-2 py-1 text-xs text-[var(--fg)] transition-colors duration-150 hover:bg-[var(--border)]"
          >
            View shortcuts
          </button>
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
