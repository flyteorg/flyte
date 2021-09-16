import { NodeExecutionPhase } from 'models/Execution/enums';
import { dTypes } from 'models/Graph/types';
import { CSSProperties } from 'react';
import { Elements, isNode, Position } from 'react-flow-renderer';
import { RFBackgroundProps } from './types';

const dagre = require('dagre');

export const COLOR_EXECUTED = '#2892f4';
export const COLOR_NOT_EXECUTED = '#c6c6c6';
export const COLOR_TASK_TYPE = '#666666';
export const COLOR_GRAPH_BACKGROUND = '#666666';

export const DISPLAY_NAME_START = 'start';
export const DISPLAY_NAME_END = 'end';
export const MAX_NESTED_DEPTH = 1;
export const HANDLE_ICON = require('assets/SmallArrow.svg') as string;

export const ReactFlowGraphConfig = {
    customNodePrefix: 'FlyteNode',
    arrowHeadType: 'arrowClosed',
    edgeType: 'default'
};

export const getGraphHandleStyle = (
    handleType: string,
    type?: dTypes
): CSSProperties => {
    let size = 8;
    const offset = 10;

    let backgroundColor = `rgba(255,255,255,1)`;
    let marginLeft,
        marginRight = 0;

    if (handleType == 'target') {
        marginLeft = 0;
        marginRight = -offset;
    } else if (handleType == 'source') {
        marginRight = 0;
        marginLeft = -offset;
    } else if (handleType == 'nestedPoint') {
        backgroundColor = 'none';
        size = 1;
    }

    const baseStyle = {
        zIndex: 99999999,
        marginLeft: `${marginLeft}px`,
        marginRight: `${marginRight}px`,
        width: `${size}px`,
        height: `${size}px`,
        background: backgroundColor,
        backgroundImage: `url(${HANDLE_ICON})`,
        backgroundRepeat: 'no-repeat',
        backgroundPosition: 'center center',
        border: 'none'
    };

    /**
     * @TODO Keeping this for future
     * */
    const overrideStyles = {
        nestedMaxDepth: {
            background: 'none',
            backgroundImage: 'none'
        }
    };

    if (type) {
        const key = String(dTypes[type]);
        const output = {
            ...baseStyle,
            ...overrideStyles[key]
        };
        return output;
    } else {
        return baseStyle;
    }
};

export const nodePhaseColorMapping = {
    [NodeExecutionPhase.FAILED]: { color: '#e90000', text: 'Failed' },
    [NodeExecutionPhase.FAILING]: { color: '#f2a4ad', text: 'Failing' },
    [NodeExecutionPhase.SUCCEEDED]: { color: '#37b789', text: 'Succeded' },
    [NodeExecutionPhase.ABORTED]: { color: '#be25d7', text: 'Aborted' },
    [NodeExecutionPhase.RUNNING]: { color: '#2892f4', text: 'Running' },
    [NodeExecutionPhase.QUEUED]: { color: '#dfd71b', text: 'Queued' },
    [NodeExecutionPhase.UNDEFINED]: { color: '#4a2839', text: 'Undefined' }
};

/**
 * Maps node execution phases to UX colors
 * @param nodeExecutionStatus
 * @returns
 */
export const getStatusColor = (
    nodeExecutionStatus: NodeExecutionPhase
): string => {
    if (nodePhaseColorMapping[nodeExecutionStatus]) {
        return nodePhaseColorMapping[nodeExecutionStatus].color;
    } else {
        /** @TODO decide what we want default color to be */
        return '#c6c6c6';
    }
};

export const getNestedGraphContainerStyle = overwrite => {
    let width = overwrite.width;
    let height = overwrite.height;

    const maxHeight = 500;
    const minHeight = 200;
    const maxWidth = 700;
    const minWidth = 300;

    if (overwrite) {
        width = width > maxWidth ? maxWidth : width;
        width = width < minWidth ? minWidth : width;
        height = height > maxHeight ? maxHeight : height;
        height = height < minHeight ? minHeight : height;
    }

    const output: React.CSSProperties = {
        width: `${width}px`,
        height: `${height}px`
    };

    return output;
};

export const getNestedContainerStyle = nodeExecutionStatus => {
    const style = {
        border: `1px dashed ${getStatusColor(nodeExecutionStatus)}`,
        borderRadius: '8px',
        background: 'rgba(255,255,255,.9)'
    } as React.CSSProperties;
    return style;
};

