import { Header } from '@/components/Header'
import { NavPanelLayout } from '@/components/NavPanel'
import clsx from 'clsx'
import { PropsWithChildren, ReactNode, Suspense } from 'react'
import { CopyButton } from '../CopyButton'
import { NavPanelLayoutProps } from '../NavPanel/NavPanelLayout'

export type DetailsMetadata = {
  leftControls?: ReactNode
  title: { value: string; badge?: ReactNode }
  // subtitle.value might be a ReactNode instead of a string value, in which case 'value' will return string "[Object object]"
  // adding 'copyValue' to be the string that's returned by the copy button when 'value' is a ReactNode
  subtitle?: [
    { label: string; disableCopy?: boolean } & (
      | {
          value: string
          copyValue?: string
        }
      | {
          value: ReactNode
          copyValue: string
        }
    ),
  ]
  dataList: { label: string; value: ReactNode; className?: string }[]
}

type DetailsLayoutProps = PropsWithChildren & {
  metadata: DetailsMetadata
  actionBtn?: ReactNode
  classes?: { metadataContainer?: string; childrenContainer?: string }
  navPanelLayoutProps?: Omit<NavPanelLayoutProps, 'children'>
}

export const DetailsLayout = ({
  metadata,
  actionBtn,
  children,
  classes,
  navPanelLayoutProps,
}: DetailsLayoutProps) => {
  return (
    <Suspense>
      <main className="flex h-full w-full">
        <NavPanelLayout
          mode={navPanelLayoutProps?.mode || 'embedded'}
          initialSize={navPanelLayoutProps?.initialSize || 'wide'}
          {...navPanelLayoutProps}
        >
          <div className="flex min-h-0 w-full flex-1 flex-col overflow-hidden">
            <div className="flex-none">
              <Header showSearch={true} />
              {/* metadata */}
              <div
                className={clsx(
                  classes?.metadataContainer,
                  'flex justify-between gap-5 bg-(--system-black) px-10 py-7',
                )}
                data-testid="details-metadata"
              >
                <div className="flex items-center gap-5">
                  {metadata.leftControls && (
                    <div className="flex flex-col items-center justify-center">
                      {metadata.leftControls}
                    </div>
                  )}

                  <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-3">
                      <h2 className="min-h-7 text-xl font-medium">
                        {metadata.title.value}
                      </h2>
                      {metadata.title.badge}
                    </div>
                    {(metadata?.subtitle?.length || 0) > 0 ? (
                      <div className="flex flex-row gap-4">
                        {metadata?.subtitle?.map(
                          ({ label, value, copyValue, disableCopy }) => (
                            <div
                              key={label}
                              className="flex h-3 flex-row items-center"
                            >
                              {label && (
                                <span className="mr-1 text-2xs font-semibold tracking-tight text-(--system-gray-5)">
                                  {label}
                                </span>
                              )}
                              <span className="truncate text-2xs font-semibold tracking-tight text-(--system-gray-5)">
                                {value}
                              </span>
                              {!disableCopy && (
                                <CopyButton
                                  size="sm"
                                  className="size-4 !bg-transparent hover:dark:[&_[data-slot=icon]]:!text-(--system-white)"
                                  value={
                                    typeof value === 'string'
                                      ? value
                                      : (copyValue ?? '')
                                  }
                                />
                              )}
                            </div>
                          ),
                        )}
                      </div>
                    ) : null}
                  </div>
                </div>

                <div className="flex min-w-0 items-center gap-6 overflow-hidden">
                  <div className="hidden min-w-0 shrink items-center gap-6 overflow-hidden lg:flex">
                    {metadata.dataList.map(
                      ({ label, value, className }, index, arr) => {
                        const isLastItem = index + 1 === arr.length
                        return (
                          <div
                            key={label}
                            className={clsx(
                              `min-h-11 min-w-14 shrink-0`,
                              isLastItem && 'mr-4',
                              className,
                            )}
                          >
                            <div className="text-2xs font-semibold text-(--system-gray-6)">
                              {label}
                            </div>
                            <div className="text-sm font-medium text-(--system-white)">
                              {value}
                            </div>
                          </div>
                        )
                      },
                    )}
                  </div>
                  <div className="shrink-0">{actionBtn}</div>
                </div>
              </div>
            </div>

            <div
              className={clsx(
                classes?.childrenContainer,
                'relative flex h-full min-h-0 flex-1',
                // bottom border
                'after:absolute after:top-0 after:right-0 after:left-0 after:z-10 after:h-px after:bg-(--system-gray-2) dark:after:bg-zinc-800',
              )}
            >
              {children}
            </div>
          </div>
        </NavPanelLayout>
      </main>
    </Suspense>
  )
}
