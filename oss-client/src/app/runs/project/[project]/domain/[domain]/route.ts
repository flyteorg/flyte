import { NextResponse } from 'next/server'
import { getPublicOrigin } from '@/app/utils/getPublicOrigin'

export async function GET(
  request: Request,
  { params }: { params: Promise<{ project: string; domain: string }> },
) {
  try {
    const resolvedParams = await params
    const url = new URL(request.url)
    const searchParams = url.searchParams.toString()
    const queryString = searchParams ? `?${searchParams}` : ''
    // Construct the redirect URL with the public origin and preserve /v2 basepath
    const publicOrigin = getPublicOrigin(request)
    const newPath = `/v2/domain/${resolvedParams.domain}/project/${resolvedParams.project}/runs${queryString}`
    const redirectUrl = new URL(newPath, publicOrigin)
    return NextResponse.redirect(redirectUrl)
  } catch (error) {
    console.error('Redirect error:', error)
    // Fallback: try to construct a basic redirect using request URL origin
    try {
      const resolvedParams = await params
      const url = new URL(request.url)
      const searchParams = url.searchParams.toString()
      const queryString = searchParams ? `?${searchParams}` : ''
      const fallbackPath = `/v2/domain/${resolvedParams.domain}/project/${resolvedParams.project}/runs${queryString}`
      return NextResponse.redirect(new URL(fallbackPath, url.origin))
    } catch (fallbackError) {
      console.error('Fallback redirect error:', fallbackError)
      return new NextResponse('Internal Server Error', { status: 500 })
    }
  }
}