export const getGraphNodeStyle = (
    type: dTypes,
    nodeExecutionStatus?: NodeExecutionPhase
): CSSProperties => {
    /** Base styles for displaying graph nodes */
    const baseStyle = {
        boxShadow: '1px 3px 5px rgba(0,0,0,.2)',
        padding: '.25rem .75rem',
        fontSize: '.6rem',
        color: '#323232',
        borderRadius: '.25rem',
        border: '.15rem solid #555',
        background: '#fff',
        minWidth: '.5rem',
        minHeight: '.5rem',
        height: 'auto',
        width: 'auto'
    };

    const nestedPoint = {
        width: '1px',
        height: '1px',
        minWidth: '1px',
        minHeight: '1px',
        padding: 0,
        boxShadow: 'none',
        border: 'none',
        background: 'none',
        borderRadius: 'none',
        color: '#fff'
    };

    let nodePrimaryColor = '';
    if (nodeExecutionStatus) {
        nodePrimaryColor = getStatusColor(nodeExecutionStatus);
    }

    /** Override the base styles with node-type specific styles */
    const overrideStyles = {
        start: {
            border: '1px solid #ddd'
        },
        end: {
            border: '1px solid #ddd'
        },
        nestedStart: {
            ...nestedPoint
        },
        nestedEnd: {
            ...nestedPoint
        },
        nestedWithChildren: {
            borderColor: nodePrimaryColor
        },
        nestedMaxDepth: {
            background: '#aaa',
            color: 'white',
            border: 'none'
        },
        branch: {
            display: 'flex',
            flexAlign: 'center',
            border: 'none',
            borderRadius: '0px',
            padding: '1rem 0',
            boxShadow: 'none',
            fontSize: '.6rem'
        },
        workflow: {
            borderColor: nodePrimaryColor
        },
        task: {
            borderColor: nodePrimaryColor
        }
    };
    const key = String(dTypes[type]);
    const output = {
        ...baseStyle,
        ...overrideStyles[key]
    };
    return output;
};

export const getRFBackground = () => {
    return {
        main: {
            background: {
                border: '1px solid #444',
                backgroundColor: 'rgba(255,255,255,1)'
            },
            gridColor: '#ccc',
            gridSpacing: 20
        } as RFBackgroundProps,
        nested: {
            gridColor: 'none',
            gridSpacing: 1
        } as RFBackgroundProps
    };
};

/**
 * Uses dagree/graphlib to compute graph layout
 * @see https://github.com/dagrejs/dagre/wiki
 * @param elements      Graph elements (nodes/edges) in JSON format
 * @param direction     Direction to render graph
 * @returns
 */
export const setReactFlowGraphLayout = (
    elements: Elements,
    direction: string,
    estimate = false
) => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));
    const isHorizontal = direction === 'LR';

    const ESTIMATE_HEIGHT = 25;
    const ESTIMATE_WIDTH_FACTOR = 6;

    dagreGraph.setGraph({
        rankdir: direction,
        edgesep: 60,
        nodesep: 30,
        ranker: 'longest-path',
        acyclicer: 'greedy'
    });

    /**
     * Note: this waits/assumes rendered dimensions from ReactFlow as .__rf
     */
    elements.forEach(el => {
        if (isNode(el)) {
            const nodeWidth = estimate
                ? el.data.text.length * ESTIMATE_WIDTH_FACTOR
                : el.__rf.width;
            const nodeHeight = estimate ? ESTIMATE_HEIGHT : el.__rf.height;
            dagreGraph.setNode(el.id, { width: nodeWidth, height: nodeHeight });
        } else {
            dagreGraph.setEdge(el.source, el.target);
        }
    });

    dagre.layout(dagreGraph);
    const graphWidth = dagreGraph.graph().width;
    const graphHeight = dagreGraph.graph().height;
    if (estimate) {
        return {
            estimatedDimensions: {
                width: graphWidth,
                height: graphHeight
            }
        };
    } else {
        return {
            graph: elements.map(el => {
                if (isNode(el)) {
                    el.targetPosition = isHorizontal
                        ? Position.Left
                        : Position.Top;
                    el.sourcePosition = isHorizontal
                        ? Position.Right
                        : Position.Bottom;
                    const nodeWidth = estimate
                        ? el.data.text.length * ESTIMATE_WIDTH_FACTOR
                        : el.__rf.width;
                    const nodeHeight = estimate
                        ? ESTIMATE_HEIGHT
                        : el.__rf.height;
                    const nodeWithPosition = dagreGraph.node(el.id);

                    /** Keep both position and .__rf.position in sync */
                    const x = nodeWithPosition.x - nodeWidth / 2;
                    const y = nodeWithPosition.y - nodeHeight / 2;
                    el.position = {
                        x: x,
                        y: y
                    };
                    el.__rf.position = {
                        x: x,
                        y: y
                    };
                }
                return el;
            }),
            dimensions: {
                width: graphWidth,
                height: graphHeight
            }
        };
    }
};

export default setReactFlowGraphLayout;
