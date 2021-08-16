import { createDebugLogger } from 'common/log';
import { createTimer } from 'common/timer';
import { cloneDeep, keyBy, values } from 'lodash';
import { identifierToString } from 'models/Common/utils';
import { CompiledWorkflowClosure } from 'models/Workflow/types';
import { nodeIds } from './constants';
import { DAGNode } from './types';

const log = createDebugLogger('models/Workflow');

/** Converts a Flyte graph spec to one which can be consumed by the `Graph`
 * component. This mostly involves gathering the nodes into an array and
 * assigning them `parentId` values based on the connections defined in the
 * graph spec.
 */
export function convertFlyteGraphToDAG(
    workflow: CompiledWorkflowClosure
): DAGNode[] {

    const timer = createTimer();

    const {
        primary: {
            template: { nodes: nodesInput },
            connections
        },
        tasks
    } = workflow;

    const tasksById = keyBy(tasks, task =>
        identifierToString(task.template.id)
    );

    const nodes = cloneDeep(nodesInput);
    // For any nodes that have a `taskNode` field, store a reference to
    // the task template for use later during rendering
    nodes.forEach(node => {
        if (!node.taskNode) {
            return;
        }
        const { referenceId } = node.taskNode;
        const task = tasksById[identifierToString(referenceId)];
        if (!task) {
            log(`Node ${node.id} references missing task: ${referenceId}`);
            return;
        }
        (node as DAGNode).taskTemplate = task.template;
    });

    const nodeMap: Record<string, DAGNode> = keyBy(nodes, 'id');
    const connectionMap: Map<string, Map<string, boolean>> = new Map();

    Object.keys(connections.downstream).forEach(parentId => {
        const edges = connections.downstream[parentId];
        edges.ids.forEach(id => {
            const node = nodeMap[id];
            if (!node) {
                log(`Ignoring edge for missing node: ${id}`);
                return;
            }
            if (!node.parentIds) {
                node.parentIds = [];
                connectionMap.set(node.id, new Map());
            }

            const connectionsForNode = connectionMap.get(node.id)!;

            // Ignore empty node ids and check for duplicates
            if (parentId.length > 0 && !connectionsForNode.has(parentId)) {
                node.parentIds.push(parentId);
                connectionsForNode.set(parentId, true);
            }
        });
    });

    // Filter out any nodes with no parents (except for the start node)
    const result = values(nodeMap).filter(
        n => n.id === nodeIds.start || (n.parentIds && n.parentIds.length > 0)
    );

    log(`Compilation time: ${timer.timeStringMS}`);
    return result;
}
