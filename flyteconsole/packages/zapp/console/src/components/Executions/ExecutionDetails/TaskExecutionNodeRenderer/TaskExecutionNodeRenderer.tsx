import { NodeRendererProps } from 'components/flytegraph/types';
import { TaskNodeRenderer } from 'components/WorkflowGraph/TaskNodeRenderer';
import { isEndNode, isStartNode } from 'components/WorkflowGraph/utils';
import { DAGNode } from 'models/Graph/types';
import * as React from 'react';
import { TaskExecutionNode } from './TaskExecutionNode';

/** Renders DAGNodes with colors based on their node type, as well as dots to
 * indicate the execution status
 */
export const TaskExecutionNodeRenderer: React.FC<NodeRendererProps<DAGNode>> = (props) => {
  if (isStartNode(props.node) || isEndNode(props.node)) {
    return <TaskNodeRenderer {...props} />;
  }
  return <TaskExecutionNode {...props} />;
};
