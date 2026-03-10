import { type Metadata } from 'next'

import { ListAppsPage } from '@/components/pages/ListApps/Main'

export const metadata: Metadata = {
  title: 'Apps',
}

export default function Home() {
  return <ListAppsPage />
}
