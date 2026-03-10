import { type Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Run Details',
}

export default function RunDetailsLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return <>{children}</>
}
