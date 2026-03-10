'use client'

import { CodeTabContent } from '@/components/CodeTab/CodeTabContent'
import { TaskDetailsPageParams } from './types'
import { useParams } from 'next/navigation'
import { useOrg } from '@/hooks/useOrg'
import { useTaskDetails, getTaskIdentifier } from '@/hooks/useTaskDetails'
import React, { useMemo } from 'react'

export const TaskDetailsCodeTab: React.FC<{
  latestVersion?: string
  version?: string
}> = ({ latestVersion, version }) => {
  const params = useParams<TaskDetailsPageParams>()
  const org = useOrg()

  const versionToRender = version || latestVersion

  const taskDetails = useTaskDetails({
    name: params.name,
    version: versionToRender!,
    project: params.project,
    domain: params.domain,
    org,
    enabled: !!versionToRender,
  })

  const taskTemplate = taskDetails.data?.details?.spec?.taskTemplate

  const taskId = useMemo(() => {
    if (!versionToRender) return undefined
    // Use the same logic as useTaskDetails
    return getTaskIdentifier({
      domain: params.domain,
      name: params.name,
      org,
      project: params.project,
      version: versionToRender,
    })
  }, [params.domain, params.name, params.project, org, versionToRender])

  return (
    <CodeTabContent
      taskTemplate={taskTemplate}
      target={taskId ? { type: 'taskId', value: taskId } : undefined}
    />
  )
}
