import { NodeExecution } from 'models';
import { useContext, useMemo } from 'react';
import { ExecutionDataCacheContext } from './contexts';
import { NodeExecutionGroup } from './types';
import { mapNodeExecutionDetails } from './utils';

/** Decorates a list of NodeExecutions, mapping the list items to
 * `DetailedNodeExecution`s. The node details are pulled from the the nearest
 * `ExecutionContext.dataCache`.
 */
export function useDetailedNodeExecutions(nodeExecutions: NodeExecution[]) {
    const dataCache = useContext(ExecutionDataCacheContext);

    return useMemo(() => mapNodeExecutionDetails(nodeExecutions, dataCache), [
        nodeExecutions,
        dataCache
    ]);
}

/** Decorates a list of `NodeExecutionGroup`s, transforming their lists of
 * `NodeExecution`s into `DetailedNodeExecution`s.
 */
export function useDetailedChildNodeExecutions(
    nodeExecutionGroups: NodeExecutionGroup[]
) {
    const dataCache = useContext(ExecutionDataCacheContext);
    return useMemo(
        () =>
            nodeExecutionGroups.map(group => ({
                ...group,
                nodeExecutions: mapNodeExecutionDetails(
                    group.nodeExecutions,
                    dataCache
                )
            })),
        [nodeExecutionGroups, dataCache]
    );
}
