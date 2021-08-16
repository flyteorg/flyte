import * as React from 'react';
import { useState, useEffect } from 'react';
import ReactFlow, {
    Background,
    ReactFlowProvider,
    useStoreState
} from 'react-flow-renderer';
import { RFWrapperProps, LayoutRCProps } from './types';
import {
    ReactFlowCustomBranchNode,
    ReactFlowCustomEndNode,
    ReactFlowCustomNestedPoint,
    ReactFlowCustomStartNode,
    ReactFlowCustomTaskNode,
    ReactFlowCustomSubworkflowNode,
    ReactFlowCustomMaxNested
} from './customNodeComponents';
import setReactFlowGraphLayout from './utils';

/**
 * Mapping for using custom nodes inside ReactFlow
 */
const CustomNodeTypes = {
    FlyteNode_task: ReactFlowCustomTaskNode,
    FlyteNode_subworkflow: ReactFlowCustomSubworkflowNode,
    FlyteNode_branch: ReactFlowCustomBranchNode,
    FlyteNode_start: ReactFlowCustomStartNode,
    FlyteNode_end: ReactFlowCustomEndNode,
    FlyteNode_nestedStart: ReactFlowCustomNestedPoint,
    FlyteNode_nestedEnd: ReactFlowCustomNestedPoint,
    FlyteNode_nestedMaxDepth: ReactFlowCustomMaxNested
};

/**
 * Renderless component waits for ReactFlow to give rendered
 * dimensions before computing layout
 * @param props:LayoutRC
 * @returns: void
 */
const LayoutRC: React.FC<LayoutRCProps> = ({
    setElements,
    setLayout
}: LayoutRCProps) => {
    const [computeLayout, setComputeLayout] = useState(true);

    /* strore is only populated onLoad for each flow */
    const nodes = useStoreState(store => store.nodes);
    const edges = useStoreState(store => store.edges);

    if (nodes.length > 0 && computeLayout) {
        if (nodes[0].__rf.width) {
            setComputeLayout(false);
        }
    }
    useEffect(() => {
        if (!computeLayout) {
            const nodesAndEdges = (nodes as any[]).concat(edges);
            const graph = setReactFlowGraphLayout(nodesAndEdges, 'LR');
            setElements(graph);
            setLayout(true);
        }
    }, [computeLayout]);

    return null;
};

/**
 * Notes:
 * To support nested graphs we wrap each flow inside its own provider/store
 * which allows us to contextualize fitView to only render when its own
 * elements change (not parents/children)
 *
 * Workflow:
 *  - set initial (unpositioned) elements and wait for onload
 *  - position elements (with rendered dimensions) in <LayoutRC>
 *  - fit view
 *
 * @see   https://reactflow.dev/docs/
 * @param props:ReactFlowWrapperProps
 * @returns rendered component
 */
export const ReactFlowWrapper: React.FC<RFWrapperProps> = ({
    rfGraphJson,
    backgroundStyle
}) => {
    const [elements, setElements] = useState(rfGraphJson);
    const [layedOut, setLayout] = useState(false);
    const [reactFlowInstance, setReactFlowInstance] = useState<null | any>(
        null
    );

    const onLoad = (rf: any) => {
        setReactFlowInstance(rf);
    };

    /**
     * Note: setLayout passed/called by <LayoutRC>
     */
    useEffect(() => {
        if (layedOut && reactFlowInstance) {
            reactFlowInstance?.fitView({ padding: 0 });
        }
    }, [layedOut, reactFlowInstance]);
    /**
     * STEPS:
     *  - have each node click return nodes {text.data}
     *  - Then figure out the input needed for the slide out (execution id? node id?)
     *  - Append that to the rf nodes {data} object.
     */

    return (
        <ReactFlowProvider>
            <ReactFlow
                elements={elements}
                onLoad={onLoad}
                nodeTypes={CustomNodeTypes}
            >
                <Background
                    style={backgroundStyle.background}
                    color={backgroundStyle.gridColor}
                    gap={backgroundStyle.gridSpacing}
                />
            </ReactFlow>
            <LayoutRC
                setLayout={setLayout}
                setElements={setElements}
            ></LayoutRC>
        </ReactFlowProvider>
    );
};
