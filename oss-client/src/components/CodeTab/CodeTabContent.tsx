'use client'

import React from 'react'
import { LicensedEditionPlaceholder } from '@/components/pages/SettingsUserManagement/LicensedEditionPlaceholder'

export type CodeTabTarget =
  | { type: 'actionAttemptId'; value: import('@/gen/flyteidl2/common/identifier_pb').ActionAttemptIdentifier }
  | { type: 'taskId'; value: import('@/gen/flyteidl2/task/task_definition_pb').TaskIdentifier }
  | { type: 'appId'; value: import('@/gen/flyteidl2/app/app_definition_pb').Identifier }

export interface CodeTabContentProps {
  taskTemplate?: import('@/gen/flyteidl2/task/task_definition_pb').TaskSpec['taskTemplate']
  container?: import('@/gen/flyteidl2/core/tasks_pb').Container
  target?: CodeTabTarget
  noPadding?: boolean
}

export const CodeTabContent: React.FC<CodeTabContentProps> = ({
  noPadding = false,
}) => {
  return (
    <div
      className={`flex min-w-0 flex-1 flex-col ${noPadding ? '' : 'p-8 pt-2.5'}`}
    >
      <LicensedEditionPlaceholder fullWidth title="Code" />
    </div>
  )
}
