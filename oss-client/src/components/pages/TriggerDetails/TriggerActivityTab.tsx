import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import React from 'react'

export const TriggerActivityTab: React.FC = () => {
  return (
    <div className="flex w-full min-w-0 flex-1 flex-col gap-2 px-8 pb-8">
      <div className="flex min-h-0 flex-1 flex-col items-center justify-center">
        <LicensedEditionPlaceholder title="Activity" fullWidth hideBorder />
      </div>
    </div>
  )
}
