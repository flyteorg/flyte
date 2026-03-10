'use client'

import { Button } from '@/components/Button'
import { DisabledButtonWithTooltip } from '@/components/pages/SettingsUserManagement/DisabledButtonWithTooltip'
import { SparklesIcon } from '@heroicons/react/24/outline'

export interface ExplainErrorButtonProps {
  org?: string
  domain?: string
  project?: string
  runName?: string
  actionId?: string
}

export const ExplainErrorButton = ({}: ExplainErrorButtonProps) => {
  return (
    <DisabledButtonWithTooltip>
      <Button
        disabled
        outline
        size="xs"
        aria-label="Explain"
        title="Get AI help with this error"
        className="items-center justify-center gap-1.5"
      >
        <SparklesIcon data-slot="icon" aria-hidden="true" className="h-4 w-4" />
        <span className="text-[13px]">Explain</span>
      </Button>
    </DisabledButtonWithTooltip>
  )
}
