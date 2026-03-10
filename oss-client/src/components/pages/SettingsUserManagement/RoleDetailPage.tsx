'use client'

import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { ChevronLeftIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Suspense } from 'react'
import { LicensedEditionPlaceholder } from './LicensedEditionPlaceholder'

export function SettingsRoleDetailPage({ roleName: _roleName }: { roleName: string }) {
  const searchParams = useSearchParams()
  const returnTo = searchParams.get('returnTo')

  const backUrl = returnTo
    ? `/settings/user-management?tab=roles&returnTo=${encodeURIComponent(returnTo)}`
    : '/settings/user-management?tab=roles'

  return (
    <Suspense>
      <NavPanelLayout initialSize="wide" mode="embedded" type="settings">
        <main className="flex h-full w-full flex-col bg-(--system-gray-1)">
          <div className="px-10 pt-10 pb-4">
            <Link
              href={backUrl}
              className="inline-flex items-center gap-1 text-sm font-medium text-(--union) hover:underline"
            >
              <ChevronLeftIcon className="size-4" />
              Back to User Management
            </Link>
          </div>
          <div className="mx-10 flex flex-1 items-start">
            <LicensedEditionPlaceholder title="Roles" />
          </div>
        </main>
      </NavPanelLayout>
    </Suspense>
  )
}

export default SettingsRoleDetailPage
