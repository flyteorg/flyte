import { NextResponse } from 'next/server'
import { getPublicOrigin } from '@/app/utils/getPublicOrigin'

export async function GET(
  request: Request,
  {
    params,
  }: { params: Promise<{ project: string; domain: string; taskName: string }> },
) {
  const resolvedParams = await params
  const url = new URL(request.url)
  const searchParams = url.searchParams.toString()
  const queryString = searchParams ? `?${searchParams}` : ''
  // Construct the redirect URL with the public origin and preserve /v2 basepath
  const publicOrigin = getPublicOrigin(request)
  const newPath = `/v2/domain/${resolvedParams.domain}/project/${resolvedParams.project}/triggers/${resolvedParams.taskName}${queryString}`
  const redirectUrl = new URL(newPath, publicOrigin)
  return NextResponse.redirect(redirectUrl)
}
