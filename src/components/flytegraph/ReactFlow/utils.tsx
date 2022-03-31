import * as React from 'react';
import { NodeExecutionPhase } from 'models/Execution/enums';
import { dTypes } from 'models/Graph/types';
import { CSSProperties } from 'react';
import { RFBackgroundProps } from './types';

const dagre = require('dagre');

export const COLOR_EXECUTED = '#2892f4';
export const COLOR_NOT_EXECUTED = '#c6c6c6';
export const COLOR_TASK_TYPE = '#666666';
export const COLOR_GRAPH_BACKGROUND = '#666666';
export const GRAPH_PADDING_FACTOR = 50;

export const DISPLAY_NAME_START = 'start';
export const DISPLAY_NAME_END = 'end';
export const MAX_NESTED_DEPTH = 1;
export const HANDLE_ICON = require('assets/SmallArrow.svg') as string;

export const ReactFlowGraphConfig = {
  customNodePrefix: 'FlyteNode',
  arrowHeadType: 'arrowClosed',
  edgeType: 'default',
};

/**
 * Function replaces all retry values with '0' to be used a key between static/runtime
 * values.
 * @param id   NodeExcution nodeId.
 * @returns    nodeId with all retry values replaces with '0'
 */
export const retriesToZero = (id: string): string => {
  const output = id.replace(/(-[0-9]-)/g, '-0-');
  return output;
};

export const getGraphHandleStyle = (handleType: string, type?: dTypes): CSSProperties => {
  let size = 8;
  const offset = 10;

  let backgroundColor = `rgba(255,255,255,1)`;
  let marginLeft,
    marginRight = 0;

  if (handleType === 'target') {
    marginLeft = 0;
    marginRight = -offset;
  } else if (handleType === 'source') {
    marginRight = 0;
    marginLeft = -offset;
  } else if (handleType === 'nestedPoint') {
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
    border: 'none',
  };

  /**
   * @TODO Keeping this for future
   * */
  const overrideStyles = {
    nestedMaxDepth: {
      background: 'none',
      backgroundImage: 'none',
    },
  };

  if (type) {
    const key = String(dTypes[type]);
    const output = {
      ...baseStyle,
      ...overrideStyles[key],
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
  [NodeExecutionPhase.UNDEFINED]: { color: '#4a2839', text: 'Undefined' },
};

/**
 * Maps node execution phases to UX colors
 * @param nodeExecutionStatus
 * @returns
 */
export const getStatusColor = (nodeExecutionStatus: NodeExecutionPhase): string => {
  if (nodePhaseColorMapping[nodeExecutionStatus]) {
    return nodePhaseColorMapping[nodeExecutionStatus].color;
  } else {
    /** @TODO decide what we want default color to be */
    return '#c6c6c6';
  }
};

export const getNestedGraphContainerStyle = (overwrite) => {
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
    height: `${height}px`,
  };

  return output;
};

export const getNestedContainerStyle = (nodeExecutionStatus) => {
  const style = {
    border: `1px dashed ${getStatusColor(nodeExecutionStatus)}`,
    borderRadius: '8px',
    background: 'rgba(255,255,255,.9)',
    width: '100%',
    height: '100%',
    padding: '.25rem',
  } as React.CSSProperties;
  return style;
};

export const getGraphNodeStyle = (
  type: dTypes,
  nodeExecutionStatus?: NodeExecutionPhase,
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
    width: 'auto',
    zIndex: 100000,
    position: 'relative',
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
    color: '#fff',
  };

  let nodePrimaryColor = '';
  if (nodeExecutionStatus) {
    nodePrimaryColor = getStatusColor(nodeExecutionStatus);
  }

  /** Override the base styles with node-type specific styles */
  const overrideStyles = {
    start: {
      border: '1px solid #ddd',
    },
    end: {
      border: '1px solid #ddd',
    },
    nestedStart: {
      ...nestedPoint,
    },
    nestedEnd: {
      ...nestedPoint,
    },
    nestedWithChildren: {
      borderColor: nodePrimaryColor,
    },
    nestedMaxDepth: {
      background: '#aaa',
      color: 'white',
      border: 'none',
    },
    branch: {
      display: 'flex',
      flexAlign: 'center',
      border: 'none',
      borderRadius: '0px',
      padding: '1rem 0',
      boxShadow: 'none',
      fontSize: '.6rem',
    },
    workflow: {
      borderColor: nodePrimaryColor,
    },
    task: {
      borderColor: nodePrimaryColor,
    },
    staticNode: {
      backgroundColor: '#fff',
      borderColor: '#bfbfbf',
      borderWidth: '.05rem',
    },
    staticNestedNode: {
      backgroundColor: '#dfdfdf',
      border: 'none',
    },
  };
  const key = String(dTypes[type]);
  const output = {
    ...baseStyle,
    ...overrideStyles[key],
  };
  return output;
};

