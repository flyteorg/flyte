/**
 * Notifier for auth-expired (401) so the LoginPanel can show when any code path
 * detects expired auth (e.g. refreshAuth or interceptors), not only query cache updates.
 */
type Listener = () => void
const listeners: Set<Listener> = new Set()

export function subscribeAuthExpired(listener: Listener): () => void {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

export function notifyAuthExpired(): void {
  listeners.forEach((l) => l())
}
