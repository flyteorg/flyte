import {
  nodeExecutionPhaseConstants,
  workflowExecutionPhaseConstants,
} from 'components/Executions/constants';
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
    data: workflowExecutionPhaseConstants[WorkflowExecutionPhase.QUEUED].value,
  },
  running: {
    label: 'In Progress',
    value: 'running',
    data: workflowExecutionPhaseConstants[WorkflowExecutionPhase.RUNNING].value,
  },
  succeeded: {
    label: 'Successful',
    value: 'succeeded',
    data: workflowExecutionPhaseConstants[WorkflowExecutionPhase.SUCCEEDED].value,
  },
  failed: {
    label: 'Failed',
    value: 'failed',
    data: workflowExecutionPhaseConstants[WorkflowExecutionPhase.FAILED].value,
  },
  aborted: {
    label: 'Aborted',
    value: 'aborted',
    data: workflowExecutionPhaseConstants[WorkflowExecutionPhase.ABORTED].value,
  },
  unknown: {
    label: 'Unknown',
    value: 'unknown',
    data: workflowExecutionPhaseConstants[WorkflowExecutionPhase.UNDEFINED].value,
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
  | 'paused'
  | 'unknown';

/** A set of WorkflowExecution status/phase filters to be consumed by a MultiFilterState.
 */
export const nodeExecutionStatusFilters: FilterMap<NodeExecutionStatusFilterKey, string> = {
  queued: {
    label: 'Scheduled',
    value: 'queued',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.QUEUED].value,
  },
  running: {
    label: 'In Progress',
    value: 'running',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.RUNNING].value,
  },
  succeeded: {
    label: 'Successful',
    value: 'succeeded',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.SUCCEEDED].value,
  },
  failed: {
    label: 'Failed',
    value: 'failed',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.FAILED].value,
  },
  aborted: {
    label: 'Aborted',
    value: 'aborted',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.ABORTED].value,
  },
  timedOut: {
    label: 'Timed Out',
    value: 'timedOut',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.TIMED_OUT].value,
  },
  skipped: {
    label: 'Skipped',
    value: 'skipped',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.SKIPPED].value,
  },
  paused: {
    label: 'Paused',
    value: 'paused',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.PAUSED].value,
  },
  unknown: {
    label: 'Unknown',
    value: 'unknown',
    data: nodeExecutionPhaseConstants[NodeExecutionPhase.UNDEFINED].value,
  },
};
