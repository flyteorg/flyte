import { Fragment } from 'react'
import { FormProvider, UseFormReturn } from 'react-hook-form'
import { ChevronRightIcon } from '@/components/icons/ChevronRightIcon'
import Drawer from '@/components/Drawer'
import { LaunchFormTabs } from './Tabs/LaunchFormTabs'
import { LaunchFormState } from './Tabs/types'
import { useLaunchFormState } from '@/hooks/useLaunchFormState'

interface LaunchFormDrawer {
  formMethods: UseFormReturn<LaunchFormState, unknown, LaunchFormState>
  drawerMeta: {
    title: string
    breadcrumbs: { label?: string; value?: string }[]
  }
  isDataFetched: boolean
  // allows consumers to pass custom handler if necessary
  setIsLaunchFormOpen?: () => void
}

export const LaunchFormDrawer = ({
  formMethods,
  isDataFetched,
  drawerMeta,
  ...props
}: LaunchFormDrawer) => {
  const { isOpen, setIsOpen } = useLaunchFormState()
  const setIsLaunchFormOpen = props.setIsLaunchFormOpen || setIsOpen
  // Only allow the drawer to be open if data is fetched
  const effectiveIsOpen = isOpen && isDataFetched
  return (
    <Drawer
      isOpen={effectiveIsOpen}
      setIsOpen={setIsLaunchFormOpen}
      tabs={
        isDataFetched ? (
          <FormProvider {...formMethods}>
            <LaunchFormTabs />
          </FormProvider>
        ) : null
      }
      size={818}
      hasFullscreen
      title={drawerMeta.title}
      titleSection={
        <div className="flex min-w-0 flex-1 items-center gap-1.5 overflow-hidden">
          {drawerMeta.breadcrumbs.map(({ label, value }, index, arr) => (
            <Fragment key={`${label}-${value}-${index}`}>
              <div className="flex min-w-0 gap-1 rounded-md bg-(--bg-gray) px-2 py-0.5">
                {label ? (
                  <span className="shrink-0 text-2xs dark:text-(--system-gray-5)">
                    {label}
                  </span>
                ) : null}
                <span className="truncate text-2xs dark:text-(--system-white)">
                  {value}
                </span>
              </div>

              {index + 1 === arr.length ? null : (
                <ChevronRightIcon className="shrink-0" height={6} />
              )}
            </Fragment>
          ))}
        </div>
      }
    />
  )
}
