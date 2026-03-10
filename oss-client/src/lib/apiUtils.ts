'use client'

// assumes localDev is using devbox server
export const BASE_ADMIN_API = process.env.NEXT_PUBLIC_ADMIN_API_URL || '/'
import { createConnectTransport } from '@connectrpc/connect-web'
import { QueryClient } from '@tanstack/react-query'
import { isAuthError } from '@/lib/errorUtils'
import { getWindow } from './windowUtils'

export const isLocalDev = process.env.NODE_ENV !== 'production'
const isDevBox = BASE_ADMIN_API?.includes('localhost')
/** Default redirect path after login (must include /v2/). */
const DEFAULT_REDIRECT_PATH = '/v2/projects'

/** Login URL with optional redirect. When omitted, uses current page in browser, else default path. */
export function getLoginUrl(redirect?: string): string {
  const baseUrl = (BASE_ADMIN_API || '/').replace(/\/$/, '')
  const currentHref = getWindow()?.location?.href
  const redirectUrl =
    redirect ??
    (currentHref ||
      (baseUrl ? `${baseUrl}${DEFAULT_REDIRECT_PATH}` : DEFAULT_REDIRECT_PATH))
  return `${baseUrl || ''}/login?redirect_url=${encodeURIComponent(redirectUrl)}`
}

export function createTransport(_useBinaryFormat: boolean) {
  return createConnectTransport({
    baseUrl: BASE_ADMIN_API,
    // useBinaryFormat,
    ...(!isDevBox && {
      fetch: (input, init) => {
        return fetch(input, {
          ...init,
          credentials: 'include', // This enables sending cookies in cross-origin requests
        })
      },
    }),
  })
}

/** Default transport (binary format). Used for all Connect RPC requests. */
export const finalTransport = createTransport(true)

// Create QueryClient outside component to prevent memory leaks
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Add some reasonable defaults
      refetchOnWindowFocus: false,
      // Don't retry 401s: let the cache subscriber run refreshAuth, then refetch or show login modal
      retry: (failureCount, error) => !isAuthError(error) && failureCount < 3,
    },
  },
})
