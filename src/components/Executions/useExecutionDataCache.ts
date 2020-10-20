import { log } from 'common/log';
import { getCacheKey } from 'components/Cache';
import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import {
    fetchNodeExecutions,
    fetchTaskExecutionChildren
} from 'components/hooks';
import {
    extractAndIdentifyNodes,
    extractTaskTemplates
} from 'components/hooks/utils';
import { NotFoundError } from 'errors';
import {
    Execution,
    GloballyUniqueNode,
    Identifier,
    NodeExecution,
    NodeExecutionIdentifier,
    nodeExecutionQueryParams,
    NodeId,
    RequestConfig,
    TaskExecutionIdentifier,
    TaskTemplate,
    Workflow,
    WorkflowExecutionIdentifier,
    WorkflowId
} from 'models';
import { useState } from 'react';
import { ExecutionDataCache } from './types';
import { fetchTaskExecutions } from './useTaskExecutions';
import { getNodeExecutionSpecId } from './utils';

function cacheItems<T extends { id: object | string }>(
    map: Map<string, T>,
    values: T[]
) {
    values.forEach(v => map.set(getCacheKey(v.id), v));
}

/** Creates a new ExecutionDataCache which will use the provided API context
 * to fetch entities.
 */
export function createExecutionDataCache(
    apiContext: APIContextValue
): ExecutionDataCache {
    const workflowsById: Map<string, Workflow> = new Map();
    const nodesById: Map<string, GloballyUniqueNode> = new Map();
    const taskTemplatesById: Map<string, TaskTemplate> = new Map();
    const workflowExecutionIdToWorkflowId: Map<string, WorkflowId> = new Map();

    const insertNodes = (nodes: GloballyUniqueNode[]) => {
        cacheItems(nodesById, nodes);
    };

    const insertTaskTemplates = (templates: TaskTemplate[]) => {
        cacheItems(taskTemplatesById, templates);
    };

    const insertWorkflow = (workflow: Workflow) => {
        workflowsById.set(getCacheKey(workflow.id), workflow);
        insertNodes(extractAndIdentifyNodes(workflow));
        insertTaskTemplates(extractTaskTemplates(workflow));
    };

    const insertExecution = (execution: Execution) => {
        workflowExecutionIdToWorkflowId.set(
            getCacheKey(execution.id),
            execution.closure.workflowId
        );
    };

    const insertWorkflowExecutionReference = (
        executionId: WorkflowExecutionIdentifier,
        workflowId: WorkflowId
    ) => {
        workflowExecutionIdToWorkflowId.set(
            getCacheKey(executionId),
            workflowId
        );
    };

    const getWorkflow = async (id: WorkflowId) => {
        const key = getCacheKey(id);
        if (workflowsById.has(key)) {
            return workflowsById.get(key)!;
        }
        const workflow = await apiContext.getWorkflow(id);
        insertWorkflow(workflow);
        return workflow;
    };

    const getNode = (id: NodeId) => {
        const node = nodesById.get(getCacheKey(id));
        if (node === undefined) {
            log.error('Unexpected Node missing from cache:', id);
        }
        return node;
    };

    const getNodeForNodeExecution = (nodeExecution: NodeExecution) => {
        const { executionId } = nodeExecution.id;
        const nodeId = getNodeExecutionSpecId(nodeExecution);
        const workflowExecutionKey = getCacheKey(executionId);
        if (!workflowExecutionIdToWorkflowId.has(workflowExecutionKey)) {
            log.error(
                'Unexpected missing parent workflow execution: ',
                executionId
            );
            return null;
        }
        const workflowId = workflowExecutionIdToWorkflowId.get(
            workflowExecutionKey
        )!;
        return getNode({ nodeId, workflowId });
    };

    const getNodeExecutions = async (
        id: WorkflowExecutionIdentifier,
        config: RequestConfig
    ) => {
        const execution = await getWorkflowExecution(id);
        // Fetching workflow to ensure node definitions exist
        const [_, nodeExecutions] = await Promise.all([
            getWorkflow(execution.closure.workflowId),
            fetchNodeExecutions({ config, id }, apiContext)
        ]);
        return nodeExecutions;
    };

    const getNodeExecutionsForParentNode = async (
        { executionId, nodeId }: NodeExecutionIdentifier,
        config: RequestConfig
    ) => {
        const childrenPromise = fetchNodeExecutions(
            {
                config: {
                    ...config,
                    params: {
                        ...config.params,
                        [nodeExecutionQueryParams.parentNodeId]: nodeId
                    }
                },
                id: executionId
            },
            apiContext
        );
        const workflowPromise = getWorkflowIdForWorkflowExecution(
            executionId
        ).then(workflowId => getWorkflow(workflowId));

        const [children] = await Promise.all([
            childrenPromise,
            workflowPromise
        ]);
        return children;
    };

    const getTaskTemplate = (id: Identifier) => {
        const template = taskTemplatesById.get(getCacheKey(id));
        if (template === undefined) {
            log.error('Unexpected TaskTemplate missing from cache:', id);
        }
        return template;
    };

    const getOrFetchTaskTemplate = async (id: Identifier) => {
        const key = getCacheKey(id);
        if (taskTemplatesById.has(key)) {
            return taskTemplatesById.get(key);
        }
        try {
            const { template } = (
                await apiContext.getTask(id)
            ).closure.compiledTask;
            taskTemplatesById.set(key, template);
            return template;
        } catch (e) {
            if (e instanceof NotFoundError) {
                log.warn('No task template found for task: ', id);
                return;
            }
            throw e;
        }
    };

    const getWorkflowExecution = async (id: WorkflowExecutionIdentifier) => {
        const execution = await apiContext.getExecution(id);
        insertWorkflowExecutionReference(
            execution.id,
            execution.closure.workflowId
        );
        return execution;
    };

    const getWorkflowIdForWorkflowExecution = async (
        id: WorkflowExecutionIdentifier
    ) => {
        const key = getCacheKey(id);
        if (workflowExecutionIdToWorkflowId.has(key)) {
            return workflowExecutionIdToWorkflowId.get(key)!;
        }
        return (await getWorkflowExecution(id)).closure.workflowId;
    };

    const getTaskExecutions = async (id: NodeExecutionIdentifier) =>
        fetchTaskExecutions(id, apiContext);

    const getTaskExecutionChildren = async (
        id: TaskExecutionIdentifier,
        config: RequestConfig
    ) => {
        const childrenPromise = fetchTaskExecutionChildren(
            { config, taskExecutionId: id },
            apiContext
        );
        const workflowIdPromise = getWorkflowIdForWorkflowExecution(
            id.nodeExecutionId.executionId
        );

        const cacheTaskTemplatePromise = await getOrFetchTaskTemplate(
            id.taskId
        );

        const [children, workflowId, _] = await Promise.all([
            childrenPromise,
            workflowIdPromise,
            cacheTaskTemplatePromise
        ]);

        // We need to synthesize a record for each child node,
        // as they won't exist in any Workflow closure.
        const nodes = children.map<GloballyUniqueNode>(node => ({
            id: {
                workflowId,
                nodeId: node.id.nodeId
            },
            node: {
                id: node.id.nodeId,
                taskNode: {
                    referenceId: id.taskId
                }
            }
        }));
        insertNodes(nodes);

        return children;
    };

    return {
        getNode,
        getNodeForNodeExecution,
        getNodeExecutions,
        getNodeExecutionsForParentNode,
        getTaskExecutions,
        getTaskExecutionChildren,
        getTaskTemplate,
        getWorkflow,
        getWorkflowExecution,
        getWorkflowIdForWorkflowExecution,
        insertExecution,
        insertNodes,
        insertTaskTemplates,
        insertWorkflow,
        insertWorkflowExecutionReference
    };
}

/** A hook for creating a new ExecutionDataCache wired to the nearest `APIContext` */
export function useExecutionDataCache() {
    const apiContext = useAPIContext();
    const [dataCache] = useState(() => createExecutionDataCache(apiContext));
    return dataCache;
}
