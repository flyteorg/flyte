import { dEdge, dNode, dTypes } from 'models/Graph/types';
import {
    DISPLAY_NAME_START,
    DISPLAY_NAME_END
} from 'components/WorkflowGraph/utils';
import { MAX_RENDER_DEPTH, ReactFlowGraphConfig } from './utils';
import { Edge, Elements, Node, Position } from 'react-flow-renderer';
import { NodeExecutionsById } from 'models/Execution/types';
import { NodeExecutionPhase } from 'models/Execution/enums';

export const buildCustomNodeName = (type: dTypes) => {
    return `${ReactFlowGraphConfig.customNodePrefix}_${dTypes[type]}`;
};

export const buildReactFlowEdge = (edge: dEdge): Edge => {
    return {
        id: `[${edge.sourceId}]->[${edge.targetId}]`,
        source: edge.sourceId,
        target: edge.targetId,
        sourceHandle: 'left-handle',
        arrowHeadType: ReactFlowGraphConfig.arrowHeadType,
        type: ReactFlowGraphConfig.edgeType
    } as Edge;
};

export const buildReactFlowNode = (
    dNode: dNode,
    dag: any = [],
    nodeExecutionsById: NodeExecutionsById,
    typeOverride?: dTypes | null,
    onNodeSelectionChanged?: any | null
): Node => {
    const type = typeOverride ? typeOverride : dNode.type;
    const taskType = dNode?.value?.template ? dNode.value.template.type : null;

    /**
     * @TODO decide which to display after demo
     */
    const displayName =
        dNode.name == DISPLAY_NAME_START || dNode.name == DISPLAY_NAME_END
            ? dNode.name
            : dNode.scopedId;

    const mapNodeExecutionStatus = () => {
        if (nodeExecutionsById[dNode.scopedId]) {
            return nodeExecutionsById[dNode.scopedId].closure
                .phase as NodeExecutionPhase;
        } else {
            return NodeExecutionPhase.SKIPPED;
        }
    };
    const nodeExecutionStatus = mapNodeExecutionStatus();

    return {
        id: dNode.scopedId,
        type: buildCustomNodeName(type),
        data: {
            nodeExecutionStatus: nodeExecutionStatus,
            text: displayName,
            handles: [],
            nodeType: type,
            scopedId: dNode.scopedId,
            dag: dag,
            taskType: taskType,
            onNodeSelectionChanged: () => {
                onNodeSelectionChanged([dNode.scopedId]);
            }
        },
        position: { x: 0, y: 0 },
        sourcePosition: Position.Right,
        targetPosition: Position.Left
    } as Node;
};

export const nodeMapToArr = map => {
    const output: any[] = [];
    for (const k in map) {
        output.push(map[k]);
    }
    return output;
};

export const dagToReactFlow = (
    dag: dNode,
    nodeExecutionsById: NodeExecutionsById,
    currentDepth = 0,
    onNodeSelectionChanged
) => {
    const nodes: any = {};
    const edges: any = {};

    dag.nodes?.map(dNode => {
        if (dNode.nodes?.length > 0 && currentDepth <= MAX_RENDER_DEPTH) {
            /* Note: currentDepth will be replaced once nested toggle */
            if (currentDepth == MAX_RENDER_DEPTH) {
                nodes[dNode.id] = buildReactFlowNode(
                    dNode,
                    [],
                    nodeExecutionsById,
                    dTypes.nestedMaxDepth,
                    onNodeSelectionChanged
                );
            } else {
                nodes[dNode.id] = buildReactFlowNode(
                    dNode,
                    dagToReactFlow(
                        dNode,
                        nodeExecutionsById,
                        currentDepth + 1,
                        onNodeSelectionChanged
                    ),
                    nodeExecutionsById,
                    null,
                    onNodeSelectionChanged
                );
            }
        } else {
            nodes[dNode.id] = buildReactFlowNode(
                dNode,
                [],
                nodeExecutionsById,
                null,
                onNodeSelectionChanged
            );
        }
    });
    dag.edges?.map(edge => {
        const rfEdge = buildReactFlowEdge(edge);
        edges[rfEdge.id] = rfEdge;
    });
    const output = nodeMapToArr(nodes).concat(nodeMapToArr(edges));
    return output;
};

export const ConvertFlyteDagToReactFlows = (
    root: dNode,
    nodeExecutionsById: NodeExecutionsById,
    onNodeSelectionChanged
): Elements => {
    const rfJson = dagToReactFlow(
        root,
        nodeExecutionsById,
        0,
        onNodeSelectionChanged
    );

    return rfJson;
};
