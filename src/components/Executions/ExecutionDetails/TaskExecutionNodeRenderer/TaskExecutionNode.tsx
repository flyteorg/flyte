import { getNodeExecutionPhaseConstants } from 'components/Executions/utils';
import { NodeRendererProps, Point } from 'components/flytegraph/types';
import { TaskNodeRenderer } from 'components/WorkflowGraph/TaskNodeRenderer';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { DAGNode } from 'models/Graph/types';
import * as React from 'react';
import { NodeExecutionsContext } from '../../contexts';
import { StatusIndicator } from './StatusIndicator';

/** Renders DAGNodes with colors based on their node type, as well as dots to
 * indicate the execution status
 */
export const TaskExecutionNode: React.FC<NodeRendererProps<DAGNode>> = (props) => {
  const { node, config, selected } = props;
  const nodeExecutions = React.useContext(NodeExecutionsContext);
  const nodeExecution = nodeExecutions[node.id];

  const phase = nodeExecution ? nodeExecution.closure.phase : NodeExecutionPhase.UNDEFINED;
  const { badgeColor: color } = getNodeExecutionPhaseConstants(phase);
  const renderStatus = phase !== NodeExecutionPhase.UNDEFINED;

  const height = config.fontSize + config.textPadding * 2;
  const width = node.textWidth + config.textPadding * 2;

  // Position status indicator centered on left border
  const statusPosition: Point = {
    x: -width / 2,
    y: 0,
  };

  const noStatusOverlayProps = {
    height,
    width,
    fill: 'white',
    opacity: selected ? 0 : 0.35,
    stroke: 'white',
    strokeWidth: selected ? 0 : config.strokeWidth,
    rx: config.cornerRounding,
    ry: config.cornerRounding,
    x: -width / 2,
    y: -height / 2,
  };

  const nodeChildren = renderStatus ? (
    <StatusIndicator
      color={color}
      config={config}
      pulse={phase === NodeExecutionPhase.RUNNING || phase === NodeExecutionPhase.DYNAMIC_RUNNING}
      position={statusPosition}
    />
  ) : (
    <rect {...noStatusOverlayProps} />
  );

  return <TaskNodeRenderer {...props}>{nodeChildren}</TaskNodeRenderer>;
};
