import { Admin, Core, Protobuf } from 'flyteidl';
import {
  Container,
  Identifier,
  RetryStrategy,
  RuntimeMetadata,
  TypedInterface,
} from 'models/Common/types';

/** Additional, optional metadata pertaining to a task template */

export interface TaskMetadata extends Core.ITaskMetadata {
  discoverable?: boolean;
  runtime?: RuntimeMetadata;
  retries?: RetryStrategy;
  discoveryVersion?: string;
  deprecated?: string;
}

/** Represents a task that can be executed independently, usually as a node in a
 * Workflow graph
 */
export interface TaskTemplate extends Core.ITaskTemplate {
  container?: Container;
  custom?: Protobuf.IStruct;
  id: Identifier;
  interface?: TypedInterface;
  metadata?: TaskMetadata;
  closure?: TaskClosure;
  type: string;
}

/** An instance of a task which has been serialized into a `TaskClosure` */
export interface CompiledTask extends Core.ICompiledTask {
  template: TaskTemplate;
}

/** A serialized version of all information needed to execute a task */
export interface TaskClosure extends Admin.ITaskClosure {
  compiledTask: CompiledTask;
  createdAt: Protobuf.Timestamp;
}

export interface Task extends Admin.ITask {
  id: Identifier;
  closure: TaskClosure;
}
