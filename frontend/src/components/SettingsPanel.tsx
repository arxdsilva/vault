import { Check } from 'lucide-react'
import { ACCENT_PRESETS } from '../store/settings'

interface SettingsPanelProps {
  accent: string
  onSelectAccent: (value: string) => void
  onClose: () => void
}

// Minimal settings popover anchored under the gear button (spec §4.4):
// accent color presets only, theme itself stays OS-keyed.
export function SettingsPanel({ accent, onSelectAccent, onClose }: SettingsPanelProps) {
  return (
    <div className="fixed inset-0 z-50" onClick={onClose}>
      <div
        onClick={(e) => e.stopPropagation()}
        className="absolute top-14 right-3 flex w-56 flex-col gap-3 rounded-[var(--radius)] border border-[var(--border)] bg-[var(--bg)] p-3 shadow-lg"
      >
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
    </div>
  )
}
