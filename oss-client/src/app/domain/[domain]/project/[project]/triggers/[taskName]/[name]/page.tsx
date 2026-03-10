import { type Metadata } from 'next'

import { TriggerDetailsPage } from '@/components/pages/TriggerDetails/Main'

export const metadata: Metadata = {
  title: 'Trigger Details',
}

export default function Home() {
  return <TriggerDetailsPage />
}