export const getRFBackground = () => {
  return {
    main: {
      background: {
        border: '1px solid #444',
        backgroundColor: 'rgba(255,255,255,1)',
      },
      gridColor: 'none',
    } as RFBackgroundProps,
    nested: {
      gridColor: 'none',
    } as RFBackgroundProps,
    static: {
      background: {
        border: 'none',
        backgroundColor: 'rgb(255,255,255)',
      },
      gridColor: 'none',
    } as RFBackgroundProps,
  };
};

export interface PositionProps {
  nodes: any;
  edges: any;
  parentMap?: any;
  direction?: string;
}

/**
 * Computes positions for provided nodes
 * @param PositionProps
 * @returns
 */
export const computeChildNodePositions = ({ nodes, edges, direction = 'LR' }: PositionProps) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  dagreGraph.setGraph({
    rankdir: direction,
    edgesep: 60,
    nodesep: 30,
    ranker: 'longest-path',
    acyclicer: 'greedy',
  });
  nodes.map((n) => {
    dagreGraph.setNode(n.id, n.dimensions);
  });
  edges.map((e) => {
    dagreGraph.setEdge(e.source, e.target);
  });
  dagre.layout(dagreGraph);
  const dimensions = {
    width: dagreGraph.graph().width,
    height: dagreGraph.graph().height,
  };
  const graph = nodes.map((el) => {
    const node = dagreGraph.node(el.id);
    const x = node.x - node.width / 2;
    const y = node.y - node.height / 2;
    return {
      ...el,
      position: {
        x: x,
        y: y,
      },
    };
  });
  return { graph, dimensions };
};

/**
 * Computes positions for root-level nodes in a graph by filtering out
 * all children (nodes that have parents).
 * @param PositionProps
 * @returns
 */
export const computeRootNodePositions = ({
  nodes,
  edges,
  parentMap,
  direction = 'LR',
}: PositionProps) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  dagreGraph.setGraph({
    rankdir: direction,
    edgesep: 60,
    nodesep: 30,
    ranker: 'longest-path',
    acyclicer: 'greedy',
  });

  /* Filter all children from creating dagree nodes */
  nodes.map((n) => {
    if (n.isRootParentNode) {
      dagreGraph.setNode(n.id, {
        width: parentMap[n.id].childGraphDimensions.width,
        height: parentMap[n.id].childGraphDimensions.height,
      });
    } else if (!n.parentNode) {
      dagreGraph.setNode(n.id, n.dimensions);
    }
  });

  /* Filter all children from creating dagree edges */
  edges.map((e) => {
    if (!e.parent) {
      dagreGraph.setEdge(e.source, e.target);
    }
  });

  /* Compute graph posistions for root-level nodes */
  dagre.layout(dagreGraph);
  const dimensions = {
    width: dagreGraph.graph().width,
    height: dagreGraph.graph().height,
  };

  /* Merge dagre positions to rf elements */
  const graph = nodes.map((el) => {
    const node = dagreGraph.node(el.id);
    if (node) {
      const x = node.x - node.width / 2;
      const y = node.y - node.height / 2;
      if (parentMap && el.isRootParentNode) {
        el.style = parentMap[el.id].childGraphDimensions;
      }
      return {
        ...el,
        position: {
          x: x,
          y: y,
        },
      };
    } else if (parentMap) {
      /* Case: Overwrite children positions with computed values */
      const parent = parentMap[el.parentNode];
      for (let i = 0; i < parent.nodes.length; i++) {
        const node = parent.nodes[i];
        if (node.id == el.id) {
          return {
            ...el,
            position: { ...node.position },
          };
        }
      }
    }
  });
  return { graph, dimensions };
};

