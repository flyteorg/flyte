import * as React from 'react';
import { useState, useEffect, useContext } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import { dTypes } from 'models/Graph/types';
import { NodeExecutionPhase, TaskExecutionPhase } from 'models/Execution/enums';
import { RENDER_ORDER } from 'components/Executions/TaskExecutionsList/constants';
import { whiteColor } from 'components/Theme/constants';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import { Tooltip } from '@material-ui/core';
import { COLOR_SPECTRUM } from 'components/Theme/colorSpectrum';
import { getNodeFrontendPhase } from 'components/Executions/utils';
import { CacheStatus } from 'components/Executions/CacheStatus';
import { LaunchFormDialog } from 'components/Launch/LaunchForm/LaunchFormDialog';
import { useNodeExecutionContext } from 'components/Executions/contextProvider/NodeExecutionDetails';
import { NodeExecutionsByIdContext } from 'components/Executions/contexts';
import {
  COLOR_GRAPH_BACKGROUND,
  getGraphHandleStyle,
  getGraphNodeStyle,
  getNestedContainerStyle,
  getStatusColor,
} from './utils';
import { RFHandleProps, RFNode } from './types';
import t from './strings';

const taskContainerStyle: React.CSSProperties = {
  position: 'absolute',
  top: '-.55rem',
  zIndex: 0,
  right: '.15rem',
};

const taskTypeStyle: React.CSSProperties = {
  backgroundColor: COLOR_GRAPH_BACKGROUND,
  color: 'white',
  padding: '.1rem .2rem',
  fontSize: '.3rem',
};

const renderTaskType = (taskType: dTypes | undefined) => {
  if (!taskType) {
    return null;
  }
  return (
    <div style={taskContainerStyle}>
      <div style={taskTypeStyle}>{taskType}</div>
    </div>
  );
};

const renderBasicNode = (
  taskType: dTypes | undefined,
  text: string,
  scopedId: string,
  styles: React.CSSProperties,
  onClick?: () => void,
) => {
  return (
    <div onClick={onClick}>
      {renderTaskType(taskType)}
      <div style={styles}>{text}</div>
      {renderDefaultHandles(scopedId, getGraphHandleStyle('source'), getGraphHandleStyle('target'))}
    </div>
  );
};

export const renderDefaultHandles = (id: string, sourceStyle: any, targetStyle: any) => {
  const leftHandleProps: RFHandleProps = {
    id: `edge-left-${id}`,
    type: 'target',
    position: Position.Left,
    style: targetStyle,
  };

  const rightHandleProps: RFHandleProps = {
    id: `edge-right-${id}`,
    type: 'source',
    position: Position.Right,
    style: sourceStyle,
  };
  return (
    <>
      <Handle {...leftHandleProps} />
      <Handle {...rightHandleProps} />
    </>
  );
};

export const renderStardEndHandles = (nodeType: dTypes, scopedId: string) => {
  const isStart = nodeType === dTypes.nestedStart || nodeType === dTypes.start;
  const idPrefix = isStart ? 'start' : 'end';
  const position = isStart ? Position.Right : Position.Left;
  const type = isStart ? 'source' : 'target';

  /**
   * @TODO
   * Resepect the actual node type once toggle nested graphs is implemented
   * For now we force nestedMaxDepth for any nested types
   */
  const style =
    nodeType === dTypes.nestedStart || nodeType === dTypes.nestedEnd
      ? getGraphHandleStyle(type, dTypes.nestedMaxDepth)
      : getGraphHandleStyle(type);

  const handleProps: RFHandleProps = {
    id: `rf-handle-${idPrefix}-${scopedId}`,
    type: type,
    position: position,
    style: style,
  };

  return <Handle {...handleProps} />;
};

/**
 * Styles start/end nodes as a point; used for nested workflows
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomNestedPoint = ({ data }: RFNode) => {
  const { nodeType, scopedId } = data;
  const containerStyle = getGraphNodeStyle(nodeType);
  return (
    <>
      <div style={containerStyle} />
      {renderStardEndHandles(nodeType, scopedId)}
    </>
  );
};

/**
 * @TODO
 * This component used as a stop gap until we support fully nested
 * workflows; renders nested data (branch/workflow) as a single node
 * denoted by solid color.
 * @param props.data data property of ReactFlowGraphNodeData
 */

