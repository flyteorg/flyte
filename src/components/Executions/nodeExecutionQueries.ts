import { QueryInput, QueryType } from 'components/data/types';
import { useConditionalQuery } from 'components/hooks/useConditionalQuery';
import { isEqual } from 'lodash';
import { RequestConfig } from 'models/AdminEntity/types';
import {
    getNodeExecution,
    listNodeExecutions,
    listTaskExecutionChildren
} from 'models/Execution/api';
import { nodeExecutionQueryParams } from 'models/Execution/constants';
import {
    NodeExecution,
    NodeExecutionIdentifier,
    TaskExecutionIdentifier,
    WorkflowExecutionIdentifier
} from 'models/Execution/types';
import { endNodeId, startNodeId } from 'models/Node/constants';
import { QueryClient, QueryObserverResult, useQueryClient } from 'react-query';
import { fetchTaskExecutionList } from './taskExecutionQueries';
import { formatRetryAttempt } from './TaskExecutionsList/utils';
import { NodeExecutionGroup } from './types';
import { isParentNode, nodeExecutionIsTerminal } from './utils';

const ignoredNodeIds = [startNodeId, endNodeId];
function removeSystemNodes(nodeExecutions: NodeExecution[]): NodeExecution[] {
    return nodeExecutions.filter(ne => {
        if (ignoredNodeIds.includes(ne.id.nodeId)) {
            return false;
        }
        const specId = ne.metadata?.specNodeId;
        if (specId != null && ignoredNodeIds.includes(specId)) {
            return false;
        }
        return true;
    });
}

/** A query for fetching a single `NodeExecution` by id. */
export function makeNodeExecutionQuery(
    id: NodeExecutionIdentifier
): QueryInput<NodeExecution> {
    return {
        queryKey: [QueryType.NodeExecution, id],
        queryFn: () => getNodeExecution(id)
    };
}

/** Composable fetch function which wraps `makeNodeExecutionQuery` */
export function fetchNodeExecution(
    queryClient: QueryClient,
    id: NodeExecutionIdentifier
) {
    return queryClient.fetchQuery(makeNodeExecutionQuery(id));
}

// On successful node execution list queries, extract and store all
// executions so they are individually fetchable from the cache.
function cacheNodeExecutions(
    queryClient: QueryClient,
    nodeExecutions: NodeExecution[]
) {
    nodeExecutions.forEach(ne =>
        queryClient.setQueryData([QueryType.NodeExecution, ne.id], ne)
    );
}

/** A query for fetching a list of `NodeExecution`s which are children of a given
 * `Execution`.
 */
export function makeNodeExecutionListQuery(
    queryClient: QueryClient,
    id: WorkflowExecutionIdentifier,
    config?: RequestConfig
): QueryInput<NodeExecution[]> {
    return {
        queryKey: [QueryType.NodeExecutionList, id, config],
        queryFn: async () => {
            const nodeExecutions = removeSystemNodes(
                (await listNodeExecutions(id, config)).entities
            );
            nodeExecutions.map(exe => {
                if (exe.metadata) {
                    return (exe.scopedId = exe.metadata.specNodeId);
                } else {
                    return (exe.scopedId = exe.id.nodeId);
                }
            });
            cacheNodeExecutions(queryClient, nodeExecutions);
            return nodeExecutions;
        }
    };
}

/** Composable fetch function which wraps `makeNodeExecutionListQuery`. */
export function fetchNodeExecutionList(
    queryClient: QueryClient,
    id: WorkflowExecutionIdentifier,
    config?: RequestConfig
) {
    return queryClient.fetchQuery(
        makeNodeExecutionListQuery(queryClient, id, config)
    );
}

/** A query for fetching a list of `NodeExecution`s which are children of a given
 * `TaskExecution`.
 */
export function makeTaskExecutionChildListQuery(
    queryClient: QueryClient,
    id: TaskExecutionIdentifier,
    config?: RequestConfig
): QueryInput<NodeExecution[]> {
    return {
        queryKey: [QueryType.TaskExecutionChildList, id, config],
        queryFn: async () => {
            const nodeExecutions = removeSystemNodes(
                (await listTaskExecutionChildren(id, config)).entities
            );
            cacheNodeExecutions(queryClient, nodeExecutions);
            return nodeExecutions;
        },
        onSuccess: nodeExecutions => {
            nodeExecutions.forEach(ne =>
                queryClient.setQueryData([QueryType.NodeExecution, ne.id], ne)
            );
        }
    };
}

