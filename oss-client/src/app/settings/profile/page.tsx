import { type Metadata } from 'next'

import SettingsProfilePage from '@/components/pages/SettingsProfile/Main'

export const metadata: Metadata = {
  title: 'Profile',
}

export default function Page() {
  return <SettingsProfilePage />
}
