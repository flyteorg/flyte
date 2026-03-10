import { Code, ConnectError } from '@connectrpc/connect'

/**
 * Checks if an error is a 403 Forbidden or CORS error.
 * These errors typically indicate storage bucket configuration issues
 * that won't be resolved by retrying.
 *
 * @param error - The error to check
 * @param hasPresignedUrl - Optional: if true and error is a network error, treat as CORS
 * @returns true if the error is a 403 or CORS error
 */
export function is403OrCorsError(
  error: unknown,
  hasPresignedUrl?: boolean,
): boolean {
  if (!error) return false

  // Check for ConnectError with PermissionDenied code
  if (error instanceof ConnectError) {
    return error.code === Code.PermissionDenied
  }

  // Check for TypeError (often indicates CORS or network errors)
  // CORS errors typically show up as TypeError with "Failed to fetch"
  if (error instanceof TypeError) {
    const errorMessage = error.message.toLowerCase()
    return (
      errorMessage.includes('failed to fetch') ||
      errorMessage.includes('networkerror') ||
      errorMessage.includes('load failed') ||
      errorMessage.includes('network request failed') ||
      errorMessage.includes('cors') ||
      errorMessage.includes('access-control-allow-origin') ||
      errorMessage.includes('cross-origin')
    )
  }

  // Check for Error with 403, forbidden, or CORS in message
  // Also check error name in case it was a TypeError that got re-thrown
  if (error instanceof Error) {
    const errorMessage = error.message.toLowerCase()
    const errorName = error.name?.toLowerCase() || ''

    // Check if it's a network/CORS related error
    const isNetworkError =
      errorName === 'typeerror' ||
      errorMessage.includes('failed to fetch') ||
      errorMessage.includes('networkerror') ||
      errorMessage.includes('load failed') ||
      errorMessage.includes('network request failed') ||
      errorMessage.includes('cors') ||
      errorMessage.includes('access-control-allow-origin') ||
      errorMessage.includes('cross-origin')

    // If it's a network error and we have a presignedUrl,
    // it's likely a CORS issue
    if (isNetworkError && hasPresignedUrl) {
      return true
    }

    // Check for 403/forbidden errors
    return (
      errorMessage.includes('403') ||
      errorMessage.includes('forbidden') ||
      errorMessage.includes('permission denied')
    )
  }

  return false
}

/**
 * Checks if an error is a 404 Not Found error.
 * These errors indicate the resource doesn't exist and won't be resolved by retrying.
 *
 * @param error - The error to check
 * @returns true if the error is a 404 error
 */
export function is404Error(error: unknown): boolean {
  if (!error) return false

  // Check for ConnectError with NotFound code
  if (error instanceof ConnectError) {
    return error.code === Code.NotFound
  }

  // Check if error object has a code property set to "not_found"
  if (error && typeof error === 'object' && 'code' in error) {
    const errorCode = (error as { code?: unknown }).code
    if (
      errorCode === 'not_found' ||
      errorCode === 'NotFound' ||
      errorCode === Code.NotFound
    ) {
      return true
    }
  }

  // Check for Error with 404 in message
  if (error instanceof Error) {
    const errorMessage = error.message.toLowerCase()
    return (
      errorMessage.includes('404') ||
      errorMessage.includes('not found') ||
      errorMessage.includes('not_found') ||
      errorMessage.includes('failed to head object before signing') ||
      errorMessage.includes('failed to create signed url')
    )
  }

  return false
}

/**
 * Checks if an error is a FailedPrecondition error.
 * Used when the backend indicates a prerequisite is missing (e.g., no code bundle URL).
 *
 * @param error - The error to check
 * @returns true if the error is a FailedPrecondition error
 */
export function isFailedPreconditionError(error: unknown): boolean {
  if (!error) return false

  // Check for ConnectError with FailedPrecondition code
  if (error instanceof ConnectError) {
    return error.code === Code.FailedPrecondition
  }

  // Check if error object has a code property set to "failed_precondition"
  if (error && typeof error === 'object' && 'code' in error) {
    const errorCode = (error as { code?: unknown }).code
    if (
      errorCode === 'failed_precondition' ||
      errorCode === 'FailedPrecondition' ||
      errorCode === Code.FailedPrecondition
    ) {
      return true
    }
  }

  // Check for Error with failed_precondition or related text in message
  if (error instanceof Error) {
    const errorMessage = error.message.toLowerCase()
    return (
      errorMessage.includes('failed_precondition') ||
      errorMessage.includes('failed precondition') ||
      errorMessage.includes('no code bundle url found in the selected attempt')
    )
  }

  return false
}

const AUTH_ERROR_MESSAGE_PATTERNS = [
  /\bHTTP\s+401\b/i,
  /\bstatus\s*(code)?\s*:?\s*401\b/i,
  /\b401\s+(Unauthorized|unauthenticated)\b/i,
  /\bUnauthenticated\b/,
  /\bunauthenticated\b/,
]

