import { Admin, Core, Protobuf } from 'flyteidl';
import { Identifier, LiteralMap, LiteralMapBlob, TaskLog, UrlBlob } from 'models/Common/types';
import { ExecutionMode, NodeExecutionPhase, TaskExecutionPhase, WorkflowExecutionPhase } from './enums';

export type WorkflowExecutionIdentifier = RequiredNonNullable<Core.IWorkflowExecutionIdentifier>;
export type ExecutionError = RequiredNonNullable<Core.IExecutionError>;

export interface BaseExecutionClosure {
  createdAt: Protobuf.ITimestamp;
  duration?: Protobuf.Duration;
  error?: ExecutionError;
  startedAt?: Protobuf.ITimestamp;
}

export interface ExecutionClosure extends Admin.IExecutionClosure {
  workflowId: Identifier;
  createdAt: Protobuf.ITimestamp;
  startedAt?: Protobuf.ITimestamp;
  duration?: Protobuf.Duration;
  computedInputs?: LiteralMap;
  outputs?: LiteralMapBlob;
  phase: WorkflowExecutionPhase;
  error?: ExecutionError;
  stateChangeDetails?: Admin.IExecutionStateChangeDetails;
}

export interface ExecutionMetadata extends Admin.IExecutionMetadata {
  mode: ExecutionMode;
  principal: string;
  nesting: number;
  referenceExecution?: WorkflowExecutionIdentifier;
  parentNodeExecution?: NodeExecutionIdentifier;
}

export interface ExecutionSpec extends Admin.IExecutionSpec {
  authRole?: Admin.IAuthRole;
  securityContext?: Core.ISecurityContext;
  inputs: LiteralMap;
  launchPlan: Identifier;
  metadata: ExecutionMetadata;
  notifications: RequiredNonNullable<Admin.INotificationList>;
}

export interface Execution extends Admin.IExecution {
  closure: ExecutionClosure;
  id: WorkflowExecutionIdentifier;
  spec: ExecutionSpec;
}

/** Node executions **/
export interface WorkflowNodeMetadata {
  executionId: WorkflowExecutionIdentifier;
}

export interface CatalogMetadata extends Core.ICatalogMetadata {
  artifactTag?: RequiredNonNullable<Core.ICatalogArtifactTag>;
  datasetId: Identifier;
  sourceTaskExecution: TaskExecutionIdentifier;
}

export interface TaskNodeMetadata extends Admin.ITaskNodeMetadata {
  cacheStatus: Core.CatalogCacheStatus;
  catalogKey?: CatalogMetadata;
}
export interface NodeExecutionIdentifier extends Core.INodeExecutionIdentifier {
  nodeId: string;
  executionId: WorkflowExecutionIdentifier;
}

export interface NodeExecutionMetadata extends Admin.INodeExecutionMetaData {
  retryGroup?: string;
  isParentNode?: boolean;
  specNodeId?: string;
}

export interface NodeExecution extends Admin.INodeExecution {
  id: NodeExecutionIdentifier;
  inputUri: string;
  closure: NodeExecutionClosure;
  metadata?: NodeExecutionMetadata;
  scopedId?: string;
}

export interface NodeExecutionsById {
  [k: string]: NodeExecution;
}

export interface NodeExecutionClosure extends Admin.INodeExecutionClosure {
  createdAt: Protobuf.ITimestamp;
  duration?: Protobuf.Duration;
  error?: ExecutionError;
  outputUri: string;
  phase: NodeExecutionPhase;
  startedAt?: Protobuf.ITimestamp;
  taskNodeMetadata?: TaskNodeMetadata;
  workflowNodeMetadata?: WorkflowNodeMetadata;
}

/** Task executions **/

export interface TaskExecutionIdentifier extends Core.ITaskExecutionIdentifier {
  taskId: Identifier;
  nodeExecutionId: NodeExecutionIdentifier;
  retryAttempt: number;
}

export interface TaskExecution extends Admin.ITaskExecution {
  id: TaskExecutionIdentifier;
  inputUri: string;
  isParent: boolean;
  closure: TaskExecutionClosure;
}
export interface TaskExecutionClosure extends Admin.ITaskExecutionClosure {
  createdAt: Protobuf.ITimestamp;
  duration?: Protobuf.Duration;
  error?: ExecutionError;
  logs?: TaskLog[];
  outputUri: string;
  phase: TaskExecutionPhase;
  startedAt?: Protobuf.ITimestamp;
}

/** Execution data */
export interface ExecutionData {
  inputs: UrlBlob;
  outputs: UrlBlob;
  fullInputs: LiteralMap | null;
  fullOutputs: LiteralMap | null;
}
