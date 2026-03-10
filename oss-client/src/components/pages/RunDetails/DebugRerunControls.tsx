import { Button } from '@/components/Button'
import ComboButton, { type ComboButtonProps } from '@/components/ComboButton'
import { LaunchFormDrawer } from '@/components/LaunchForm'
import { useActionLaunchFormData } from '@/components/LaunchForm/hooks/useActionLaunchFormData'
import { useSelectedItem } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import {
  LaunchFormStateProvider,
  useLaunchFormState,
} from '@/hooks/useLaunchFormState'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { isTaskSpec } from '@/lib/actionSpecUtils'
import { isActionTerminal } from '@/lib/actionUtils'
import React, { useEffect, useMemo, useState } from 'react'
import { AbortActionModal } from './AbortActionModal'

type ComboOption = ComboButtonProps['options'][number]

const createOption = (
  label: string,
  onClick: () => void,
  options?: {
    size?: 'sm' | 'md'
    color?: 'default' | 'rose' | 'black'
    icon?: React.ReactNode
  },
): ComboOption => {
  const sizeClass = options?.size === 'sm' ? 'text-xs/5' : 'text-[13px]'
  const colorClass =
    options?.color === 'rose'
      ? 'text-rose-600 dark:text-rose-400'
      : options?.color === 'black'
        ? 'text-(--system-black)'
        : 'text-(--system-gray-5)'

  return {
    name: (
      <div
        className={`flex flex-row items-center justify-between gap-1.5 ${sizeClass} ${colorClass}`}
      >
        {options?.icon || label}
      </div>
    ),
    onClick,
  }
}

export const SummaryPopoverWrapper: React.FC<{
  actionDetailsQuery: ReturnType<typeof useWatchActionDetails>
}> = ({ actionDetailsQuery }) => {
  const { selectedItem } = useSelectedItem()
  const { drawerMeta, formMethods, isDataFetched } = useActionLaunchFormData({
    id: selectedItem?.id,
    actionDetailsQuery,
  })

  const { setIsOpen, setTaskSpec } = useLaunchFormState()
  const [isAbortModalOpen, setIsAbortModalOpen] = useState(false)

  const isNonTerminal = actionDetailsQuery.data
    ? !isActionTerminal(actionDetailsQuery.data)
    : false

  useEffect(() => {
    if (isTaskSpec(actionDetailsQuery.data?.spec)) {
      setTaskSpec(actionDetailsQuery.data.spec.value)
    }
  }, [actionDetailsQuery.data?.spec, setTaskSpec])

  const options = useMemo(() => {
    if (!isNonTerminal) return []

    return [
      createOption('Abort action', () => setIsAbortModalOpen(true), {
        size: 'sm',
        color: 'rose',
      }),
      createOption('Rerun action', () => setIsOpen(true), {
        size: 'sm',
      }),
    ]
  }, [isNonTerminal, setIsOpen, setIsAbortModalOpen])

  const shouldShowComboButton = isNonTerminal && options.length > 0
  const buttonContainerClassName =
    'rounded-xl border-1 border-zinc-900/30 dark:border-(--system-gray-4)'
  const comboButtonWrapperClassName =
    '[&>div>div>button]:flex [&>div>div>button]:items-center [&>div>div>button]:!px-3 [&>div>div>button]:!py-1 [&>div>div>button]:!leading-none'
  const singleButtonClassName =
    '!flex !items-center rounded-xl !border-transparent !px-3 !py-1 !leading-none text-zinc-600 dark:text-zinc-400'

  return (
    <>
      {shouldShowComboButton ? (
        <div className={buttonContainerClassName}>
          <div className={comboButtonWrapperClassName}>
            <ComboButton
              size="xs"
              color="dark/zinc"
              options={options}
              outline
            />
          </div>
        </div>
      ) : (
        <div className={buttonContainerClassName}>
          <Button
            size="md"
            color="dark/zinc"
            outline
            onClick={() => setIsOpen(true)}
            className={singleButtonClassName}
          >
            <div className="flex flex-row items-center justify-between gap-1.5 text-xs/5 text-zinc-600 dark:text-zinc-400">
              Rerun action
            </div>
          </Button>
        </div>
      )}

      <LaunchFormDrawer
        drawerMeta={drawerMeta}
        formMethods={formMethods}
        isDataFetched={isDataFetched}
      />

      {isNonTerminal && (
        <AbortActionModal
          actionDetails={actionDetailsQuery.data}
          isOpen={isAbortModalOpen}
          setIsOpen={setIsAbortModalOpen}
        />
      )}
    </>
  )
}

export const DebugRerunControls: React.FC<{
  actionDetailsQuery: ReturnType<typeof useWatchActionDetails>
}> = ({ actionDetailsQuery }) => {
  return (
    <LaunchFormStateProvider buttonText="Rerun">
      <SummaryPopoverWrapper actionDetailsQuery={actionDetailsQuery} />
    </LaunchFormStateProvider>
  )
}
