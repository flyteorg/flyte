/**
 * Gets the public origin from the request, handling proxy headers in production.
 * In production behind a proxy, the request URL may have internal hostnames,
 * so we check X-Forwarded-Host and Host headers first.
 */
export function getPublicOrigin(request: Request): string {
  try {
    const url = new URL(request.url)
    // In production behind a proxy, use X-Forwarded-Host or Host header
    const forwardedHost = request.headers.get('x-forwarded-host')
    const host = request.headers.get('host')
    const forwardedProto = request.headers.get('x-forwarded-proto')

    // Determine protocol
    let protocol = 'https:'
    if (forwardedProto) {
      protocol = forwardedProto.endsWith(':')
        ? forwardedProto
        : `${forwardedProto}:`
    } else if (url.protocol) {
      protocol = url.protocol
    }

    if (forwardedHost) {
      return `${protocol}//${forwardedHost}`
    }
    if (host) {
      return `${protocol}//${host}`
    }
    return url.origin
  } catch (error) {
    // Fallback to request URL origin if anything fails
    try {
      return new URL(request.url).origin
    } catch {
      // Last resort: return a basic origin structure
      // This should rarely happen, but prevents complete failure
      const host =
        request.headers.get('host') || request.headers.get('x-forwarded-host')
      if (host) {
        return `https://${host}`
      }
      throw error // Re-throw if we can't construct anything
    }
  }
}
