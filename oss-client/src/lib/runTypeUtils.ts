import {
  AbortInfo,
  ErrorInfo,
  ActionDetails,
} from '@/gen/flyteidl2/workflow/run_definition_pb'

export function isErrorInfoResult(result?: ActionDetails['result']): result is {
  /**
   * Error info for the run, if failed.
   *
   * @generated from field: cloudidl.workflow.ErrorInfo error_info = 4;
   */
  value: ErrorInfo
  case: 'errorInfo'
} {
  return result?.case === 'errorInfo'
}

export function isAbortInfoResult(result?: ActionDetails['result']): result is {
  /**
   * Abort info for the run, if aborted.
   *
   * @generated from field: cloudidl.workflow.AbortInfo abort_info = 5;
   */
  value: AbortInfo
  case: 'abortInfo'
} {
  return result?.case === 'abortInfo'
}
