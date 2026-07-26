import { Component, type ErrorInfo, type ReactNode } from 'react'

interface ErrorBoundaryProps {
  children: ReactNode
}

interface ErrorBoundaryState {
  error: Error | null
}

// Last-resort safety net: no single error anywhere in the tree should take
// down the whole app (e.g. a bound Go call rejecting from a TCC-protected
// path). Catches render/lifecycle errors that slip past local handling and
// shows a recoverable fallback instead of a blank, frozen window — React
// unmounts the whole tree on an uncaught error otherwise, with no boundary
// to stop at.
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Vault crashed:', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex h-full w-full flex-col items-center justify-center gap-3 bg-[var(--bg)] p-6 text-center text-[var(--fg)]">
          <p className="text-sm font-medium">Something went wrong.</p>
          <p className="max-w-md text-xs break-words text-[var(--muted)]">{this.state.error.message}</p>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="rounded-[var(--radius)] border border-[var(--border)] bg-[var(--surface)] px-3 py-1.5 text-xs hover:bg-[var(--border)]"
          >
            Reload
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
