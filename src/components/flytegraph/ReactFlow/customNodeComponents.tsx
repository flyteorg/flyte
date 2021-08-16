import * as React from 'react';
import { useState, useEffect } from 'react';
import {
    Handle,
    getBezierPath,
    getMarkerEnd,
    Position
} from 'react-flow-renderer';
import { dTypes } from 'models/Graph/types';
import { ReactFlowWrapper } from './ReactFlowWrapper';
import {
    COLOR_TASK_TYPE,
    COLOR_GRAPH_BACKGROUND,
    getGraphHandleStyle,
    getGraphNodeStyle,
    getRFBackground
} from './utils';
import { RFGraphTypes, RFHandleProps } from './types';

export const customEdge = (
    id,
    sourceX,
    sourceY,
    targetX,
    targetY,
    sourcePosition,
    targetPosition,
    style = {},
    data,
    arrowHeadType,
    markerEndId
) => {
    const edgePath = getBezierPath({
        sourceX,
        sourceY,
        sourcePosition,
        targetX,
        targetY,
        targetPosition
    });
    const markerEnd = getMarkerEnd(arrowHeadType, markerEndId);

    return (
        <>
            <path
                id={id}
                style={style}
                className="react-flow__edge-path"
                d={edgePath}
                markerEnd={markerEnd}
            />
        </>
    );
};

export const renderDefaultHandles = (
    id: string,
    sourceStyle: any,
    targetStyle: any
) => {
    const leftHandleProps: RFHandleProps = {
        id: `rf-handle-left-${id}`,
        type: 'target',
        position: Position.Left,
        style: targetStyle
    };

    const rightHandleProps: RFHandleProps = {
        id: `rf-handle-right-${id}`,
        type: 'source',
        position: Position.Right,
        style: sourceStyle
    };
    return (
        <>
            <Handle {...leftHandleProps} />
            <Handle {...rightHandleProps} />
        </>
    );
};

export const renderStardEndHandles = (data: any) => {
    const isStart =
        data.nodeType == dTypes.nestedStart || data.nodeType == dTypes.start;
    const idPrefix = isStart ? 'start' : 'end';
    const position = isStart ? Position.Right : Position.Left;
    const type = isStart ? 'source' : 'target';

    /**
     * @TODO
     * Resepect the actual node type once toggle nested graphs is implemented
     * For now we force nestedMaxDepth for any nested types
     */
    const style =
        data.nodeType == dTypes.nestedStart || data.nodeType == dTypes.nestedEnd
            ? getGraphHandleStyle(type, dTypes.nestedMaxDepth)
            : getGraphHandleStyle(type);

    const handleProps: RFHandleProps = {
        id: `rf-handle-${idPrefix}-${data.scopedId}`,
        type: type,
        position: position,
        style: style
    };

    return (
        <>
            <Handle {...handleProps} />
        </>
    );
};

/**
 * Styles start/end nodes as a point; used for nested workflows
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomNestedPoint = ({ data }: any) => {
    const containerStyle = getGraphNodeStyle(data.nodeType);
    return (
        <>
            <div style={containerStyle} />
            {renderStardEndHandles(data)}
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

export const ReactFlowCustomMaxNested = ({ data }: any) => {
    const styles = getGraphNodeStyle(dTypes.nestedMaxDepth);
    const containerStyle = {};
    const taskContainerStyle: React.CSSProperties = {
        position: 'absolute',
        top: '-.55rem',
        zIndex: 0,
        right: '.15rem'
    };
    const taskTypeStyle: React.CSSProperties = {
        backgroundColor: COLOR_GRAPH_BACKGROUND,
        color: 'white',
        padding: '.1rem .2rem',
        fontSize: '.3rem'
    };

    const renderTaskType = () => {
        return (
            <div style={taskContainerStyle}>
                <div style={taskTypeStyle}>{data.taskType}</div>
            </div>
        );
    };

    return (
        <div style={containerStyle}>
            {data.taskType ? renderTaskType() : null}
            <div style={styles}>{data.text}</div>
            {renderDefaultHandles(
                data.scopedId,
                getGraphHandleStyle('source'),
                getGraphHandleStyle('target')
            )}
        </div>
    );
};

/**
 * Custom component used by ReactFlow.  Renders a label (text)
 * and any edge handles.
 * @param props.data data property of ReactFlowGraphNodeData
 */

