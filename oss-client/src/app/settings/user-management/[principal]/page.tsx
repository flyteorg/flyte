import { type Metadata } from 'next'

import SettingsUserDetailPage from '@/components/pages/SettingsUserManagement/UserDetailPage'

export const metadata: Metadata = {
  title: 'User Details',
}

export default async function Page({
  params,
}: {
  params: Promise<{ principal: string }>
}) {
  const resolvedParams = await params
  return <SettingsUserDetailPage principal={resolvedParams.principal} />
}
