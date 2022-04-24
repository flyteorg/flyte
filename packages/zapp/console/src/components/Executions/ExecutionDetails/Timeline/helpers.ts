import { endNodeId, startNodeId } from 'models/Node/constants';
import { isExpanded } from 'components/WorkflowGraph/utils';
import { dNode } from 'models/Graph/types';

export const TimeZone = {
  Local: 'local',
  UTC: 'utc',
};

export function isTransitionNode(node: dNode) {
  // In case of bracnhNode childs, start and end nodes could be present as 'n0-start-node' etc.
  return node.id.includes(startNodeId) || node.id.includes(endNodeId);
}

export function convertToPlainNodes(nodes: dNode[], level = 0): dNode[] {
  const result: dNode[] = [];
  if (!nodes || nodes.length === 0) {
    return result;
  }
  nodes.forEach((node) => {
    if (isTransitionNode(node)) {
      return;
    }
    result.push({ ...node, level });
    if (node.nodes.length > 0 && isExpanded(node)) {
      result.push(...convertToPlainNodes(node.nodes, level + 1));
    }
  });
  return result;
}
