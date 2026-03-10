// import React from 'react'
import { type Metadata } from 'next'
import { ListTriggersPage } from '@/components/pages/ListTriggers'

export const metadata: Metadata = {
  title: 'Triggers',
}

export default function Home() {
  return <ListTriggersPage />
}
