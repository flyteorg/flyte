import { LogViewer } from '@/components/LogViewer'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'
import { App, Condition } from '@/gen/flyteidl2/app/app_definition_pb'
import {
  LogLine,
  LogLineOriginator,
} from '@/gen/flyteidl2/logs/dataplane/payload_pb'
import { useEffect, useState } from 'react'
import { AppLogType, LogSwitch } from './LogSwitch'

const mapAppConditionToLogline = (condition: Condition): LogLine =>
  ({
    message: condition.message,
    originator: LogLineOriginator.SYSTEM,
    timestamp: condition.lastTransitionTime,
  }) as LogLine

export const AppLogsTab = ({ app }: { app: App | undefined }) => {
  const [logsType, setLogsType] = useState<AppLogType>(AppLogType.APP)
  const [scalingLogs, setScalingLogs] = useState(
    app?.status?.conditions?.map(mapAppConditionToLogline) ?? [],
  )

  useEffect(() => {
    setScalingLogs(
      app?.status?.conditions?.map(mapAppConditionToLogline) ?? [],
    )
  }, [app?.status?.conditions])

  const isAppLogs = logsType === AppLogType.APP

  return (
    <>
      <LogSwitch currentValue={logsType} onChange={setLogsType} />
      <div className="flex h-full w-full flex-col gap-3 overflow-hidden rounded-2xl border border-(--system-gray-3) bg-(--system-black) px-5 py-3">
        {isAppLogs ? (
          <div className="flex min-h-0 flex-1 flex-col items-center justify-center">
            <LicensedEditionPlaceholder title="Logs" fullWidth hideBorder />
          </div>
        ) : (
          <LogViewer
            done={true}
            enableSourceFilter={false}
            logType={logsType}
            logs={scalingLogs}
            shouldSkipIcon={true}
            waiting={false}
          />
        )}
      </div>
    </>
  )
}
