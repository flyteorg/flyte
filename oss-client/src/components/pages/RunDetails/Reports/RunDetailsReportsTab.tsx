import { Button } from '@/components/Button'
import { ChartIcon } from '@/components/icons/ChartIcon'
import { useSelectedActionId } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import { ArtifactType } from '@/gen/clouddataproxy/payload_pb'
import { useRunReports } from '@/hooks/useRunReports'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { isTaskSpec } from '@/lib/actionSpecUtils'
import { isActionInitializing, isActionTerminal } from '@/lib/actionUtils'
import { ArrowUpRightIcon } from '@heroicons/react/24/solid'
import { isNumber } from 'lodash'
import React, { useState } from 'react'
import { useSelectedAttemptStore } from '../state/AttemptStore'
import { ReportDialog } from './ReportDialog'
import { ReportIframe } from './ReportIframe'
import { DOCS_BYOC_USER_GUIDE_URL } from '@/lib/constants'

const texts = {
  loading: 'Loading reports...',
  noData: 'No reports data',
  error:
    "Reports have been deleted due to your organization's data retention policy",
}

export const RunDetailsReportsTab: React.FC = ({}) => {
  const selectedActionId = useSelectedActionId()
  const selectedActionDetails = useWatchActionDetails(selectedActionId)
  const { selectedAttempt } = useSelectedAttemptStore()

  const [isOpen, setIsOpen] = useState(false)

  const isTerminal = isActionTerminal(selectedActionDetails.data)
  const isInitializing = isActionInitializing(selectedActionDetails.data)
  const generatesDeck = isTaskSpec(selectedActionDetails.data?.spec)
    ? !!selectedActionDetails.data?.spec.value?.taskTemplate?.metadata
        ?.generatesDeck
    : false

  const reports = useRunReports({
    artifactType: ArtifactType.REPORT,
    attempt: selectedAttempt?.attempt,
    actionId: selectedActionDetails.data?.id,
    enabled:
      !isInitializing &&
      !!selectedActionDetails.data &&
      generatesDeck &&
      isNumber(selectedAttempt?.attempt),
    isActionTerminal: isTerminal,
  })

  if (!reports.data) {
    let text = ''
    let showLink = false

    if (!!selectedActionDetails.data && !generatesDeck) {
      // only show no data message if the action does not generate reports
      text = texts.noData
      showLink = true
    } else if (reports.isError && isTerminal) {
      // Only show retention error when run is terminal (avoid flash while still running)
      text = texts.error
    } else {
      // Loading, or error on non-terminal run (reports may not exist yet)
      text = texts.loading
    }

    return (
      <div className="p-8 pt-2.5">
        <div
          className={`flex w-full flex-col items-center justify-center gap-3 rounded-2xl border border-white/15 bg-(--system-black) p-4 ${showLink ? 'h-[260px] items-center' : 'h-[145px]'}`}
        >
          <div className="flex gap-2 text-(--system-gray-5)">
            <ChartIcon className="mt-0.5 size-5" />
            <span className="max-w-[255px] text-left text-sm">{text}</span>
          </div>

          {showLink ? (
            <>
              <span className="max-w-[380px] text-center text-sm text-(--system-gray-5)">
                The reports feature allows you to display and update custom
                output in the UI during task execution.
              </span>
              <Button
                outline
                href={`${DOCS_BYOC_USER_GUIDE_URL}/task-programming/reports/`}
                target="_blank"
              >
                <span className="text-2xs">How to create report</span>
                <ArrowUpRightIcon
                  data-slot="icon"
                  className="!size-3 text-(--system-gray-5)"
                />
              </Button>
            </>
          ) : null}
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="relative h-full p-8 pt-2.5">
        <ReportIframe
          reportUrl={reports.data?.[0]}
          isOpen={isOpen}
          setIsOpen={setIsOpen}
          handleRefetch={() => reports.refetch()}
        />
      </div>
      <ReportDialog isOpen={isOpen} closeDialog={() => setIsOpen(false)}>
        <div className="relative h-full">
          <ReportIframe
            reportUrl={reports.data?.[0]}
            isOpen={isOpen}
            setIsOpen={setIsOpen}
            handleRefetch={() => reports.refetch()}
          />
        </div>
      </ReportDialog>
    </>
  )
}
