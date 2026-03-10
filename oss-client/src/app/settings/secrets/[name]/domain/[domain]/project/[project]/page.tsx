import { type Metadata } from 'next'

import { SettingsSecretDetailsPage } from '@/components/pages/SettingsSecrets/SecretDetailsPage'

export const metadata: Metadata = {
  title: 'Secret Details',
}

export default async function Page({
  params,
}: {
  params: Promise<{ name: string }>
}) {
  const resolvedParams = await params
  return <SettingsSecretDetailsPage secretName={resolvedParams.name} />
}
