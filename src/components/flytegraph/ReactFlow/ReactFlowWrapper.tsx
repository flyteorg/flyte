import * as React from 'react';
import { useState, useEffect, useCallback } from 'react';
import ReactFlow, { Background } from 'react-flow-renderer';
import { createDebugLogger } from 'common/log';
import { RFWrapperProps } from './types';
import {
  ReactFlowCustomEndNode,
  ReactFlowCustomNestedPoint,
  ReactFlowCustomStartNode,
  ReactFlowCustomTaskNode,
  ReactFlowSubWorkflowContainer,
  ReactFlowCustomMaxNested,
  ReactFlowStaticNested,
  ReactFlowStaticNode,
} from './customNodeComponents';
import { getPositionedNodes, ReactFlowIdHash } from './utils';

const debug = createDebugLogger('@ReactFlowWrapper');

/**
 * Mapping for using custom nodes inside ReactFlow
 */
const CustomNodeTypes = {
  FlyteNode_task: ReactFlowCustomTaskNode,
  FlyteNode_subworkflow: ReactFlowSubWorkflowContainer,
  FlyteNode_start: ReactFlowCustomStartNode,
  FlyteNode_end: ReactFlowCustomEndNode,
  FlyteNode_nestedStart: ReactFlowCustomNestedPoint,
  FlyteNode_nestedEnd: ReactFlowCustomNestedPoint,
  FlyteNode_nestedMaxDepth: ReactFlowCustomMaxNested,
  FlyteNode_staticNode: ReactFlowStaticNode,
  FlyteNode_staticNestedNode: ReactFlowStaticNested,
};

export const ReactFlowWrapper: React.FC<RFWrapperProps> = ({
  rfGraphJson,
  backgroundStyle,
  currentNestedView,
  version,
}) => {
  const [state, setState] = useState({
    shouldUpdate: true,
    nodes: rfGraphJson.nodes,
    edges: rfGraphJson.edges,
    version: version,
  });

  const [reactFlowInstance, setReactFlowInstance] = useState<null | any>(null);

  useEffect(() => {
    if (reactFlowInstance && state.shouldUpdate == false) {
      reactFlowInstance?.fitView();
    }
  }, [state.shouldUpdate, reactFlowInstance]);

  useEffect(() => {
    setState((state) => ({
      ...state,
      shouldUpdate: true,
      nodes: rfGraphJson.nodes,
      edges: rfGraphJson.edges,
    }));
  }, [rfGraphJson]);

  const onLoad = (rf: any) => {
    setReactFlowInstance(rf);
  };

  const onNodesChange = useCallback(
    (changes) => {
      if (changes.length == state.nodes.length && state.shouldUpdate) {
        const nodesWithDimensions: any[] = [];
        for (let i = 0; i < changes.length; i++) {
          nodesWithDimensions.push({
            ...state.nodes[i],
            ['dimensions']: changes[i].dimensions,
          });
        }
        const positionedNodes = getPositionedNodes(
          nodesWithDimensions,
          state.edges,
          currentNestedView,
          'LR',
        );
        const { hashGraph, hashEdges } = ReactFlowIdHash(positionedNodes, state.edges);

        setState((state) => ({
          ...state,
          shouldUpdate: false,
          nodes: hashGraph,
          edges: hashEdges,
        }));
      }
    },
    [state.shouldUpdate],
  );

  const reactFlowStyle: React.CSSProperties = {
    display: 'flex',
    flex: `1 1 100%`,
    flexDirection: 'column',
  };

  return (
    <ReactFlow
      nodes={state.nodes}
      edges={state.edges}
      nodeTypes={CustomNodeTypes}
      onNodesChange={onNodesChange}
      style={reactFlowStyle}
      onPaneReady={onLoad}
      fitViewOnInit
    >
      <Background
        style={backgroundStyle.background}
        color={backgroundStyle.gridColor}
        gap={backgroundStyle.gridSpacing}
      />
    </ReactFlow>
  );
};
