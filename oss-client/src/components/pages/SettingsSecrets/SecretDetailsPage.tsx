'use client'

import { NavPanelLayout } from '@/components/NavPanel/NavPanelLayout'
import { ChevronLeftIcon } from '@heroicons/react/20/solid'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { Suspense } from 'react'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'

export const SettingsSecretDetailsPage = ({
  secretName: _secretName,
}: {
  secretName: string
}) => {
  const searchParams = useSearchParams()
  const returnTo = searchParams.get('returnTo')

  const backUrl = returnTo
    ? `/settings/secrets?returnTo=${encodeURIComponent(returnTo)}`
    : '/settings/secrets'

  return (
    <Suspense>
      <NavPanelLayout initialSize="wide" mode="embedded" type="settings">
        <main className="flex h-full w-full min-w-0 flex-col bg-(--system-gray-1)">
          <div className="px-10 pt-10 pb-4">
            <Link
              href={backUrl}
              className="inline-flex items-center gap-1 text-sm font-medium text-(--union) hover:underline"
            >
              <ChevronLeftIcon className="size-4" />
              Back to Secrets
            </Link>
          </div>
          <div className="mx-10 flex min-w-0 flex-1 flex-col">
            <LicensedEditionPlaceholder fullWidth title="Secrets" />
          </div>
        </main>
      </NavPanelLayout>
    </Suspense>
  )
}
