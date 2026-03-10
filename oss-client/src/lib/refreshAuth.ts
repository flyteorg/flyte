import type { QueryClient } from '@tanstack/react-query'
import { getLoginUrl } from './apiUtils'
import { isStrongAuthError } from './errorUtils'
import { notifyAuthExpired } from './authExpiredNotifier'

/** Query key for the /me (identity) request. Only this query's 401 triggers token refresh. */
export const IDENTITY_QUERY_KEY = ['identity'] as const

export type RefreshAuthOptions = {
  queryClient?: QueryClient
}

/**
 * On strong auth error (ConnectError Unauthenticated / HTTP 401): try to
 * refresh tokens by fetching the login endpoint (server may set new cookies
 * from the Okta refresh token). Only shows the login panel if the refresh
 * fetch throws. If refresh succeeds, refetches active queries so they can
 * succeed with the new cookies.
 */
export const refreshAuth = async (
  error?: unknown,
  options?: RefreshAuthOptions,
): Promise<void> => {
  if (!isStrongAuthError(error)) {
    return
  }

  try {
    const loginUrl = getLoginUrl()
    await fetch(loginUrl, {
      method: 'GET',
      headers: { Accept: 'text/html' },
      credentials: 'include',
      redirect: 'follow',
      mode: 'no-cors',
    })
    const { queryClient } = options ?? {}
    if (queryClient?.refetchQueries) {
      await queryClient.refetchQueries({ type: 'active' })
    }
    return
  } catch (e) {
    console.warn(
      '[auth] refreshAuth: fetch threw, falling through to show login panel',
      e,
    )
  }

  notifyAuthExpired()
}