export const ReactFlowCustomMaxNested = ({ data }: RFNode) => {
  const { text, taskType, scopedId, onAddNestedView } = data;
  const styles = getGraphNodeStyle(dTypes.nestedMaxDepth);

  const onClick = () => {
    onAddNestedView();
  };

  return renderBasicNode(taskType, text, scopedId, styles, onClick);
};

export const ReactFlowStaticNested = ({ data }: RFNode) => {
  const { text, taskType, scopedId } = data;
  const styles = getGraphNodeStyle(dTypes.staticNestedNode);
  return renderBasicNode(taskType, text, scopedId, styles);
};

export const ReactFlowStaticNode = ({ data }: RFNode) => {
  const { text, taskType, scopedId } = data;
  const styles = getGraphNodeStyle(dTypes.staticNode);
  return renderBasicNode(taskType, text, scopedId, styles);
};

/**
 * Component that renders a map task item within the mapped node.
 * @param props.numberOfTasks number of tasks of certain completion phase
 * @param props.color string value of color of the block
 * @param props.phase phase of the current map task item
 * @param props.onPhaseSelectionChanged callback from the parent component
 */

interface TaskPhaseItemProps {
  numberOfTasks: number;
  color: string;
  phase: TaskExecutionPhase;
  setSelectedPhase: (phase: TaskExecutionPhase) => void;
  setSelectedNode: (val: boolean) => void;
}

const TaskPhaseItem = ({
  numberOfTasks,
  color,
  phase,
  setSelectedPhase,
  setSelectedNode,
}: TaskPhaseItemProps) => {
  const taskPhaseStyles: React.CSSProperties = {
    borderRadius: '2px',
    backgroundColor: color,
    color: whiteColor,
    margin: '0 1px',
    padding: '0 2px',
    fontSize: '8px',
    lineHeight: '14px',
    minWidth: '14px',
    textAlign: 'center',
    cursor: 'pointer',
  };

  const handleMapTaskClick = (e) => {
    e.stopPropagation();
    setSelectedNode(true);
    setSelectedPhase(phase);
  };

  return (
    <div style={taskPhaseStyles} onClick={handleMapTaskClick}>
      ×{numberOfTasks}
    </div>
  );
};

/**
 * Custom component used by ReactFlow.  Renders a label (text)
 * and any edge handles.
 * @param props.data data property of ReactFlowGraphNodeData
 */

export const ReactFlowGateNode = ({ data }: RFNode) => {
  const { compiledWorkflowClosure } = useNodeExecutionContext();
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);
  const { nodeType, nodeExecutionStatus, text, scopedId, onNodeSelectionChanged } = data;
  const phase = getNodeFrontendPhase(nodeExecutionStatus, true);
  const styles = getGraphNodeStyle(nodeType, phase);
  const [showResumeForm, setShowResumeForm] = useState<boolean>(false);

  const compiledNode = (compiledWorkflowClosure?.primary.template.nodes ?? []).find(
    (node) => node.id === nodeExecutionsById[scopedId]?.id?.nodeId,
  );

  const iconStyles: React.CSSProperties = {
    width: '10px',
    height: '10px',
    marginLeft: '4px',
    marginTop: '1px',
    color: COLOR_GRAPH_BACKGROUND,
    cursor: 'pointer',
  };

  const handleNodeClick = () => {
    onNodeSelectionChanged(true);
  };

  const onResumeClick = (e) => {
    e.stopPropagation();
    setShowResumeForm(true);
  };

  return (
    <div onClick={handleNodeClick}>
      <div style={styles}>
        {text}
        {phase === NodeExecutionPhase.PAUSED && (
          <Tooltip title={t('resumeTooltip')}>
            <PlayCircleOutlineIcon onClick={onResumeClick} style={iconStyles} />
          </Tooltip>
        )}
      </div>
      {renderDefaultHandles(scopedId, getGraphHandleStyle('source'), getGraphHandleStyle('target'))}
      {compiledNode && (
        <LaunchFormDialog
          compiledNode={compiledNode}
          initialParameters={undefined}
          nodeId={scopedId}
          showLaunchForm={showResumeForm}
          setShowLaunchForm={setShowResumeForm}
        />
      )}
    </div>
  );
};

