import { ErrorBannerForm } from '@/components/ErrorBanner'
import { Tabs, TabType } from '@/components/Tabs'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'
import { ExclamationCircleIcon } from '@heroicons/react/24/solid'
import React, { useEffect, useMemo } from 'react'
import { useFormContext } from 'react-hook-form'
import { LaunchFormButtons } from './LaunchFormButtons'
import { LaunchFormContext } from './LaunchFormContext'
import { LaunchFormEnvVars } from './LaunchFormEnvVars'
import { LaunchFormInputs } from './LaunchFormInputs'
import { LaunchFormLabels } from './LaunchFormLabels'
import { LaunchFormSettings } from './LaunchFormSettings'
import { LaunchFormState, LaunchFormTab } from './types'

export const TabLayout = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className="flex h-full flex-col justify-between dark:text-(--accent-gray)">
      <div className="flex-1 overflow-y-auto p-2 px-4">{children}</div>
    </div>
  )
}

const tabErrors: Record<LaunchFormTab, (keyof LaunchFormState)[]> = {
  inputs: ['inputs'],
  context: ['context'],
  settings: ['runName'],
  'env-vars': ['envs'],
  labels: ['labels'],
  debug: [],
}

export const LaunchFormTabs = () => {
  const { launchFormTab, setLaunchFormTab } = useLaunchFormState()
  const { clearErrors, formState } = useFormContext<LaunchFormState>()
  const errors = formState.errors

  const handleTabClick = (newTab: LaunchFormTab) => {
    setLaunchFormTab(newTab)
  }

  const launchFormTabConfig: TabType<LaunchFormTab>[] = useMemo(() => {
    const tabs: {
      path: LaunchFormTab
      label: string
      content: React.ReactNode
    }[] = [
      { path: 'inputs', label: 'Inputs', content: <LaunchFormInputs /> },
      { path: 'context', label: 'Context', content: <LaunchFormContext /> },
      { path: 'settings', label: 'Settings', content: <LaunchFormSettings /> },
      { path: 'env-vars', label: 'Env vars', content: <LaunchFormEnvVars /> },
      { path: 'labels', label: 'Labels', content: <LaunchFormLabels /> },
    ]
    return tabs.map((tab) => {
      const hasError = tabErrors[tab.path].some((key) => errors[key])
      return {
        ...tab,
        trailingIcon: hasError ? (
          <>
            <span className="sr-only">, has errors</span>
            <ExclamationCircleIcon
              className="size-3.5 shrink-0 text-red-500"
              aria-hidden
            />
          </>
        ) : undefined,
      }
    })
    // Intentionally depend on specific error keys so config only updates when tab errors change
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    errors.inputs,
    errors.context,
    errors.runName,
    errors.envs,
    errors.labels,
  ])

  useEffect(() => {
    return () => {
      setLaunchFormTab('inputs')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      {/* api errors */}
      {formState.errors.root?.message ? (
        <div className="p-5">
          <ErrorBannerForm
            message={formState.errors.root.message}
            onClose={() => clearErrors('root')}
          />
        </div>
      ) : null}
      <Tabs<LaunchFormTab>
        currentTab={launchFormTab}
        onClickTab={handleTabClick}
        tabs={launchFormTabConfig}
        classes={{
          tabListContainer: 'px-4! py-3! border-b-1 border-(--bg-gray)',
        }}
      />
      <div className="shrink-0 border-t border-(--bg-gray)">
        <div className="flex gap-3 px-6 py-4">
          <LaunchFormButtons />
        </div>
      </div>
    </>
  )
}