export const ReactFlowCustomTaskNode = ({ data }: any) => {
    const styles = getGraphNodeStyle(data.nodeType, data.nodeExecutionStatus);
    const onNodeSelectionChanged = data.onNodeSelectionChanged;
    const [selectedNode, setSelectedNode] = useState(false);

    useEffect(() => {
        if (selectedNode == true) {
            onNodeSelectionChanged(selectedNode);
            setSelectedNode(false);
        }
    }, [selectedNode, onNodeSelectionChanged]);

    const taskContainerStyle: React.CSSProperties = {
        position: 'absolute',
        top: '-.55rem',
        zIndex: 0,
        right: '.15rem'
    };
    const taskTypeStyle: React.CSSProperties = {
        backgroundColor: COLOR_TASK_TYPE,
        color: 'white',
        padding: '.1rem .2rem',
        fontSize: '.3rem'
    };

    const handleClick = e => {
        setSelectedNode(true);
    };

    const renderTaskType = () => {
        return (
            <div style={taskContainerStyle}>
                <div style={taskTypeStyle}>{data.taskType}</div>
            </div>
        );
    };

    /**
     * @TODO
     * Decide if we want to make all nodes clickable
     */
    // const isClickable =
    //     data.nodeExecutionStatus == RFNodeExecutionStatus.executed;

    return (
        // <div
        //     {...(isClickable && {
        //         onClick: handleClick
        //     })}
        // >
        <div onClick={handleClick}>
            {data.taskType ? renderTaskType() : null}
            <div style={styles}>{data.text}</div>
            {renderDefaultHandles(
                data.scopedId,
                getGraphHandleStyle('source'),
                getGraphHandleStyle('target')
            )}
        </div>
    );
};

/**
 * Custom component renders subworkflows as indepdenet flow
 * and any edge handles.
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomSubworkflowNode = ({ data }: any) => {
    const { dag } = data;
    const backgroundStyle = getRFBackground(data.nodeExecutionStatus).nested;

    const rfContainerStyle: React.CSSProperties = {
        width: `300px`,
        height: `200px`
    };

    return (
        <>
            {renderDefaultHandles(
                data.scopedId,
                getGraphHandleStyle('source'),
                getGraphHandleStyle('target')
            )}
            <div style={rfContainerStyle}>
                <ReactFlowWrapper
                    rfGraphJson={dag}
                    backgroundStyle={backgroundStyle}
                    type={RFGraphTypes.nested}
                />
            </div>
        </>
    );
};

/**
 * Custom component renders Branch nodes as indepdenet flow
 * and any edge handles.
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomBranchNode = ({ data }: any) => {
    const { dag } = data;
    const backgroundStyle = getRFBackground(data.nodeExecutionStatus).nested;

    const rfContainerStyle: React.CSSProperties = {
        width: `300px`,
        height: `200px`
    };

    return (
        <>
            {renderDefaultHandles(
                data.scopedId,
                getGraphHandleStyle('source'),
                getGraphHandleStyle('target')
            )}
            <div style={rfContainerStyle}>
                <ReactFlowWrapper
                    rfGraphJson={dag}
                    backgroundStyle={backgroundStyle}
                    type={RFGraphTypes.nested}
                />
            </div>
        </>
    );
};

/**
 * Custom component renders start node
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomStartNode = ({ data }: any) => {
    const styles = getGraphNodeStyle(data.nodeType);
    return (
        <>
            <div style={styles}>{data.text}</div>
            {renderStardEndHandles(data)}
        </>
    );
};

/**
 * Custom component renders start node
 * @param props.data data property of ReactFlowGraphNodeData
 */
export const ReactFlowCustomEndNode = ({ data }: any) => {
    const styles = getGraphNodeStyle(data.nodeType);
    return (
        <>
            <div style={styles}>{data.text}</div>
            {renderStardEndHandles(data)}
        </>
    );
};
