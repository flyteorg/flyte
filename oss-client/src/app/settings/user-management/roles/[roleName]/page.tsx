import { type Metadata } from 'next'

import SettingsRoleDetailPage from '@/components/pages/SettingsUserManagement/RoleDetailPage'

export const metadata: Metadata = {
  title: 'Role Details',
}

export default async function Page({
  params,
}: {
  params: Promise<{ roleName: string }>
}) {
  const resolvedParams = await params
  return <SettingsRoleDetailPage roleName={resolvedParams.roleName} />
}
