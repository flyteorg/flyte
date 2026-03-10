import { type Metadata } from 'next'

import SettingsSecretsPage from '@/components/pages/SettingsSecrets/Main'

export const metadata: Metadata = {
  title: 'Secrets',
}

export default function Page() {
  return <SettingsSecretsPage />
}
