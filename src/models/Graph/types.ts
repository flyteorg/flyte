import { NodeExecution } from 'models/Execution/types';
import { TaskTemplate } from 'models/Task/types';

/**
 * @TODO Deprecate this type once new graphUX is complete
 * A flyte-graph compatible node representation which also includes all of the
 * additional task data needed for our custom rendering
 */
export interface DAGNode {
    execution?: NodeExecution;
    id: string;
    parentIds?: string[];
    taskTemplate?: TaskTemplate;
}

/* Types of nodes */
export enum dTypes {
    task,
    primary,
    branch,
    subworkflow,
    start,
    end,
    nestedEnd,
    nestedStart,
    nestedMaxDepth,
    staticNode,
    staticNestedNode
}

/**
 * DAG edge
 * @sourceId    dNode.id
 * @targetId    dNode.id
 */
export interface dEdge {
    id: string;
    sourceId: string;
    targetId: string;
}

/**
 * DAG node
 * @id      used for mapping edges
 * @type    determines which UX component to render
 * @name    for display in UX
 * @value   flyte node data bound to this node
 * @nodes   children
 * @edges   edges
 */
export interface dNode {
    id: string;
    scopedId: string;
    type: dTypes;
    name: string;
    value?: any;
    nodes: Array<dNode>;
    edges: Array<dEdge>;
    expanded?: boolean;
    level?: number;
    execution?: NodeExecution;
}
