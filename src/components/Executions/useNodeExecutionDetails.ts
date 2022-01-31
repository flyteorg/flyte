import { log } from 'common/log';
import { QueryType } from 'components/data/types';
import { fetchTaskTemplate } from 'components/Task/taskQueries';
import { fetchWorkflow } from 'components/Workflow/workflowQueries';
import { Identifier } from 'models/Common/types';
import { NodeExecution } from 'models/Execution/types';
import { CompiledNode } from 'models/Node/types';
import { TaskTemplate } from 'models/Task/types';
import { CompiledWorkflow, Workflow } from 'models/Workflow/types';
import { QueryClient, useQuery, useQueryClient } from 'react-query';
import { fetchTaskExecutionList } from './taskExecutionQueries';
import {
    CompiledBranchNode,
    CompiledWorkflowNode,
    NodeExecutionDetails,
    NodeExecutionDisplayType,
    WorkflowNodeExecution
} from './types';
import { fetchWorkflowExecution } from './useWorkflowExecution';
import {
    getNodeExecutionSpecId,
    isCompiledBranchNode,
    isCompiledTaskNode,
    isCompiledWorkflowNode,
    isWorkflowNodeExecution,
    flattenBranchNodes
} from './utils';
function createExternalWorkflowNodeExecutionDetails(
    workflow: Workflow
): NodeExecutionDetails {
    return {
        displayId: workflow.id.name,
        displayType: NodeExecutionDisplayType.Workflow
    };
}

function createWorkflowNodeExecutionDetails(
    nodeExecution: NodeExecution,
    node: CompiledWorkflowNode
): NodeExecutionDetails {
    const displayType = NodeExecutionDisplayType.Workflow;
    let displayId = '';
    let displayName = '';
    const { launchplanRef, subWorkflowRef } = node.workflowNode;
    const identifier = (launchplanRef
        ? launchplanRef
        : subWorkflowRef) as Identifier;
    if (!identifier) {
        log.warn(
            `Unexpected workflow node with no ref: ${getNodeExecutionSpecId(
                nodeExecution
            )}`
        );
    } else {
        displayId = node.id;
        displayName = identifier.name;
    }

    return {
        displayId,
        displayName,
        displayType
    };
}

// TODO: https://github.com/lyft/flyte/issues/655
function createBranchNodeExecutionDetails(
    node: CompiledBranchNode
): NodeExecutionDetails {
    return {
        displayId: node.id,
        displayName: 'branchNode',
        displayType: NodeExecutionDisplayType.BranchNode
    };
}

function createTaskNodeExecutionDetails(
    taskTemplate: TaskTemplate,
    displayId: string | undefined
): NodeExecutionDetails {
    return {
        taskTemplate,
        displayId: displayId,
        displayName: taskTemplate.id.name,
        displayType: taskTemplate.type
    };
}

function createUnknownNodeExecutionDetails(): NodeExecutionDetails {
    return {
        displayId: '',
        displayType: NodeExecutionDisplayType.Unknown
    };
}

async function fetchExternalWorkflowNodeExecutionDetails(
    queryClient: QueryClient,
    nodeExecution: WorkflowNodeExecution
): Promise<NodeExecutionDetails> {
    const workflowExecution = await fetchWorkflowExecution(
        queryClient,
        nodeExecution.closure.workflowNodeMetadata.executionId
    );
    const workflow = await fetchWorkflow(
        queryClient,
        workflowExecution.closure.workflowId
    );

    return createExternalWorkflowNodeExecutionDetails(workflow);
}

function findCompiledNode(
    nodeId: string,
    compiledWorkflows: CompiledWorkflow[]
) {
    for (let i = 0; i < compiledWorkflows.length; i += 1) {
        const found = compiledWorkflows[i].template.nodes
            .map(flattenBranchNodes)
            .flat()
            .find(({ id }) => id === nodeId);
        if (found) {
            return found;
        }
    }
    return undefined;
}