/**
 * Custom component used by ReactFlow.  Renders a label (text)
 * and any edge handles.
 * @param props.data data property of ReactFlowGraphNodeData
 */

export const ReactFlowCustomTaskNode = ({ data }: RFNode) => {
  const {
    nodeType,
    nodeExecutionStatus,
    selectedPhase: initialPhase,
    taskType,
    text,
    nodeLogsByPhase,
    cacheStatus,
    scopedId,
    onNodeSelectionChanged,
    onPhaseSelectionChanged,
  } = data;
  const styles = getGraphNodeStyle(nodeType, nodeExecutionStatus);
  const [selectedNode, setSelectedNode] = useState<boolean>(false);
  const [selectedPhase, setSelectedPhase] = useState<TaskExecutionPhase | undefined>(initialPhase);

  useEffect(() => {
    if (selectedNode === true) {
      onNodeSelectionChanged(selectedNode);
      setSelectedNode(false);
      onPhaseSelectionChanged(selectedPhase);
      setSelectedPhase(selectedPhase);
    }
  }, [selectedNode, onNodeSelectionChanged, selectedPhase, onPhaseSelectionChanged]);

  const mapTaskContainerStyle: React.CSSProperties = {
    position: 'absolute',
    top: '-.82rem',
    zIndex: 0,
    right: '.15rem',
  };
  const taskNameStyle: React.CSSProperties = {
    backgroundColor: getStatusColor(nodeExecutionStatus),
    color: 'white',
    padding: '.1rem .2rem',
    fontSize: '.4rem',
    borderRadius: '.15rem',
  };

  const cacheIconStyles: React.CSSProperties = {
    width: '8px',
    height: '8px',
    marginLeft: '4px',
    marginTop: '1px',
    color: COLOR_GRAPH_BACKGROUND,
  };
  const mapTaskWrapper: React.CSSProperties = {
    display: 'flex',
  };

  const handleNodeClick = (_e) => {
    setSelectedNode(true);
    setSelectedPhase(undefined);
  };

  const renderTaskName = () => {
    return (
      <div style={mapTaskContainerStyle}>
        <div style={taskNameStyle}>{text}</div>
      </div>
    );
  };

  const renderTaskPhases = (logsByPhase) => {
    return (
      <div style={mapTaskWrapper}>
        {RENDER_ORDER.map((phase, id) => {
          if (!logsByPhase.has(phase)) {
            return null;
          }
          const defaultColor = getStatusColor();
          const phaseColor = getStatusColor(phase);
          const color = !selectedPhase || phase === selectedPhase ? phaseColor : defaultColor;
          const key = `${id}-${phase}`;
          return (
            <TaskPhaseItem
              numberOfTasks={logsByPhase.get(phase).length}
              color={color}
              phase={phase}
              setSelectedPhase={setSelectedPhase}
              setSelectedNode={setSelectedNode}
              key={key}
            />
          );
        })}
      </div>
    );
  };

  return (
    <div onClick={handleNodeClick}>
      {nodeLogsByPhase ? renderTaskName() : renderTaskType(taskType)}
      <div style={styles}>
        {nodeLogsByPhase ? renderTaskPhases(nodeLogsByPhase) : text}
        <CacheStatus cacheStatus={cacheStatus} variant="iconOnly" iconStyles={cacheIconStyles} />
      </div>
      {renderDefaultHandles(scopedId, getGraphHandleStyle('source'), getGraphHandleStyle('target'))}
    </div>
  );
};

