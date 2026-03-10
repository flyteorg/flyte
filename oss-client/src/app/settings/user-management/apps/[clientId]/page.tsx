import { type Metadata } from 'next'

import SettingsOAuthAppDetailPage from '@/components/pages/SettingsUserManagement/OAuthAppDetailPage'

export const metadata: Metadata = {
  title: 'OAuth App Details',
}

export default async function Page({
  params,
}: {
  params: Promise<{ clientId: string }>
}) {
  const resolvedParams = await params
  return <SettingsOAuthAppDetailPage clientId={resolvedParams.clientId} />
}
