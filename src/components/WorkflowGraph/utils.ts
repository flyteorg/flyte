import { DAGNode, endNodeId, startNodeId } from 'models';

// TODO: Checking start/end should also be based on presence of parents/children
export function isStartNode(node: DAGNode) {
    return node.id === startNodeId;
}

export function isEndNode(node: DAGNode) {
    return node.id === endNodeId;
}