/**
 * Custom component renders subworkflows as indepdenet flow
 * and any edge handles.
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowSubWorkflowContainer = ({ data }: RFNode) => {
  const { nodeExecutionStatus, text, scopedId, currentNestedView, onRemoveNestedView } = data;
  const BREAD_FONT_SIZE = '9px';
  const BREAD_COLOR_ACTIVE = COLOR_SPECTRUM.purple60.color;
  const BREAD_COLOR_INACTIVE = COLOR_SPECTRUM.black.color;
  const borderStyle = getNestedContainerStyle(nodeExecutionStatus);

  const handleNestedViewClick = (e) => {
    const index = e.target.id.substr(e.target.id.indexOf('_') + 1, e.target.id.length);
    onRemoveNestedView(scopedId, index);
  };

  const handleRootClick = () => {
    onRemoveNestedView(scopedId, -1);
  };

  const currentNestedDepth = currentNestedView?.length || 0;

  const BreadElement = ({ nestedView, index }) => {
    const liStyles: React.CSSProperties = {
      cursor: 'pointer',
      fontSize: BREAD_FONT_SIZE,
      color: BREAD_COLOR_ACTIVE,
    };

    const liStyleInactive: React.CSSProperties = { ...liStyles };
    liStyleInactive['color'] = BREAD_COLOR_INACTIVE;

    const beforeStyle: React.CSSProperties = {
      cursor: 'pointer',
      color: BREAD_COLOR_ACTIVE,
      padding: '0 .2rem',
      fontSize: BREAD_FONT_SIZE,
    };
    const onClick = currentNestedDepth > index + 1 ? handleNestedViewClick : undefined;
    return (
      <li
        onClick={onClick}
        style={index === currentNestedDepth - 1 ? liStyleInactive : liStyles}
        id={`${scopedId}_${index}`}
      >
        {index === 0 ? <span style={beforeStyle}>{'>'}</span> : null}
        {nestedView}
        {index < currentNestedDepth - 1 ? <span style={beforeStyle}>{'>'}</span> : null}
      </li>
    );
  };

  const BorderElement = (props) => {
    return <div style={borderStyle}>{props.children}</div>;
  };

  const BorderContainer = (props) => {
    let output = BorderElement(props);
    for (let i = 0; i < currentNestedDepth; i++) {
      output = <BorderElement>{output}</BorderElement>;
    }
    return output;
  };

  const renderBreadCrumb = () => {
    const breadContainerStyle: React.CSSProperties = {
      position: 'absolute',
      display: 'flex',
      width: '100%',
      marginTop: '-1rem',
    };
    const olStyles: React.CSSProperties = {
      margin: 0,
      padding: 0,
      display: 'flex',
      listStyle: 'none',
      listStyleImage: 'none',
      minWidth: '1rem',
    };
    const headerStyle: React.CSSProperties = {
      color: BREAD_COLOR_ACTIVE,
      fontSize: BREAD_FONT_SIZE,
      margin: 0,
      padding: 0,
    };

    const rootClick = currentNestedDepth > 0 ? handleRootClick : undefined;
    return (
      <div style={breadContainerStyle}>
        <header style={headerStyle} onClick={rootClick}>
          {text}
        </header>
        <ol style={olStyles}>
          {currentNestedView?.map((nestedView, i) => {
            return <BreadElement nestedView={nestedView} index={i} key={nestedView} />;
          })}
        </ol>
      </div>
    );
  };

  return (
    <>
      {renderBreadCrumb()}
      <BorderContainer>
        {renderDefaultHandles(
          scopedId,
          getGraphHandleStyle('source'),
          getGraphHandleStyle('target'),
        )}
      </BorderContainer>
    </>
  );
};

/**
 * Custom component renders start node
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomStartNode = ({ data }: RFNode) => {
  const { text, nodeType, scopedId } = data;
  const styles = getGraphNodeStyle(nodeType);
  return (
    <>
      <div style={styles}>{text}</div>
      {renderStardEndHandles(nodeType, scopedId)}
    </>
  );
};

/**
 * Custom component renders start node
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomEndNode = ({ data }: RFNode) => {
  const { text, nodeType, scopedId } = data;
  const styles = getGraphNodeStyle(nodeType);
  return (
    <>
      <div style={styles}>{text}</div>
      {renderStardEndHandles(nodeType, scopedId)}
    </>
  );
};
