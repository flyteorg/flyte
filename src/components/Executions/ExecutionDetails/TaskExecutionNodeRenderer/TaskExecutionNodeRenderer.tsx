import * as React from 'react';

import { NodeRendererProps } from 'components/flytegraph';

import {
    isEndNode,
    isStartNode,
    TaskNodeRenderer
} from 'components/WorkflowGraph';
import { DAGNode } from 'models/Graph';
import { TaskExecutionNode } from './TaskExecutionNode';

/** Renders DAGNodes with colors based on their node type, as well as dots to
 * indicate the execution status
 */
export const TaskExecutionNodeRenderer: React.FC<NodeRendererProps<
    DAGNode
>> = props => {
    if (isStartNode(props.node) || isEndNode(props.node)) {
        return <TaskNodeRenderer {...props} />;
    }
    return <TaskExecutionNode {...props} />;
};
