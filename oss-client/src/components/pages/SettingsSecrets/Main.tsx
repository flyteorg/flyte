'use client'

import { Button } from '@/components/Button'
import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { Suspense } from 'react'
import { DisabledButtonWithTooltip } from '@/components/pages/SettingsUserManagement/DisabledButtonWithTooltip'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'

export function SettingsSecretsPage() {
  return (
    <Suspense>
      <NavPanelLayout initialSize="wide" mode="embedded" type="settings">
        <main className="flex h-full w-full min-w-0 flex-col bg-(--system-gray-1)">
          <div className="flex min-h-0 items-center justify-between gap-2 px-10 pt-10 pb-5">
            <h1 className="text-xl font-medium">Secrets</h1>
            <DisabledButtonWithTooltip>
              <Button disabled color="union" size="sm">
                Create New Secret
              </Button>
            </DisabledButtonWithTooltip>
          </div>
          <div className="mx-10 flex min-w-0 flex-1 flex-col">
            <LicensedEditionPlaceholder fullWidth title="Secrets" />
          </div>
        </main>
      </NavPanelLayout>
    </Suspense>
  )
}

export default SettingsSecretsPage
