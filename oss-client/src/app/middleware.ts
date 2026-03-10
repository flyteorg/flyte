import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * App http middleware for template rendering rewrites
 * @param request
 * @returns
 */
export function middleware(request: NextRequest) {
  if (!request.url.includes('_fe/')) {
    return NextResponse.rewrite(new URL('/v2/', request.url))
  }

  // return NextResponse.next();
}

// See "Matching Paths" below to learn more
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - healthz (Health check route)
     * - _fe (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|healthz|_fe|_next/static|_next/image).*)',
  ],
}
