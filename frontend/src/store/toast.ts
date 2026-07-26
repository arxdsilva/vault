import { create } from 'zustand'

interface ToastState {
  message: string | null
  show: (message: string) => void
  dismiss: () => void
}

// Surfaces file-system errors to the user instead of swallowing them
// (architecture.mdc: "Surface file-system errors...don't swallow or panic").
export const useToast = create<ToastState>((set) => ({
  message: null,
  show: (message) => set({ message }),
  dismiss: () => set({ message: null }),
}))