/** Composable fetch function which wraps `makeTaskExecutionChildListQuery`. */
export function fetchTaskExecutionChildList(
    queryClient: QueryClient,
    id: TaskExecutionIdentifier,
    config?: RequestConfig
) {
    return queryClient.fetchQuery(
        makeTaskExecutionChildListQuery(queryClient, id, config)
    );
}

/** --- Queries for fetching children of a NodeExecution --- **/

async function fetchGroupForTaskExecution(
    queryClient: QueryClient,
    taskExecutionId: TaskExecutionIdentifier,
    config: RequestConfig
): Promise<NodeExecutionGroup> {
    return {
        // NodeExecutions created by a TaskExecution are grouped
        // by the retry attempt of the task.
        name: formatRetryAttempt(taskExecutionId.retryAttempt),
        nodeExecutions: await fetchTaskExecutionChildList(
            queryClient,
            taskExecutionId,
            config
        )
    };
}

async function fetchGroupForWorkflowExecution(
    queryClient: QueryClient,
    executionId: WorkflowExecutionIdentifier,
    config: RequestConfig
): Promise<NodeExecutionGroup> {
    return {
        // NodeExecutions created by a workflow execution are grouped
        // by the execution id, since workflow executions are not retryable.
        name: executionId.name,
        nodeExecutions: await fetchNodeExecutionList(
            queryClient,
            executionId,
            config
        )
    };
}

async function fetchGroupsForTaskExecutionNode(
    queryClient: QueryClient,
    nodeExecution: NodeExecution,
    config: RequestConfig
): Promise<NodeExecutionGroup[]> {
    const taskExecutions = await fetchTaskExecutionList(
        queryClient,
        nodeExecution.id,
        config
    );

    // For TaskExecutions marked as parents, fetch its children and create a group.
    // Otherwise, return null and we will filter it out later.
    const groups = await Promise.all(
        taskExecutions.map(execution =>
            execution.isParent
                ? fetchGroupForTaskExecution(queryClient, execution.id, config)
                : Promise.resolve(null)
        )
    );

    // Remove any empty groups
    return groups.filter(
        group => group !== null && group.nodeExecutions.length > 0
    ) as NodeExecutionGroup[];
}

async function fetchGroupsForWorkflowExecutionNode(
    queryClient: QueryClient,
    nodeExecution: NodeExecution,
    config: RequestConfig
): Promise<NodeExecutionGroup[]> {
    if (!nodeExecution.closure.workflowNodeMetadata) {
        throw new Error('Unexpected empty `workflowNodeMetadata`');
    }
    const { executionId } = nodeExecution.closure.workflowNodeMetadata;
    // We can only have one WorkflowExecution (no retries), so there is only
    // one group to return. But calling code expects it as an array.
    const group = await fetchGroupForWorkflowExecution(
        queryClient,
        executionId,
        config
    );
    return group.nodeExecutions.length > 0 ? [group] : [];
}

async function fetchGroupsForParentNodeExecution(
    queryClient: QueryClient,
    nodeExecution: NodeExecution,
    config: RequestConfig
): Promise<NodeExecutionGroup[]> {
    const finalConfig = {
        ...config,
        params: {
            ...config.params,
            [nodeExecutionQueryParams.parentNodeId]: nodeExecution.id.nodeId
        }
    };

    /** @TODO there is likely a better way to do this; eg, in a previous call */
    if (!nodeExecution.scopedId) {
        nodeExecution.scopedId = nodeExecution.metadata?.specNodeId;
    }

    const children = await fetchNodeExecutionList(
        queryClient,
        nodeExecution.id.executionId,
        finalConfig
    );
    const groupsByName = children.reduce<Map<string, NodeExecutionGroup>>(
        (out, child) => {
            const retryAttempt = formatRetryAttempt(child.metadata?.retryGroup);
            let group = out.get(retryAttempt);
            if (!group) {
                group = { name: retryAttempt, nodeExecutions: [] };
                out.set(retryAttempt, group);
            }
            /**
             * GraphUX uses workflowClosure which uses scopedId
             * This builds a scopedId via parent nodeExecution
             * to enable mapping between graph and other components
             */
            let scopedId: string | undefined =
                nodeExecution.metadata?.specNodeId;
            if (scopedId != undefined) {
                scopedId += `-${child.metadata?.retryGroup}-${child.metadata?.specNodeId}`;
                child['scopedId'] = scopedId;
            } else {
                child['scopedId'] = child.metadata?.specNodeId;
            }
            group.nodeExecutions.push(child);
            return out;
        },
        new Map()
    );
    return Array.from(groupsByName.values());
}

