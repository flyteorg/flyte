'use client'

import dynamic from 'next/dynamic'

/**
 * https://nextjs.org/docs/pages/building-your-application/routing/dynamic-routes#optional-catch-all-segments
 */

const RunDetailsPageClient = dynamic(
  () =>
    import('@/components/pages/RunDetails/Main').then((mod) => ({
      default: mod.RunDetailsPage,
    })),
  { ssr: false },
)

export default function Home() {
  return <RunDetailsPageClient />
}
