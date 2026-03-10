import { NextResponse } from 'next/server'
import { getPublicOrigin } from '@/app/utils/getPublicOrigin'

export async function GET(
  request: Request,
  {
    params,
  }: {
    params: Promise<{ name: string; domain: string; project: string }>
  },
) {
  try {
    const resolvedParams = await params
    const url = new URL(request.url)
    const searchParams = url.searchParams.toString()
    const queryString = searchParams ? `?${searchParams}` : ''
    const publicOrigin = getPublicOrigin(request)
    const newPath = `/v2/settings/secrets/${resolvedParams.name}/domain/${resolvedParams.domain}/project/${resolvedParams.project}${queryString}`
    const redirectUrl = new URL(newPath, publicOrigin)
    return NextResponse.redirect(redirectUrl)
  } catch (error) {
    console.error('Redirect error:', error)
    try {
      const resolvedParams = await params
      const url = new URL(request.url)
      const searchParams = url.searchParams.toString()
      const queryString = searchParams ? `?${searchParams}` : ''
      const fallbackPath = `/v2/settings/secrets/${resolvedParams.name}/domain/${resolvedParams.domain}/project/${resolvedParams.project}${queryString}`
      return NextResponse.redirect(new URL(fallbackPath, url.origin))
    } catch (fallbackError) {
      console.error('Fallback redirect error:', fallbackError)
      return new NextResponse('Internal Server Error', { status: 500 })
    }
  }
}
