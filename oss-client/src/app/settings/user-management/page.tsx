import { type Metadata } from 'next'

import SettingsUserManagementPage from '@/components/pages/SettingsUserManagement/Main'

export const metadata: Metadata = {
  title: 'User Management',
}

export default function Page() {
  return <SettingsUserManagementPage />
}
