'use client'

import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { Suspense } from 'react'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'

export function SettingsProfilePage() {
  return (
    <Suspense>
      <NavPanelLayout initialSize="wide" mode="embedded" type="settings">
        <main className="flex h-full w-full min-w-0 flex-col bg-(--system-gray-1)">
          <div className="px-10 pt-10 pb-5">
            <h1 className="text-xl font-medium">Profile</h1>
          </div>
          <div className="mx-10 flex min-w-0 flex-1 flex-col">
            <LicensedEditionPlaceholder fullWidth title="Profile" />
          </div>
        </main>
      </NavPanelLayout>
    </Suspense>
  )
}

export default SettingsProfilePage
