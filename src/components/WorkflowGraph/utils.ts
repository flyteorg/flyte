import { DAGNode } from 'models/Graph/types';
import { endNodeId, startNodeId } from 'models/Node/constants';

export function isStartNode(node: DAGNode) {
    return node.id === startNodeId;
}

export function isEndNode(node: DAGNode) {
    return node.id === endNodeId;
}
