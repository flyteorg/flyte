import { type Metadata } from 'next'

import SettingsPolicyDetailPage from '@/components/pages/SettingsUserManagement/PolicyDetailPage'

export const metadata: Metadata = {
  title: 'Policy Details',
}

export default async function Page({
  params,
}: {
  params: Promise<{ policyName: string }>
}) {
  const resolvedParams = await params
  return <SettingsPolicyDetailPage policyName={resolvedParams.policyName} />
}