/**
 * Returns positioned nodes and edges.
 *
 * Note: Handles nesting by first "rendering" all child graphs to calculate their rendered
 * dimensions and setting those values as the dimentions (width/height) for parent/container.
 * Once those dimensions have been set (for parent/container nodes) we can set root-level node
 * positions.
 *
 * @param nodes
 * @param edges
 * @param currentNestedView
 * @param direction
 * @returns Array of ReactFlow nodes
 */
export const getPositionedNodes = (nodes, edges, currentNestedView, direction = 'LR') => {
  const parentMap = {};
  /* (1) Collect all child graphs in parentMap */
  nodes.forEach((node) => {
    if (node.isRootParentNode) {
      parentMap[node.id] = {
        nodes: [],
        edges: [],
        childGraphDimensions: {
          width: 0,
          height: 0,
        },
        self: node,
      };
    }
    if (node.parentNode) {
      if (parentMap[node.parentNode]) {
        if (parentMap[node.parentNode].nodes) {
          parentMap[node.parentNode].nodes.push(node);
        } else {
          parentMap[node.parentNode].nodes = [node];
        }
      }
    }
  });
  edges.forEach((edge) => {
    if (edge.parent) {
      if (parentMap[edge.parent]) {
        if (parentMap[edge.parent].edges) {
          parentMap[edge.parent].edges.push(edge);
        } else {
          parentMap[edge.parent].edges = [edge];
        }
      }
    }
  });

  /* (2) Compute child graph positiions/dimensions */
  for (const parentId in parentMap) {
    const children = parentMap[parentId];
    const childGraph = computeChildNodePositions({
      nodes: children.nodes,
      edges: children.edges,
      direction: direction,
    });
    let nestedDepth = 1;
    if (
      currentNestedView &&
      currentNestedView[parentId] &&
      currentNestedView[parentId].length > 0
    ) {
      nestedDepth = currentNestedView[parentId].length;
    }
    const borderPadding = GRAPH_PADDING_FACTOR * nestedDepth;
    const width = childGraph.dimensions.width + borderPadding;
    const height = childGraph.dimensions.height + borderPadding;

    parentMap[parentId].childGraphDimensions = {
      width: width,
      height: height,
    };
    const relativePosNodes = childGraph.graph.map((node) => {
      const position = node.position;
      position.y = position.y + GRAPH_PADDING_FACTOR / 2;
      position.x = position.x + GRAPH_PADDING_FACTOR / 2;
      return {
        ...node,
        position,
      };
    });
    parentMap[parentId].nodes = relativePosNodes;
    parentMap[parentId].self.dimensions.width = width;
    parentMap[parentId].self.dimensions.height = height;
  }
  /* (3) Compute positions of root-level nodes */
  const { graph, dimensions } = computeRootNodePositions({
    nodes: nodes,
    edges: edges,
    direction: direction,
    parentMap: parentMap,
  });
  return graph;
};

export const ReactFlowIdHash = (nodes, edges) => {
  const key = Math.floor(Math.random() * 10000).toString();
  const properties = ['id', 'source', 'target', 'parent', 'parentNode'];
  const hashGraph = nodes.map((node) => {
    const updates = {};
    properties.forEach((prop) => {
      if (node[prop]) {
        updates[prop] = `${key}-${node[prop]}`;
      }
    });
    return { ...node, ...updates };
  });

  const hashEdges = edges.map((edge) => {
    const updates = {};
    properties.forEach((prop) => {
      if (edge[prop]) {
        updates[prop] = `${key}-${edge[prop]}`;
      }
    });
    return { ...edge, ...updates };
  });
  return { hashGraph, hashEdges };
};
