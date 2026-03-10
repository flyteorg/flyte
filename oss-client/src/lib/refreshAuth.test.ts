import { describe, it, expect, vi, beforeEach } from 'vitest'
import { Code, ConnectError } from '@connectrpc/connect'
import { refreshAuth } from './refreshAuth'
import { getLoginUrl } from './apiUtils'
import { notifyAuthExpired } from './authExpiredNotifier'

vi.mock('./apiUtils', () => ({
  getLoginUrl: vi.fn(() => 'https://example.com/login'),
}))

vi.mock('./authExpiredNotifier', () => ({
  notifyAuthExpired: vi.fn(),
}))

vi.mocked(getLoginUrl)
const mockedNotifyAuthExpired = vi.mocked(notifyAuthExpired)

describe('refreshAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.stubGlobal('fetch', vi.fn())
  })

  it('does nothing for non-auth errors', async () => {
    await refreshAuth(new Error('random error'))
    expect(fetch).not.toHaveBeenCalled()
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('does nothing for null/undefined errors', async () => {
    await refreshAuth(undefined)
    expect(fetch).not.toHaveBeenCalled()
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('attempts token refresh for error with rawMessage "HTTP 401"', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('ok'))

    await refreshAuth({ rawMessage: 'HTTP 401' })

    expect(fetch).toHaveBeenCalledWith('https://example.com/login', {
      method: 'GET',
      headers: { Accept: 'text/html' },
      credentials: 'include',
      redirect: 'follow',
      mode: 'no-cors',
    })
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('refetches active queries after successful token refresh', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('ok'))

    const refetchQueries = vi.fn().mockResolvedValueOnce(undefined)
    // eslint-disable-next-line
    const queryClient = { refetchQueries } as any

    await refreshAuth({ rawMessage: 'HTTP 401' }, { queryClient })

    expect(refetchQueries).toHaveBeenCalledWith({ type: 'active' })
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('notifies auth expired when token refresh fails', async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new Error('network error'))

    await refreshAuth({ rawMessage: 'HTTP 401' })

    expect(fetch).toHaveBeenCalled()
    expect(mockedNotifyAuthExpired).toHaveBeenCalled()
  })

  it('attempts token refresh for ConnectError Unauthenticated', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('ok'))
    const error = new ConnectError('unauth', Code.Unauthenticated)

    await refreshAuth(error)

    expect(fetch).toHaveBeenCalledWith('https://example.com/login', {
      method: 'GET',
      headers: { Accept: 'text/html' },
      credentials: 'include',
      redirect: 'follow',
      mode: 'no-cors',
    })
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('attempts token refresh for object with code 16', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('ok'))

    await refreshAuth({ code: 16 })

    expect(fetch).toHaveBeenCalled()
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('does not refetch if queryClient is not provided', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('ok'))

    await refreshAuth({ rawMessage: 'HTTP 401' })

    expect(fetch).toHaveBeenCalled()
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })

  it('does not refetch if queryClient has no refetchQueries', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(new Response('ok'))
    // eslint-disable-next-line
    await refreshAuth({ rawMessage: 'HTTP 401' }, { queryClient: {} as any })

    expect(fetch).toHaveBeenCalled()
    expect(mockedNotifyAuthExpired).not.toHaveBeenCalled()
  })
})
