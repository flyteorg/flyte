import { BigQueryLogoIcon } from '@/components/icons/BigQueryLogoIcon'
import { CircularProgressIcon } from '@/components/icons/CircularProgressIcon'
import { CloudwatchLogoIcon } from '@/components/icons/CloudwatchLogoIcon'
import { CometLogoIcon } from '@/components/icons/CometLogoIcon'
import { DaskLogoIcon } from '@/components/icons/DaskLogoIcon'
import { DatabricksLogoIcon } from '@/components/icons/DatabricksLogoIcon'
import { NeptuneLogoIcon } from '@/components/icons/NeptuneLogoIcon'
import { PyTorchLogoIcon } from '@/components/icons/PyTorchLogoIcon'
import { RayDashboardLogoIcon } from '@/components/icons/RayDashboardLogoIcon'
import { SnowflakeLogoIcon } from '@/components/icons/SnowflakeLogoIcon'
import { SparkDriverLogoIcon } from '@/components/icons/SparkDriverLogoIcon'
import { SparkUiLogoIcon } from '@/components/icons/SparkUiLogoIcon'
import { StackdriverLogoIcon } from '@/components/icons/StackdriverLogoIcon'
import { TensorflowLogoIcon } from '@/components/icons/TensorflowLogoIcon'
import { WandbLogoIcon } from '@/components/icons/WandbLogoIcon'
import type { ExternalLinkUrlProps } from '@/components/pages/RunDetails/types'
import {
  type TaskLog,
  TaskLog_LinkType,
} from '@/gen/flyteidl2/core/execution_pb'
import { ActionAttempt } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { useCallback, useMemo } from 'react'

export const useExternalLogUrls = (
  logInfo?: ActionAttempt['logInfo'],
  logsLinkType: TaskLog_LinkType = TaskLog_LinkType.EXTERNAL,
): ExternalLinkUrlProps[] => {
  const getLogIconByName = useCallback(
    (name: string, url: string): React.FC | undefined => {
      const lcName = name.toLowerCase()
      // spark driver
      if (lcName.startsWith('sparkdriver')) {
        return SparkDriverLogoIcon
      }
      // spark
      if (lcName.startsWith('spark')) {
        return SparkUiLogoIcon
      }
      // ray
      if (lcName.startsWith('ray')) {
        return RayDashboardLogoIcon
      }
      // wandb
      {
        let isWandbUrl = false
        if (url) {
          try {
            const parsed = new URL(url)
            const hostname = parsed.hostname.toLowerCase()
            isWandbUrl =
              hostname === 'wandb.ai' || hostname.endsWith('.wandb.ai')
          } catch {
            const lowerUrl = url.toLowerCase()
            // Fallback for non-absolute/invalid URIs:
            // attempt to extract a hostname, otherwise default to false.
            const match = lowerUrl.match(
              /^[a-z][a-z0-9+.-]*:\/\/([^/]+)\/?/,
            )
            const hostname = match ? match[1] : ''
            isWandbUrl =
              hostname === 'wandb.ai' || hostname.endsWith('.wandb.ai')
          }
        }
        if (lcName.startsWith('wandb') || isWandbUrl) {
          return WandbLogoIcon
        }
      }
      // neptune
      if (lcName.startsWith('neptune')) {
        return NeptuneLogoIcon
      }
      // dask
      if (lcName.startsWith('dask')) {
        return DaskLogoIcon
      }
      // pytorch
      if (lcName.startsWith('pytorch')) {
        return PyTorchLogoIcon
      }
      // tensorflow
      if (lcName.startsWith('tensorflow')) {
        return TensorflowLogoIcon
      }
      // bigquery
      if (lcName.startsWith('bigquery')) {
        return BigQueryLogoIcon
      }
      // databricks
      if (lcName.startsWith('databricks')) {
        return DatabricksLogoIcon
      }
      // snowflake
      if (lcName.startsWith('snowflake')) {
        return SnowflakeLogoIcon
      }
      // cloudwatch
      if (lcName.startsWith('cloudwatch')) {
        return CloudwatchLogoIcon
      }
      // stackdriver
      if (lcName.startsWith('stackdriver')) {
        return StackdriverLogoIcon
      }

      // comet
      if (lcName.startsWith('comet')) {
        return CometLogoIcon
      }
      return
    },
    [],
  )

  const isExtLog = useCallback(
    ({ linkType }: TaskLog): boolean => linkType === logsLinkType,
    [logsLinkType],
  )

  return useMemo(
    () =>
      logInfo
        ?.filter((log) => isExtLog(log))
        ?.map(({ name, uri, ready }) => ({
          name,
          url: uri,
          ready,
          icon: ready
            ? getLogIconByName(name, uri)
            : ({ className }) => (
                <CircularProgressIcon className={className} isStatic={true} />
              ),
        })) ?? [],
    [logInfo, getLogIconByName, isExtLog],
  )
}