function fetchChildNodeExecutionGroups(
    queryClient: QueryClient,
    nodeExecution: NodeExecution,
    config: RequestConfig
) {
    const { workflowNodeMetadata } = nodeExecution.closure;
    // Newer NodeExecution structures can directly indicate their parent
    // status and have their children fetched in bulk.
    if (isParentNode(nodeExecution)) {
        return fetchGroupsForParentNodeExecution(
            queryClient,
            nodeExecution,
            config
        );
    }
    // Otherwise, we need to determine the type of the node and
    // recursively fetch NodeExecutions for the corresponding Workflow
    // or Task executions.
    if (
        workflowNodeMetadata &&
        !isEqual(workflowNodeMetadata.executionId, nodeExecution.id.executionId)
    ) {
        return fetchGroupsForWorkflowExecutionNode(
            queryClient,
            nodeExecution,
            config
        );
    }
    return fetchGroupsForTaskExecutionNode(queryClient, nodeExecution, config);
}

/**
 * Query returns all children for a list of `nodeExecutions`
 * Note: diffrent from fetchGroupsForParentNodeExecution in that it expects a
 * list of nodeExecitions
 */
async function fetchAllChildNodeExecutions(
    queryClient: QueryClient,
    nodeExecutions: NodeExecution[],
    config: RequestConfig
): Promise<Array<NodeExecutionGroup[]>> {
    const executions: Array<NodeExecutionGroup[]> = await Promise.all(
        nodeExecutions.map(exe => {
            return fetchChildNodeExecutionGroups(queryClient, exe, config);
        })
    );
    return executions;
}
/**
 *
 * @param nodeExecutions list of parent node executionId's
 * @param config
 * @returns
 */
export function useAllChildNodeExecutionGroupsQuery(
    nodeExecutions: NodeExecution[],
    config: RequestConfig
): QueryObserverResult<Array<NodeExecutionGroup[]>, Error> {
    const queryClient = useQueryClient();
    const shouldEnableFn = groups => {
        if (nodeExecutions[0] && groups.length > 0) {
            if (!nodeExecutionIsTerminal(nodeExecutions[0])) {
                return true;
            }
            return groups.some(group => {
                if (group.nodeExecutions?.length > 0) {
                    return group.nodeExecutions.some(ne => {
                        return !nodeExecutionIsTerminal(ne);
                    });
                } else {
                    return false;
                }
            });
        } else {
            return false;
        }
    };

    return useConditionalQuery<Array<NodeExecutionGroup[]>>(
        {
            queryKey: [
                QueryType.NodeExecutionChildList,
                nodeExecutions[0]?.id,
                config
            ],
            queryFn: () =>
                fetchAllChildNodeExecutions(queryClient, nodeExecutions, config)
        },
        shouldEnableFn
    );
}

/** Fetches and groups `NodeExecution`s which are direct children of the given
 * `NodeExecution`.
 */
export function useChildNodeExecutionGroupsQuery(
    nodeExecution: NodeExecution,
    config: RequestConfig
): QueryObserverResult<NodeExecutionGroup[], Error> {
    const queryClient = useQueryClient();
    // Use cached data if the parent node execution is terminal and all children
    // in all groups are terminal
    const shouldEnableFn = (groups: NodeExecutionGroup[]) => {
        if (!nodeExecutionIsTerminal(nodeExecution)) {
            return true;
        }
        return groups.some(group =>
            group.nodeExecutions.some(ne => !nodeExecutionIsTerminal(ne))
        );
    };

    return useConditionalQuery<NodeExecutionGroup[]>(
        {
            queryKey: [
                QueryType.NodeExecutionChildList,
                nodeExecution.id,
                config
            ],
            queryFn: () =>
                fetchChildNodeExecutionGroups(
                    queryClient,
                    nodeExecution,
                    config
                )
        },
        shouldEnableFn
    );
}
