import * as React from 'react';

import { Node, NodeRendererProps } from 'components/flytegraph';
import { taskColors } from 'components/Theme';
import { DAGNode } from 'models/Graph';
import { TaskType } from 'models/Task';

import { InputOutputNodeRenderer } from './InputOutputNodeRenderer';
import { isEndNode, isStartNode } from './utils';

const TaskNode: React.FC<NodeRendererProps<DAGNode>> = props => {
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
export const TaskNodeRenderer: React.FC<NodeRendererProps<DAGNode>> = props => {
    if (isStartNode(props.node)) {
        return <InputOutputNodeRenderer {...props} label="inputs" />;
    }
    if (isEndNode(props.node)) {
        return <InputOutputNodeRenderer {...props} label="outputs" />;
    }
    return <TaskNode {...props} />;
};
