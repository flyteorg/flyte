import { type Metadata } from 'next'

import React from 'react'
import ListRunsPage from '@/components/pages/ListRuns/Main'

export const metadata: Metadata = {
  title: 'Runs',
}

export default function Home() {
  return <ListRunsPage />
}
