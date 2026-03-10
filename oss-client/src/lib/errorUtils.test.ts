import { describe, it, expect } from 'vitest'
import { ConnectError, Code } from '@connectrpc/connect'
import {
  is403OrCorsError,
  is404Error,
  isAuthError,
  isStrongAuthError,
  isAbortedError,
  isNonRetryableError,
  isNetworkError,
  getErrorInfo,
} from './errorUtils'

describe('errorUtils', () => {
  describe('is403OrCorsError', () => {
    it('returns false for falsy values', () => {
      expect(is403OrCorsError(null)).toBe(false)
      expect(is403OrCorsError(undefined)).toBe(false)
    })

    it('returns true for ConnectError with PermissionDenied', () => {
      expect(
        is403OrCorsError(new ConnectError('denied', Code.PermissionDenied)),
      ).toBe(true)
    })

    it('returns false for ConnectError with other codes', () => {
      expect(
        is403OrCorsError(new ConnectError('not found', Code.NotFound)),
      ).toBe(false)
    })

    it('returns true for TypeError with "Failed to fetch"', () => {
      expect(is403OrCorsError(new TypeError('Failed to fetch'))).toBe(true)
    })

    it('returns true for TypeError with CORS-related messages', () => {
      expect(is403OrCorsError(new TypeError('CORS error'))).toBe(true)
      expect(
        is403OrCorsError(new TypeError('access-control-allow-origin')),
      ).toBe(true)
      expect(is403OrCorsError(new TypeError('cross-origin blocked'))).toBe(true)
      expect(is403OrCorsError(new TypeError('NetworkError'))).toBe(true)
      expect(is403OrCorsError(new TypeError('Load failed'))).toBe(true)
      expect(is403OrCorsError(new TypeError('network request failed'))).toBe(
        true,
      )
    })

    it('returns true for Error with 403 in message', () => {
      expect(is403OrCorsError(new Error('HTTP 403'))).toBe(true)
      expect(is403OrCorsError(new Error('forbidden'))).toBe(true)
      expect(is403OrCorsError(new Error('permission denied'))).toBe(true)
    })

    it('returns true for network error with presignedUrl flag', () => {
      const err = new Error('Failed to fetch')
      // Manually set name to simulate a re-thrown TypeError
      err.name = 'TypeError'
      expect(is403OrCorsError(err, true)).toBe(true)
    })

    it('returns false for generic Error without matching message', () => {
      expect(is403OrCorsError(new Error('something else'))).toBe(false)
    })
  })

  describe('is404Error', () => {
    it('returns false for falsy values', () => {
      expect(is404Error(null)).toBe(false)
      expect(is404Error(undefined)).toBe(false)
    })

    it('returns true for ConnectError with NotFound', () => {
      expect(is404Error(new ConnectError('missing', Code.NotFound))).toBe(true)
    })

    it('returns false for ConnectError with other codes', () => {
      expect(
        is404Error(new ConnectError('denied', Code.PermissionDenied)),
      ).toBe(false)
    })

    it('returns true for object with code "not_found"', () => {
      expect(is404Error({ code: 'not_found' })).toBe(true)
      expect(is404Error({ code: 'NotFound' })).toBe(true)
      expect(is404Error({ code: Code.NotFound })).toBe(true)
    })

    it('returns true for Error with 404-related messages', () => {
      expect(is404Error(new Error('HTTP 404'))).toBe(true)
      expect(is404Error(new Error('resource not found'))).toBe(true)
      expect(is404Error(new Error('not_found'))).toBe(true)
      expect(
        is404Error(new Error('failed to head object before signing')),
      ).toBe(true)
      expect(is404Error(new Error('failed to create signed url'))).toBe(true)
    })

    it('returns false for generic Error', () => {
      expect(is404Error(new Error('something else'))).toBe(false)
    })
  })

  describe('isAuthError', () => {
    it('returns false for falsy values', () => {
      expect(isAuthError(null)).toBe(false)
      expect(isAuthError(undefined)).toBe(false)
    })

    it('returns true for ConnectError with Unauthenticated', () => {
      expect(
        isAuthError(new ConnectError('unauth', Code.Unauthenticated)),
      ).toBe(true)
    })

    it('returns true for object with Unauthenticated code', () => {
      expect(isAuthError({ code: Code.Unauthenticated })).toBe(true)
      expect(isAuthError({ code: 16 })).toBe(true)
    })

    it('returns true for object with rawMessage "HTTP 401"', () => {
      expect(isAuthError({ rawMessage: 'HTTP 401' })).toBe(true)
    })

    it('returns false for object with different rawMessage', () => {
      expect(isAuthError({ rawMessage: 'HTTP 403' })).toBe(false)
    })

    it('returns true for Error with auth-related messages', () => {
      expect(isAuthError(new Error('HTTP 401'))).toBe(true)
      expect(isAuthError(new Error('status code: 401'))).toBe(true)
      expect(isAuthError(new Error('401 Unauthorized'))).toBe(true)
      expect(isAuthError(new Error('Unauthenticated'))).toBe(true)
      expect(isAuthError(new Error('unauthenticated'))).toBe(true)
    })

    it('returns false for non-auth errors', () => {
      expect(isAuthError(new Error('HTTP 403'))).toBe(false)
      expect(isAuthError(new Error('something else'))).toBe(false)
    })
  })

  describe('isStrongAuthError', () => {
    it('returns false for falsy values', () => {
      expect(isStrongAuthError(null)).toBe(false)
      expect(isStrongAuthError(undefined)).toBe(false)
    })

    it('returns true for ConnectError with Unauthenticated', () => {
      expect(
        isStrongAuthError(new ConnectError('unauth', Code.Unauthenticated)),
      ).toBe(true)
    })

    it('returns true for object with Unauthenticated code', () => {
      expect(isStrongAuthError({ code: Code.Unauthenticated })).toBe(true)
      expect(isStrongAuthError({ code: 16 })).toBe(true)
    })

    it('returns true for object with rawMessage "HTTP 401"', () => {
      expect(isStrongAuthError({ rawMessage: 'HTTP 401' })).toBe(true)
    })

    it('does NOT match Error message patterns (strong signals only)', () => {
      expect(isStrongAuthError(new Error('HTTP 401'))).toBe(false)
      expect(isStrongAuthError(new Error('Unauthenticated'))).toBe(false)
    })

    it('returns false for ConnectError with other codes', () => {
      expect(
        isStrongAuthError(new ConnectError('denied', Code.PermissionDenied)),
      ).toBe(false)
    })
  })

  describe('isAbortedError', () => {
    it('returns false for falsy values', () => {
      expect(isAbortedError(null)).toBe(false)
      expect(isAbortedError(undefined)).toBe(false)
    })

    it('returns true for ConnectError with Canceled', () => {
      expect(isAbortedError(new ConnectError('canceled', Code.Canceled))).toBe(
        true,
      )
    })

    it('returns true for ConnectError with Aborted', () => {
      expect(isAbortedError(new ConnectError('aborted', Code.Aborted))).toBe(
        true,
      )
    })

    it('returns true for Error with abort-related messages', () => {
      expect(isAbortedError(new Error('request aborted'))).toBe(true)
      expect(isAbortedError(new Error('request cancelled'))).toBe(true)
      expect(isAbortedError(new Error('request canceled'))).toBe(true)
    })

    it('returns true for AbortError', () => {
      const err = new Error('abort')
      err.name = 'AbortError'
      expect(isAbortedError(err)).toBe(true)
    })

    it('returns false for generic Error', () => {
      expect(isAbortedError(new Error('something else'))).toBe(false)
    })
  })

  describe('isNonRetryableError', () => {
    it('returns false for falsy values', () => {
      expect(isNonRetryableError(null)).toBe(false)
    })

    it('returns true for aborted errors', () => {
      expect(
        isNonRetryableError(new ConnectError('canceled', Code.Canceled)),
      ).toBe(true)
    })

    it('returns true for ConnectError with non-retryable codes', () => {
      const nonRetryableCodes = [
        Code.Canceled,
        Code.Aborted,
        Code.NotFound,
        Code.PermissionDenied,
        Code.InvalidArgument,
        Code.Unauthenticated,
        Code.FailedPrecondition,
      ]
      for (const code of nonRetryableCodes) {
        expect(isNonRetryableError(new ConnectError('err', code))).toBe(true)
      }
    })

    it('returns false for ConnectError with retryable codes', () => {
      expect(
        isNonRetryableError(new ConnectError('unavail', Code.Unavailable)),
      ).toBe(false)
      expect(
        isNonRetryableError(new ConnectError('timeout', Code.DeadlineExceeded)),
      ).toBe(false)
    })

    it('returns true for 404 errors', () => {
      expect(isNonRetryableError(new Error('HTTP 404'))).toBe(true)
    })

    it('returns true for 403 errors', () => {
      expect(isNonRetryableError(new Error('HTTP 403'))).toBe(true)
    })

    it('returns false for generic Error', () => {
      expect(isNonRetryableError(new Error('timeout'))).toBe(false)
    })
  })

  describe('isNetworkError', () => {
    it('returns false for falsy values', () => {
      expect(isNetworkError(null)).toBe(false)
    })

    it('returns true for ConnectError with network-related codes', () => {
      expect(
        isNetworkError(new ConnectError('unavail', Code.Unavailable)),
      ).toBe(true)
      expect(
        isNetworkError(new ConnectError('timeout', Code.DeadlineExceeded)),
      ).toBe(true)
      expect(
        isNetworkError(new ConnectError('exhausted', Code.ResourceExhausted)),
      ).toBe(true)
      expect(isNetworkError(new ConnectError('unknown', Code.Unknown))).toBe(
        true,
      )
    })

    it('returns false for ConnectError with non-network codes', () => {
      expect(
        isNetworkError(new ConnectError('unauth', Code.Unauthenticated)),
      ).toBe(false)
    })

    it('returns true for TypeError with network messages', () => {
      expect(isNetworkError(new TypeError('Failed to fetch'))).toBe(true)
      expect(isNetworkError(new TypeError('NetworkError'))).toBe(true)
      expect(isNetworkError(new TypeError('Load failed'))).toBe(true)
      expect(isNetworkError(new TypeError('network request failed'))).toBe(true)
    })

    it('returns false for TypeError with non-network messages', () => {
      expect(isNetworkError(new TypeError('Cannot read properties'))).toBe(
        false,
      )
    })

    it('returns false for generic Error', () => {
      expect(isNetworkError(new Error('Failed to fetch'))).toBe(false)
    })
  })

  describe('getErrorInfo', () => {
    it('extracts info from ConnectError', () => {
      const err = new ConnectError('test error', Code.NotFound)
      const info = getErrorInfo(err)
      expect(info.type).toBe('ConnectError')
      expect(info.code).toBe(Code.NotFound)
      expect(info.name).toBe('ConnectError')
      expect(info.message).toContain('test error')
    })

    it('extracts info from regular Error', () => {
      const err = new Error('regular error')
      const info = getErrorInfo(err)
      expect(info.type).toBe('Error')
      expect(info.message).toBe('regular error')
      expect(info.name).toBe('Error')
      expect(info.code).toBeUndefined()
    })

    it('extracts info from TypeError', () => {
      const err = new TypeError('type error')
      const info = getErrorInfo(err)
      expect(info.type).toBe('TypeError')
      expect(info.name).toBe('TypeError')
    })

    it('handles non-Error values', () => {
      const info = getErrorInfo('string error')
      expect(info.message).toBe('string error')
      expect(info.type).toBe('string')
    })
  })
})