function findNodeInWorkflow(
    nodeId: string,
    workflow: Workflow
): CompiledNode | undefined {
    if (!workflow.closure?.compiledWorkflow) {
        return undefined;
    }
    const { primary, subWorkflows = [] } = workflow.closure?.compiledWorkflow;
    return findCompiledNode(nodeId, [primary, ...subWorkflows]);
}

async function fetchTaskNodeExecutionDetails(
    queryClient: QueryClient,
    taskId: Identifier,
    displayId: string | undefined
) {
    const taskTemplate = await fetchTaskTemplate(queryClient, taskId);
    if (!taskTemplate) {
        throw new Error(
            `Unexpected missing task template while fetching NodeExecution details: ${JSON.stringify(
                taskId
            )}`
        );
    }
    return createTaskNodeExecutionDetails(taskTemplate, displayId);
}

async function fetchNodeExecutionDetailsFromNodeSpec(
    queryClient: QueryClient,
    nodeExecution: NodeExecution
): Promise<NodeExecutionDetails> {
    const nodeId = getNodeExecutionSpecId(nodeExecution);
    const workflowExecution = await fetchWorkflowExecution(
        queryClient,
        nodeExecution.id.executionId
    );
    const workflow = await fetchWorkflow(
        queryClient,
        workflowExecution.closure.workflowId
    );

    // If the source workflow spec has a node matching this execution, we
    // can parse out the node information and set our details based on that.
    const compiledNode = findNodeInWorkflow(nodeId, workflow);
    if (compiledNode) {
        if (isCompiledTaskNode(compiledNode)) {
            return fetchTaskNodeExecutionDetails(
                queryClient,
                compiledNode.taskNode.referenceId,
                compiledNode.id
            );
        }
        if (isCompiledWorkflowNode(compiledNode)) {
            return createWorkflowNodeExecutionDetails(
                nodeExecution,
                compiledNode
            );
        }
        if (isCompiledBranchNode(compiledNode)) {
            return createBranchNodeExecutionDetails(compiledNode);
        }
    }

    // Fall back to attempting to locate a task execution for this node and
    // subsequently fetching its task spec.
    const taskExecutions = await fetchTaskExecutionList(
        queryClient,
        nodeExecution.id
    );
    if (taskExecutions.length > 0) {
        return fetchTaskNodeExecutionDetails(
            queryClient,
            taskExecutions[0].id.taskId,
            undefined
        );
    }

    return createUnknownNodeExecutionDetails();
}

async function doFetchNodeExecutionDetails(
    queryClient: QueryClient,
    nodeExecution: NodeExecution
) {
    try {
        if (isWorkflowNodeExecution(nodeExecution)) {
            return fetchExternalWorkflowNodeExecutionDetails(
                queryClient,
                nodeExecution
            );
        }

        // Attempt to find node information in the source workflow spec
        // or via any associated TaskExecution's task spec.
        return fetchNodeExecutionDetailsFromNodeSpec(
            queryClient,
            nodeExecution
        );
    } catch (e) {
        return createUnknownNodeExecutionDetails();
    }
}

export function fetchNodeExecutionDetails(
    queryClient: QueryClient,
    nodeExecution: NodeExecution
) {
    return queryClient.fetchQuery({
        queryKey: [QueryType.NodeExecutionDetails, nodeExecution.id],
        queryFn: () => doFetchNodeExecutionDetails(queryClient, nodeExecution)
    });
}

export function useNodeExecutionDetails(nodeExecution?: NodeExecution) {
    const queryClient = useQueryClient();
    return useQuery<NodeExecutionDetails, Error>({
        enabled: !!nodeExecution,
        // Once we successfully map these details, we don't need to do it again.
        staleTime: Infinity,
        queryKey: [QueryType.NodeExecutionDetails, nodeExecution?.id],
        queryFn: () => doFetchNodeExecutionDetails(queryClient, nodeExecution!)
    });
}
