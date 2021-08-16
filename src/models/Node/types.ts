import { Core } from 'flyteidl';
import { Alias, Binding, Identifier } from 'models/Common/types';
import { TaskTemplate } from 'models/Task/types';

/** A graph node indicating a subworkflow execution */
export type WorkflowNode = Core.IWorkflowNode;
/** A graph node indicating a branching decision. */
export type BranchNode = Core.IBranchNode;

/** A graph node indicating a task to be executed. This is the most common
 * node type in a Flyte graph.
 */
export interface TaskNode extends Core.ITaskNode {
    referenceId: Identifier;
}

/** Additional, optional metadata attached to a `CompiledNode` instance */
export interface CompiledNodeMetadata extends Core.INodeMetadata {
    name: string;
    timeout?: any;
    retries?: any;
}

/** The shared properties used by all node types in a compiled Workflow graph
 * closure. Additional type-specific properties will be contained in
 * `branchNode`/`taskNode`/`workflowNode`.
 */
export interface CompiledNode extends Core.INode {
    branchNode?: BranchNode;
    id: string;
    inputs?: Binding[];
    metadata?: CompiledNodeMetadata;
    outputAliases?: Alias[];
    taskNode?: TaskNode;
    upstreamNodeIds?: string[];
    workflowNode?: WorkflowNode;
}

/** Holds all connections/edges for a given `CompiledNode` */
export interface CompiledNodeConnection {
    ids: string[];
}

/** Contains all connections in a given workflow graph template */
export interface ConnectionSet extends Core.IConnectionSet {
    downstream: Record<string, CompiledNodeConnection>;
    upstream: Record<string, CompiledNodeConnection>;
}

export interface NodeId {
    workflowId: Identifier;
    nodeId: string;
}

export interface GloballyUniqueNode {
    id: NodeId;
    node: CompiledNode;
}
