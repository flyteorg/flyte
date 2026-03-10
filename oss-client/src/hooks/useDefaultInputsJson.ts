import { useMemo } from 'react'
import { NamedLiteral } from '@/gen/flyteidl2/task/common_pb'
import {
  Variable,
  VariableEntrySchema,
  VariableMapSchema,
} from '@/gen/flyteidl2/core/interface_pb'
import { useLiteralToJson } from '@/hooks/useLiteralToJsonSchema'
import { useParams } from 'next/navigation'
import { TaskDetailsPageParams } from '@/components/pages/TaskDetails/types'
import { useOrg } from '@/hooks/useOrg'
import { useTaskDetails } from '@/hooks/useTaskDetails'
import { create } from '@bufbuild/protobuf'

export function useDefaultInputsJson(
  version: string,
  taskName?: string,
  project?: string,
  domain?: string,
) {
  const params = useParams<TaskDetailsPageParams>()
  const org = useOrg()
  // Use provided parameters if available, otherwise fall back to params
  const taskNameToUse = taskName ?? params.name
  const projectToUse = project ?? params.project
  const domainToUse = domain ?? params.domain

  const taskDetails = useTaskDetails({
    name: taskNameToUse,
    version: version,
    project: projectToUse,
    domain: domainToUse,
    org: org,
    enabled: !!version && !!taskNameToUse && !!projectToUse && !!domainToUse,
  })
  const literals = useMemo(
    () =>
      taskDetails.data?.details?.spec?.defaultInputs?.map(
        (item) =>
          ({
            name: item.name,
            value: item.parameter?.behavior.value,
          }) as NamedLiteral,
      ) ?? [],
    [taskDetails.data?.details?.spec?.defaultInputs],
  )
  const variables = useMemo(
    () =>
      taskDetails.data?.details?.spec?.defaultInputs?.reduce(
        (acc, param) => {
          if (param.parameter?.var) {
            acc.variables.push(
              create(VariableEntrySchema, {
                key: `${param.name}`,
                value: param.parameter?.var as unknown as Variable,
              }),
            )
          }
          return acc
        },
        create(VariableMapSchema, { variables: [] }),
      ) ?? create(VariableMapSchema, { variables: [] }),
    [taskDetails.data?.details?.spec?.defaultInputs],
  )

  return useLiteralToJson({
    literals,
    variables,
  })
}
