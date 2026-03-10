import { Button } from '@/components/Button'
import { RefreshIcon } from '@/components/icons/RefreshIcon'
import { Tooltip } from '@/components/Tooltip'
import {
  ArrowsPointingInIcon,
  ArrowsPointingOutIcon,
} from '@heroicons/react/24/solid'
import React from 'react'

// Taken from https://developer.mozilla.org/en-US/docs/Web/HTML/Element/iframe#sandbox
const sandboxRules = [
  'allow-forms',
  'allow-modals',
  'allow-orientation-lock',
  'allow-pointer-lock',
  'allow-popups',
  'allow-popups-to-escape-sandbox',
  'allow-presentation',
  'allow-same-origin',
  'allow-scripts',
  'allow-top-navigation-by-user-activation',
  'allow-downloads',
].join(' ')

type ReportIframeProps = {
  reportUrl: string | undefined
  isOpen: boolean
  setIsOpen: React.Dispatch<React.SetStateAction<boolean>>
  handleRefetch: () => void
}

export function ReportIframe({
  reportUrl,
  isOpen,
  setIsOpen,
  handleRefetch,
}: ReportIframeProps) {
  return (
    <>
      {reportUrl ? (
        <div
          className={`${isOpen ? '' : 'p-10'} absolute top-3 right-3 flex items-center gap-5`}
        >
          <Tooltip content="Refresh" placement="bottom">
            <Button
              title="Refresh"
              plain
              color="zinc"
              onClick={() => handleRefetch()}
              className="!size-6 hover:!bg-zinc-100"
            >
              <RefreshIcon data-slot="icon" className="!text-zinc-400" />
            </Button>
          </Tooltip>

          {isOpen ? (
            <Button
              outline
              size="xs"
              color="zinc"
              onClick={() => setIsOpen((prev) => !prev)}
              className="!border-[#E6E6E6] text-zinc-400 hover:!bg-transparent hover:text-(--system-black)"
            >
              Exit fullscreen
              <ArrowsPointingInIcon className="!size-3.5 fill-zinc-400" />
            </Button>
          ) : (
            <Tooltip content="Fullscreen" placement="bottom">
              <Button
                title="Fullscreen"
                plain
                color="zinc"
                onClick={() => setIsOpen((prev) => !prev)}
                className="!size-6 hover:!bg-zinc-100"
              >
                <ArrowsPointingOutIcon className="!size-5 fill-zinc-400" />
              </Button>
            </Tooltip>
          )}
        </div>
      ) : null}

      <iframe
        src={reportUrl}
        title="Report"
        width="100%"
        height="100%"
        className="h-full rounded-2xl"
        sandbox={sandboxRules}
      />
    </>
  )
}
