'use client'

import { LoadingSpinner } from '@/components/LoadingSpinner'
import { useSelectedActionId } from '@/components/pages/RunDetails/hooks/useSelectedItem'
import { ActionAttemptIdentifierSchema } from '@/gen/flyteidl2/common/identifier_pb'
import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import { useWatchActionDetails } from '@/hooks/useWatchActionDetails'
import { CodeTabContent } from '@/components/CodeTab/CodeTabContent'
import { create } from '@bufbuild/protobuf'
import React, { useMemo } from 'react'

export const RunDetailsCodeTab: React.FC = () => {
  const selectedActionId = useSelectedActionId()
  const selectedActionDetails = useWatchActionDetails(selectedActionId)
  const { spec } = selectedActionDetails.data || {}
  const { taskTemplate } = (spec?.value as TaskSpec) || {}

  const actionAttemptId = useMemo(() => {
    const actionDetails = selectedActionDetails.data
    if (!actionDetails?.id) return undefined

    const latestAttempt =
      actionDetails.attempts?.[actionDetails.attempts.length - 1]
    const attempt = latestAttempt?.attempt ?? 1

    return create(ActionAttemptIdentifierSchema, {
      actionId: actionDetails.id,
      attempt,
    })
  }, [selectedActionDetails.data])

  if (selectedActionDetails.isLoading) {
    return (
      <div className="flex h-full w-full items-center justify-center">
        <LoadingSpinner />
      </div>
    )
  }

  if (selectedActionDetails.error) {
    return (
      <div className="flex h-full w-full items-center justify-center">
        <div className="text-sm text-red-600 dark:text-red-400">
          Error loading action details:{' '}
          {selectedActionDetails.error?.message || 'Unknown error'}
        </div>
      </div>
    )
  }

  if (!taskTemplate || !actionAttemptId) {
    return (
      <div className="flex h-full w-full items-center justify-center">
        <div className="text-sm text-(--system-gray-5)">
          {!taskTemplate
            ? 'No task template available for this action.'
            : 'No action attempt ID available.'}
        </div>
      </div>
    )
  }

  return (
    <CodeTabContent
      taskTemplate={taskTemplate}
      target={
        actionAttemptId
          ? { type: 'actionAttemptId', value: actionAttemptId }
          : undefined
      }
    />
  )
}
