import { LogViewer, LOG_VIEWER_MIN_WIDTH_PX } from '@/components/LogViewer'
import { useSelectedActionId } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import LogsExtLinksBar from '@/components/pages/RunDetails/LogsExtLinksBar'
import { RunK8sSwitch } from '@/components/pages/RunDetails/LogsK8sSwitch'
import { RunLogType } from '@/components/pages/RunDetails/types'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { useWatchClusterEvents } from '@/hooks/useWatchClusterEvents'
import { isAttemptTerminal } from '@/lib/attemptUtils'
import React, { useMemo, useState } from 'react'
import { useSelectedAttemptStore } from './state/AttemptStore'

export const RunDetailsLogsTab: React.FC<unknown> = () => {
  const selectedActionId = useSelectedActionId()
  const selectedActionDetails = useWatchActionDetails(selectedActionId)
  const attempt = useSelectedAttemptStore((s) => s.selectedAttempt)

  const [logsType, setLogsType] = useState<RunLogType>(RunLogType.RUN)

  const attemptNumber = attempt?.attempt ?? 0

  const clusterEvents = useWatchClusterEvents({
    actionDetails: selectedActionDetails.data,
    attempt: attemptNumber,
    enabled: !!selectedActionDetails.data && logsType === RunLogType.K8S,
  })

  const isTerminal = isAttemptTerminal(attempt)
  const isWaiting = useMemo(() => {
    if (logsType !== RunLogType.K8S) return false
    return !isTerminal && !attempt?.clusterEvents?.length
  }, [attempt, logsType, isTerminal])

  const isK8s = logsType === RunLogType.K8S

  return (
    <div
      className="flex h-full min-h-0 flex-col gap-5 p-8 pt-2.5"
      style={{ minWidth: LOG_VIEWER_MIN_WIDTH_PX }}
    >
      <div className="flex min-w-0 flex-row gap-x-5">
        <RunK8sSwitch onChange={setLogsType} currentValue={logsType} />
        {attempt?.logInfo && <LogsExtLinksBar logInfo={attempt.logInfo} />}
      </div>
      <div className="flex h-full w-full min-w-0 flex-col gap-3 overflow-hidden rounded-2xl border border-zinc-200 bg-(--system-black) px-5 py-3 dark:border-zinc-800">
        {isK8s ? (
          <LogViewer
            enableSourceFilter={false}
            logs={clusterEvents.data?.lines}
            done={clusterEvents.isFetched}
            error={clusterEvents.error}
            waiting={isWaiting}
            logType={RunLogType.K8S}
            shouldSkipIcon={true}
          />
        ) : (
          <div className="flex min-h-0 flex-1 flex-col items-center justify-center">
            <LicensedEditionPlaceholder title="Logs" fullWidth hideBorder />
          </div>
        )}
      </div>
    </div>
  )
}
