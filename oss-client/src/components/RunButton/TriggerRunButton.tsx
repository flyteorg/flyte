'use client'

import { Button } from '@/components/Button'
import { LaunchFormDrawer } from '@/components/LaunchForm'
import { TriggerName } from '@/gen/flyteidl2/common/identifier_pb'
import { PlayIcon } from '@heroicons/react/24/solid'
import { useEffect, useRef } from 'react'
import { useTriggerRun } from '@/components/RunButton/useTriggerRun'

type TriggerRunButtonProps = {
  triggerName: TriggerName | undefined
}

export const TriggerRunButton = ({ triggerName }: TriggerRunButtonProps) => {
  const {
    buttonText,
    setTriggerName,
    latestVersion,
    setIsOpen,
    isOpen,
    isDataFetched,
    isReady,
    isLoading,
    drawerMeta,
    formMethods,
  } = useTriggerRun(triggerName)

  // Set the trigger name in launch form state so LaunchFormButtons can use it
  useEffect(() => {
    if (triggerName) {
      setTriggerName(triggerName)
    }
    return () => {
      setTriggerName(null)
    }
  }, [triggerName, setTriggerName])

  // Track if user clicked while data was loading
  const pendingOpenRef = useRef(false)

  // Auto-open the form once data is ready (if user clicked while loading)
  useEffect(() => {
    if (!isOpen && isDataFetched && latestVersion && pendingOpenRef.current) {
      setIsOpen(true)
      pendingOpenRef.current = false
    }
  }, [isOpen, isDataFetched, latestVersion, setIsOpen])

  const handleClick = () => {
    // If data is ready, open immediately
    if (latestVersion && isDataFetched) {
      setIsOpen(true)
      pendingOpenRef.current = false
    } else if (latestVersion) {
      // Data is still loading, mark that we want to open when ready
      pendingOpenRef.current = true
    }
  }

  if (!triggerName) return null

  return (
    <>
      <Button
        type="button"
        className="border-none !px-3 [&::after]:shadow-none [&::before]:shadow-none"
        color="union"
        onClick={handleClick}
        disabled={!isReady || isLoading}
      >
        <span className="flex items-center">
          <PlayIcon className="mr-2 size-4 text-(--union-on-union)" />
          {isLoading ? 'Loading...' : buttonText}
        </span>
      </Button>
      <LaunchFormDrawer
        drawerMeta={{
          ...drawerMeta,
          breadcrumbs: [
            { label: 'Trigger', value: triggerName.name },
            { label: 'Task', value: triggerName.taskName },
            { label: 'Version', value: latestVersion || '' },
          ],
        }}
        formMethods={formMethods}
        isDataFetched={isDataFetched}
      />
    </>
  )
}
