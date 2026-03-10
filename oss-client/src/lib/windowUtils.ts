// eslint-disable-next-line no-restricted-globals
export const isBrowser = () => typeof window !== 'undefined'

export type SafeWindow = Window | undefined

/**
 * Returns `window` on the client, otherwise `undefined` on the server.
 */
// eslint-disable-next-line no-restricted-globals
export const getWindow = (): SafeWindow => (isBrowser() ? window : undefined)

export const getDocument = (): Document | undefined =>
  isBrowser() ? document : undefined

export const getNavigator = (): Navigator | undefined =>
  isBrowser() ? navigator : undefined

export function withWindow<T>(
  fn: (w: Window) => T,
  fallback?: T,
): T | undefined {
  const w = getWindow()
  return w ? fn(w) : fallback
}

export function getLocation() {
  const w = getWindow()
  return w?.location
    ? w.location
    : {
        hash: '',
        hostname: '',
        href: '',
        pathname: '',
        origin: '',
        port: '',
        protocol: '',
        search: '',
      }
}

/**
 * When the __ENV script tag is inserted via React hydration (DOM manipulation)
 * rather than the browser's HTML parser, the browser does NOT execute it.
 * This fallback finds the script in the DOM and evaluates its content manually.
 */
type WindowWithEnv = Window & { __ENV?: Record<string, unknown> }

function hydrateEnvFromDOM(): void {
  if (!isBrowser()) return
  // eslint-disable-next-line no-restricted-globals
  if ((window as WindowWithEnv).__ENV) return

  const scripts = Array.from(document.querySelectorAll('script'))
  for (const script of scripts) {
    const text = script.textContent || ''
    if (text.includes("window['__ENV']")) {
      const match = text.match(/window\['__ENV'\]\s*=\s*(\{.*\})/)
      if (match?.[1]) {
        try {
          // eslint-disable-next-line no-restricted-globals
          ;(window as WindowWithEnv).__ENV = JSON.parse(match[1])
        } catch {
          // Malformed JSON — ignore
        }
      }
      break
    }
  }
}

/**
 * Gets a runtime environment variable from window.__ENV
 * These are environment variables made public via next-runtime-env
 */
export function getRuntimeEnv(key: string): string | undefined {
  return withWindow((w) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    if (!(w as any).__ENV) {
      hydrateEnvFromDOM()
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (w as any)?.__ENV?.[key]
  })
}

/**
 * Gets a runtime environment variable from window.__ENV with a fallback value
 */
export function getRuntimeEnvWithFallback(
  key: string,
  fallback: string,
): string {
  return getRuntimeEnv(key) || fallback
}
