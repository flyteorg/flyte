import {
  TypedInterfaceSchema,
  VariableMapSchema,
} from '@/gen/flyteidl2/core/interface_pb'
import { ActionDetails } from '@/gen/flyteidl2/workflow/run_definition_pb'
import { TaskSpec } from '@/gen/flyteidl2/task/task_definition_pb'
import { create } from '@bufbuild/protobuf'

export const isTaskSpec = (
  spec?: ActionDetails['spec'],
): spec is {
  /**
   * @generated from field: cloudidl.workflow.TaskSpec task = 6;
   */
  value: TaskSpec
  case: 'task'
} => {
  return spec?.case === 'task'
}

export const getActionSpecInterface = (actionDetails?: ActionDetails) => {
  const emptyInterface = create(TypedInterfaceSchema, {
    inputs: create(VariableMapSchema, { variables: [] }),
    outputs: create(VariableMapSchema, { variables: [] }),
  })

  if (isTaskSpec(actionDetails?.spec)) {
    return actionDetails?.spec?.value?.taskTemplate?.interface || emptyInterface
  }
  return actionDetails?.spec?.value?.interface || emptyInterface
}

export const getActionSpecInterfaceEnvironment = (
  actionDetails?: ActionDetails,
) => {
  if (isTaskSpec(actionDetails?.spec)) {
    return actionDetails?.spec?.value?.environment
  }

  return undefined
}
