type Toast = {
  id: string
  type: 'success' | 'error' | 'info'
  message: string
}

type Listener = (t: Toast) => void
const listeners = new Set<Listener>()

export function showToast(toast: Omit<Toast, 'id'>): void {
  const t: Toast = { ...toast, id: crypto.randomUUID() }
  listeners.forEach(fn => fn(t))
}

export function subscribeToasts(fn: Listener): () => void {
  listeners.add(fn)
  return () => listeners.delete(fn)
}
