import { Node } from 'components/flytegraph/Node';
import { NodeRendererProps } from 'components/flytegraph/types';
import { taskColors } from 'components/Theme/constants';
import { DAGNode } from 'models/Graph/types';
import { TaskType } from 'models/Task/constants';
import * as React from 'react';
import { InputOutputNodeRenderer } from './InputOutputNodeRenderer';
import { isEndNode, isStartNode } from './utils';

const TaskNode: React.FC<NodeRendererProps<DAGNode>> = (props) => {
  const { node, config } = props;
  let fillColor = taskColors[TaskType.UNKNOWN];
  if (node.data && node.data.taskTemplate) {
    const mappedColor = taskColors[node.data.taskTemplate.type as TaskType];
    if (mappedColor) {
      fillColor = mappedColor;
    }
  }
  return <Node {...props} config={{ ...config, fillColor }} />;
};

/** Assigns colors to DAGNodes based on the type of task contained in `data` */
export const TaskNodeRenderer: React.FC<NodeRendererProps<DAGNode>> = (props) => {
  if (isStartNode(props.node)) {
    return <InputOutputNodeRenderer {...props} label="inputs" />;
  }
  if (isEndNode(props.node)) {
    return <InputOutputNodeRenderer {...props} label="outputs" />;
  }
  return <TaskNode {...props} />;
};
