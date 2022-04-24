import { NodeExecutionPhase, WorkflowExecutionPhase } from 'models/Execution/enums';
import { FilterMap } from './types';

export type WorkflowExecutionStatusFilterKey =
  | 'queued'
  | 'running'
  | 'succeeded'
  | 'failed'
  | 'aborted'
  | 'unknown';

/** A set of WorkflowExecution status/phase filters to be consumed by a MultiFilterState.
 */
export const workflowExecutionStatusFilters: FilterMap<WorkflowExecutionStatusFilterKey, string> = {
  queued: {
    label: 'Scheduled',
    value: 'queued',
    data: WorkflowExecutionPhase[WorkflowExecutionPhase.QUEUED],
  },
  running: {
    label: 'In Progress',
    value: 'running',
    data: WorkflowExecutionPhase[WorkflowExecutionPhase.RUNNING],
  },
  succeeded: {
    label: 'Successful',
    value: 'succeeded',
    data: WorkflowExecutionPhase[WorkflowExecutionPhase.SUCCEEDED],
  },
  failed: {
    label: 'Failed',
    value: 'failed',
    data: WorkflowExecutionPhase[WorkflowExecutionPhase.FAILED],
  },
  aborted: {
    label: 'Aborted',
    value: 'aborted',
    data: WorkflowExecutionPhase[WorkflowExecutionPhase.ABORTED],
  },
  unknown: {
    label: 'Unknown',
    value: 'unknown',
    data: WorkflowExecutionPhase[WorkflowExecutionPhase.UNDEFINED],
  },
};

export type NodeExecutionStatusFilterKey =
  | 'queued'
  | 'running'
  | 'succeeded'
  | 'failed'
  | 'aborted'
  | 'skipped'
  | 'timedOut'
  | 'unknown';

/** A set of WorkflowExecution status/phase filters to be consumed by a MultiFilterState.
 */
export const nodeExecutionStatusFilters: FilterMap<NodeExecutionStatusFilterKey, string> = {
  queued: {
    label: 'Scheduled',
    value: 'queued',
    data: NodeExecutionPhase[NodeExecutionPhase.QUEUED],
  },
  running: {
    label: 'In Progress',
    value: 'running',
    data: NodeExecutionPhase[NodeExecutionPhase.RUNNING],
  },
  succeeded: {
    label: 'Successful',
    value: 'succeeded',
    data: NodeExecutionPhase[NodeExecutionPhase.SUCCEEDED],
  },
  failed: {
    label: 'Failed',
    value: 'failed',
    data: NodeExecutionPhase[NodeExecutionPhase.FAILED],
  },
  aborted: {
    label: 'Aborted',
    value: 'aborted',
    data: NodeExecutionPhase[NodeExecutionPhase.ABORTED],
  },
  timedOut: {
    label: 'Timed Out',
    value: 'timedOut',
    data: NodeExecutionPhase[NodeExecutionPhase.TIMED_OUT],
  },
  skipped: {
    label: 'Skipped',
    value: 'skipped',
    data: NodeExecutionPhase[NodeExecutionPhase.SKIPPED],
  },
  unknown: {
    label: 'Unknown',
    value: 'unknown',
    data: NodeExecutionPhase[NodeExecutionPhase.UNDEFINED],
  },
};
