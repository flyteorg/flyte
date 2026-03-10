import { NextResponse } from 'next/server'
import { getPublicOrigin } from '@/app/utils/getPublicOrigin'

export async function GET(
  request: Request,
  {
    params,
  }: {
    params: Promise<{ name: string; domain: string }>
  },
) {
  try {
    const resolvedParams = await params
    const url = new URL(request.url)
    const searchParams = url.searchParams.toString()
    const queryString = searchParams ? `?${searchParams}` : ''
    const publicOrigin = getPublicOrigin(request)
    // Redirect to the new structure, but we need project - check if it's in query params
    const project = url.searchParams.get('project')
    if (project) {
      const newPath = `/v2/settings/secrets/${resolvedParams.name}/domain/${resolvedParams.domain}/project/${project}${queryString}`
      const redirectUrl = new URL(newPath, publicOrigin)
      return NextResponse.redirect(redirectUrl)
    }
    // If no project, redirect to secrets list (or we could keep the old path structure)
    // For now, redirect to secrets list
    const newPath = `/v2/settings/secrets${queryString}`
    const redirectUrl = new URL(newPath, publicOrigin)
    return NextResponse.redirect(redirectUrl)
  } catch (error) {
    console.error('Redirect error:', error)
    try {
      const resolvedParams = await params
      const url = new URL(request.url)
      const searchParams = url.searchParams.toString()
      const queryString = searchParams ? `?${searchParams}` : ''
      const project = url.searchParams.get('project')
      if (project) {
        const fallbackPath = `/v2/settings/secrets/${resolvedParams.name}/domain/${resolvedParams.domain}/project/${project}${queryString}`
        return NextResponse.redirect(new URL(fallbackPath, url.origin))
      }
      const fallbackPath = `/v2/settings/secrets${queryString}`
      return NextResponse.redirect(new URL(fallbackPath, url.origin))
    } catch (fallbackError) {
      console.error('Fallback redirect error:', fallbackError)
      return new NextResponse('Internal Server Error', { status: 500 })
    }
  }
}