function isAuthErrorMessage(message: string): boolean {
  return AUTH_ERROR_MESSAGE_PATTERNS.some((re) => re.test(message))
}

/**
 * Checks if an error is a 401 Unauthenticated / auth expired error.
 * Used to show the login panel when credentials have expired.
 * Prefers structured signals (ConnectError code, rawMessage) and only
 * falls back to message patterns that clearly indicate HTTP 401 / auth.
 *
 * @param error - The error to check
 * @returns true if the error is an auth/401 error
 */
export function isAuthError(error: unknown): boolean {
  if (!error) return false

  if (error instanceof ConnectError) {
    return error.code === Code.Unauthenticated
  }

  if (error && typeof error === 'object' && 'code' in error) {
    const code = (error as { code?: unknown }).code
    if (code === Code.Unauthenticated || code === 16) {
      return true
    }
  }

  if (error && typeof error === 'object' && 'rawMessage' in error) {
    return (error as { rawMessage?: string }).rawMessage === 'HTTP 401'
  }

  if (error instanceof Error) {
    return isAuthErrorMessage(error.message)
  }

  return false
}

/**
 * Strong auth signal only: structured codes or rawMessage. No message string
 * matching. Use when deciding to show the login panel from refreshAuth so we
 * don't trigger on weak or ambiguous signals.
 */
export function isStrongAuthError(error: unknown): boolean {
  if (!error) return false

  if (error instanceof ConnectError) {
    return error.code === Code.Unauthenticated
  }

  if (error && typeof error === 'object' && 'code' in error) {
    const code = (error as { code?: unknown }).code
    if (code === Code.Unauthenticated || code === 16) {
      return true
    }
  }

  if (error && typeof error === 'object' && 'rawMessage' in error) {
    return (error as { rawMessage?: string }).rawMessage === 'HTTP 401'
  }

  return false
}

/**
 * Checks if an error indicates the request was aborted/cancelled.
 * These errors should not be retried.
 *
 * @param error - The error to check
 * @returns true if the error indicates the request was aborted
 */
export function isAbortedError(error: unknown): boolean {
  if (!error) return false

  // Check for ConnectError with cancellation codes
  if (error instanceof ConnectError) {
    return error.code === Code.Canceled || error.code === Code.Aborted
  }

  // Check for Error with cancellation indicators in message or name
  if (error instanceof Error) {
    const errorMessage = error.message.toLowerCase()
    return (
      errorMessage.includes('aborted') ||
      errorMessage.includes('cancelled') ||
      errorMessage.includes('canceled') ||
      error.name === 'AbortError'
    )
  }

  return false
}

/**
 * Checks if an error is non-retryable (should not be retried).
 * These are typically client errors or permanent failures.
 *
 * @param error - The error to check
 * @returns true if the error should not be retried
 */
export function isNonRetryableError(error: unknown): boolean {
  if (!error) return false

  // Aborted errors should not be retried
  if (isAbortedError(error)) {
    return true
  }

  // Check for ConnectError with non-retryable codes
  if (error instanceof ConnectError) {
    const nonRetryableCodes = [
      Code.Canceled, // Request was cancelled
      Code.Aborted, // Operation was aborted
      Code.NotFound,
      Code.PermissionDenied,
      Code.InvalidArgument,
      Code.Unauthenticated,
      Code.FailedPrecondition,
    ]
    return nonRetryableCodes.includes(error.code)
  }

  // 404 and 403 errors should not be retried
  if (is404Error(error) || is403OrCorsError(error)) {
    return true
  }

  return false
}

/**
 * Checks if an error is a network error that might be retryable.
 * These are typically transient network issues.
 *
 * @param error - The error to check
 * @returns true if the error is a network error
 */
export function isNetworkError(error: unknown): boolean {
  if (!error) return false

  if (error instanceof ConnectError) {
    const networkErrorCodes = [
      Code.Unavailable,
      Code.DeadlineExceeded,
      Code.ResourceExhausted,
      Code.Unknown,
    ]
    return networkErrorCodes.includes(error.code)
  }

  if (error instanceof TypeError) {
    const message = error.message.toLowerCase()
    return (
      message.includes('failed to fetch') ||
      message.includes('networkerror') ||
      message.includes('load failed') ||
      message.includes('network request failed')
    )
  }

  return false
}

/**
 * Extracts error information for logging purposes.
 *
 * @param error - The error to extract information from
 * @returns An object with error details
 */
export function getErrorInfo(error: unknown): {
  message: string
  name: string
  code?: Code
  type: string
} {
  if (error instanceof ConnectError) {
    return {
      message: error.message,
      name: error.name,
      code: error.code,
      type: 'ConnectError',
    }
  }

  if (error instanceof Error) {
    return {
      message: error.message,
      name: error.name,
      type: error.constructor.name,
    }
  }

  return {
    message: String(error),
    name: typeof error,
    type: typeof error,
  }
}
