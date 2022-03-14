import * as React from 'react';
import { useState, useEffect } from 'react';
import { Handle, Position } from 'react-flow-renderer';
import { dTypes } from 'models/Graph/types';
import {
    COLOR_TASK_TYPE,
    COLOR_GRAPH_BACKGROUND,
    getGraphHandleStyle,
    getGraphNodeStyle,
    getNestedContainerStyle
} from './utils';
import { RFHandleProps } from './types';

export const renderDefaultHandles = (
    id: string,
    sourceStyle: any,
    targetStyle: any
) => {
    const leftHandleProps: RFHandleProps = {
        id: `edge-left-${id}`,
        type: 'target',
        position: Position.Left,
        style: targetStyle
    };

    const rightHandleProps: RFHandleProps = {
        id: `edge-right-${id}`,
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

    return <Handle {...handleProps} />;
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

    const onClick = e => {
        data.onAddNestedView();
    };

    return (
        <div style={containerStyle} onClick={onClick}>
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

export const ReactFlowStaticNested = ({ data }: any) => {
    const styles = getGraphNodeStyle(dTypes.staticNestedNode);
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

export const ReactFlowStaticNode = ({ data }: any) => {
    const styles = getGraphNodeStyle(dTypes.staticNode);
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

    const handleClick = _e => {
        setSelectedNode(true);
    };

    const renderTaskType = () => {
        return (
            <div style={taskContainerStyle}>
                <div style={taskTypeStyle}>{data.taskType}</div>
            </div>
        );
    };
    return (
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
export const ReactFlowSubWorkflowContainer = ({ data }: any) => {
    const BREAD_FONT_SIZE = '9px';
    const BREAD_COLOR_ACTIVE = '#8B37FF';
    const BREAD_COLOR_INACTIVE = '#000';
    const borderStyle = getNestedContainerStyle(data.nodeExecutionStatus);

    const handleNestedViewClick = e => {
        const index = e.target.id.substr(
            e.target.id.indexOf('_') + 1,
            e.target.id.length
        );
        data.onRemoveNestedView(data.scopedId, index);
    };

    const handleRootClick = () => {
        data.onRemoveNestedView(data.scopedId, -1);
    };

    const currentNestedDepth = data.currentNestedView?.length || 0;

    const BreadElement = ({ nestedView, index }) => {
        const liStyles: React.CSSProperties = {
            cursor: 'pointer',
            fontSize: BREAD_FONT_SIZE,
            color: BREAD_COLOR_ACTIVE
        };

        const liStyleInactive: React.CSSProperties = { ...liStyles };
        liStyleInactive['color'] = BREAD_COLOR_INACTIVE;

        const beforeStyle: React.CSSProperties = {
            cursor: 'pointer',
            color: BREAD_COLOR_ACTIVE,
            padding: '0 .2rem',
            fontSize: BREAD_FONT_SIZE
        };
        const onClick =
            currentNestedDepth > index + 1 ? handleNestedViewClick : undefined;
        return (
            <li
                onClick={onClick}
                style={
                    index == currentNestedDepth - 1 ? liStyleInactive : liStyles
                }
                id={`${data.scopedId}_${index}`}
            >
                {index == 0 ? <span style={beforeStyle}>{'>'}</span> : null}
                {nestedView}
                {index < currentNestedDepth - 1 ? (
                    <span style={beforeStyle}>{'>'}</span>
                ) : null}
            </li>
        );
    };

    const BorderElement = props => {
        return <div style={borderStyle}>{props.children}</div>;
    };

    const BorderContainer = props => {
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
            marginTop: '-1rem'
        };
        const olStyles: React.CSSProperties = {
            margin: 0,
            padding: 0,
            display: 'flex',
            listStyle: 'none',
            listStyleImage: 'none',
            minWidth: '1rem'
        };
        const headerStyle: React.CSSProperties = {
            color: BREAD_COLOR_ACTIVE,
            fontSize: BREAD_FONT_SIZE,
            margin: 0,
            padding: 0
        };

        const rootClick = currentNestedDepth > 0 ? handleRootClick : undefined;
        return (
            <div style={breadContainerStyle}>
                <header style={headerStyle} onClick={rootClick}>
                    {data.text}
                </header>
                <ol style={olStyles}>
                    {data.currentNestedView?.map((nestedView, i) => {
                        return (
                            <BreadElement
                                nestedView={nestedView}
                                index={i}
                                key={nestedView}
                            />
                        );
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
                    data.scopedId,
                    getGraphHandleStyle('source'),
                    getGraphHandleStyle('target')
                )}
            </BorderContainer>
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
