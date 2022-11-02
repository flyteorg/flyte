import * as React from 'react';
import { useState } from 'react';
import { Badge, Button, withStyles } from '@material-ui/core';
import { TaskNames } from 'components/Executions/ExecutionDetails/Timeline/TaskNames';
import { dNode } from 'models/Graph/types';
import { isExpanded } from 'components/WorkflowGraph/utils';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import { nodeExecutionPhaseConstants } from 'components/Executions/constants';
import { LaunchFormDialog } from 'components/Launch/LaunchForm/LaunchFormDialog';
import { useNodeExecutionContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import {
  graphButtonContainer,
  graphButtonStyle,
  leftPositionStyle,
  popupContainerStyle,
} from './commonStyles';
import t from './strings';

interface PausedTasksComponentProps {
  pausedNodes: dNode[];
  initialIsVisible?: boolean;
}

const CustomBadge = withStyles({
  badge: {
    backgroundColor: nodeExecutionPhaseConstants[NodeExecutionPhase.PAUSED].nodeColor,
    color: COLOR_SPECTRUM.white.color,
  },
})(Badge);

export const PausedTasksComponent: React.FC<PausedTasksComponentProps> = ({
  pausedNodes,
  initialIsVisible = false,
}) => {
  const { compiledWorkflowClosure } = useNodeExecutionContext();
  const [isVisible, setIsVisible] = useState(initialIsVisible);
  const [showResumeForm, setShowResumeForm] = useState<boolean>(false);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);

  const toggleVisibility = () => {
    setIsVisible(!isVisible);
  };

  const toggleNode = (id: string, scopeId: string, level: number) => {
    const searchNode = (nodes: dNode[], nodeLevel: number) => {
      if (!nodes || nodes.length === 0) {
        return;
      }
      for (let i = 0; i < nodes.length; i++) {
        const node = nodes[i];
        if (node.id === id && node.scopedId === scopeId && nodeLevel === level) {
          nodes[i].expanded = !nodes[i].expanded;
          return;
        }
        if (node.nodes.length > 0 && isExpanded(node)) {
          searchNode(node.nodes, nodeLevel + 1);
        }
      }
    };
    searchNode(pausedNodes, 0);
  };

  const onResumeClick = (nodeId: string) => {
    setSelectedNodeId(nodeId);
    setShowResumeForm(true);
  };

  const compiledNode = (compiledWorkflowClosure?.primary.template.nodes ?? []).find(
    (node) => node.id === selectedNodeId,
  );

  const renderPausedTasksBlock = () => (
    <div style={popupContainerStyle} data-testid="paused-tasks-table">
      <TaskNames nodes={pausedNodes} onToggle={toggleNode} onAction={onResumeClick} />
    </div>
  );

  return (
    <div style={leftPositionStyle}>
      <div>
        {isVisible ? renderPausedTasksBlock() : null}
        <div style={graphButtonContainer}>
          <CustomBadge badgeContent={pausedNodes.length}>
            <Button
              style={graphButtonStyle}
              color="default"
              id="graph-paused-tasks"
              onClick={toggleVisibility}
              variant="contained"
              title={t('pausedTasksButton')}
            >
              {t('pausedTasksButton')}
            </Button>
          </CustomBadge>
        </div>
      </div>
      {compiledNode && selectedNodeId ? (
        <LaunchFormDialog
          compiledNode={compiledNode}
          initialParameters={undefined}
          nodeId={selectedNodeId}
          showLaunchForm={showResumeForm}
          setShowLaunchForm={setShowResumeForm}
        />
      ) : null}
    </div>
  );
};
