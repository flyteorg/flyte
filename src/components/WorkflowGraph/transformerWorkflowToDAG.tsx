import {
    DISPLAY_NAME_END,
    DISPLAY_NAME_START
} from 'components/flytegraph/ReactFlow/utils';
import { dTypes, dEdge, dNode } from 'models/Graph/types';
import { startNodeId, endNodeId } from 'models/Node/constants';
import { CompiledNode, ConnectionSet, TaskNode } from 'models/Node/types';
import { CompiledTask } from 'models/Task/types';
import {
    CompiledWorkflow,
    CompiledWorkflowClosure,
    WorkflowTemplate
} from 'models/Workflow/types';
import {
    isEndNode,
    isStartNode,
    getThenNodeFromBranch,
    getDisplayName,
    getSubWorkflowFromId,
    getNodeTypeFromCompiledNode,
    getTaskTypeFromCompiledNode
} from './utils';

/**
 * Returns a DAG from Flyte workflow request data
 * @param context input can be either CompiledWorkflow or CompiledNode
 * @returns Display name
 */
export const transformerWorkflowToDAG = (
    workflow: CompiledWorkflowClosure
): dNode => {
    const { primary } = workflow;
    const root = buildDAG(null, primary, dTypes.primary, workflow);
    return root;
};

const createDNode = (
    compiledNode: CompiledNode,
    parentDNode?: dNode | null,
    taskTemplate?: CompiledTask | null,
    typeOverride?: dTypes | null
): dNode => {
    const nodeValue =
        taskTemplate == null
            ? compiledNode
            : { ...compiledNode, ...taskTemplate };

    /* scopedId is used for requests; this creates format used by contract */
    let scopedId = '';

    if (parentDNode && parentDNode.type != dTypes.start) {
        if (
            parentDNode.type == dTypes.branch ||
            parentDNode.type == dTypes.subworkflow
        ) {
            /* Note: request contract indicates nested (subworkflow, branch) with -0- */
            scopedId = `${parentDNode.scopedId}-0-${compiledNode.id}`;
        } else {
            scopedId = `${parentDNode.scopedId}-${compiledNode.id}`;
        }
    } else {
        /* Case: primary workflow nodes won't have parents */
        scopedId = compiledNode.id;
    }

    /**
     * @TODO decide if we want to nested/standard start/end in
     *       UX; saving untilthat is decided.
     */
    const type =
        typeOverride == null
            ? getNodeTypeFromCompiledNode(compiledNode)
            : typeOverride;

    const output = {
        id: compiledNode.id,
        scopedId: scopedId,
        value: nodeValue,
        type: type,
        name: getDisplayName(compiledNode),
        nodes: [],
        edges: []
    } as dNode;
    return output;
};

export const buildBranchStartEndNodes = (root: dNode) => {
    const startNode = createDNode(
        {
            id: `${root.id}-${startNodeId}`,
            metadata: {
                name: DISPLAY_NAME_START
            }
        } as CompiledNode,
        null,
        null,
        dTypes.nestedStart
    );

    const endNode = createDNode(
        {
            id: `${root.id}-${endNodeId}`,
            metadata: {
                name: DISPLAY_NAME_END
            }
        } as CompiledNode,
        null,
        null,
        dTypes.nestedEnd
    );

    return {
        startNode,
        endNode
    };
};

/**
 * Will parse values when dealing with a Branch and recursively find and build
 * any other node types.
 * @param root      Parent root for Branch; will render independent DAG and
 *                  add as a child node of root.
 * @param parentCompiledNode   CompiledNode of origin
 */
export const parseBranch = (
    root: dNode,
    parentCompiledNode: CompiledNode,
    workflow: CompiledWorkflowClosure
) => {
    const thenNodeCompiledNode = getThenNodeFromBranch(parentCompiledNode);
    const thenNodeDNode = createDNode(thenNodeCompiledNode, root);
    const { startNode, endNode } = buildBranchStartEndNodes(root);

    /* We must push container node regardless */
    root.nodes.push(thenNodeDNode);

    if (thenNodeCompiledNode.branchNode) {
        buildDAG(thenNodeDNode, thenNodeCompiledNode, dTypes.branch, workflow);
    } else {
        /* Find any 'other' (other means 'else', 'else if') nodes */
        const otherArr = parentCompiledNode.branchNode?.ifElse?.other;

        if (otherArr) {
            otherArr.map(otherItem => {
                const otherCompiledNode: CompiledNode = otherItem.thenNode as CompiledNode;
                if (otherCompiledNode.branchNode) {
                    const otherDNodeBranch = createDNode(
                        otherCompiledNode,
                        root
                    );
                    buildDAG(
                        otherDNodeBranch,
                        otherCompiledNode,
                        dTypes.branch,
                        workflow
                    );
                } else {
                    const taskNode = otherCompiledNode.taskNode as TaskNode;
                    let taskType: CompiledTask | null = null;
                    if (taskNode) {
                        taskType = getTaskTypeFromCompiledNode(
                            taskNode,
                            workflow.tasks
                        ) as CompiledTask;
                    }
                    const otherDNode = createDNode(
                        otherCompiledNode,
                        root,
                        taskType
                    );
                    root.nodes.push(otherDNode);
                }
            });
        }
    }

    for (let i = 0; i < root.nodes.length; i++) {
        const startEdge: dEdge = {
            sourceId: startNode.id,
            targetId: root.nodes[i].scopedId
        };
        const endEdge: dEdge = {
            sourceId: root.nodes[i].scopedId,
            targetId: endNode.id
        };
        root.edges.push(startEdge);
        root.edges.push(endEdge);
    }

    /* Add back to root */
    root.nodes.push(startNode);
    root.nodes.push(endNode);
};

export const buildOutNodesFromContext = (
    root: dNode,
    contextWf: WorkflowTemplate,
    type: dTypes,
    workflow: CompiledWorkflowClosure
): void => {
    for (let i = 0; i < contextWf.nodes.length; i++) {
        const compiledNode: CompiledNode = contextWf.nodes[i];
        let dNode: dNode;

        if (isStartNode(compiledNode) && type == dTypes.subworkflow) {
            /** @TODO Decide if we should implement this */
            /* Case: override type as nestedStart node */
            dNode = createDNode(compiledNode);
        } else if (isEndNode(compiledNode) && type == dTypes.subworkflow) {
            /** @TODO Decide if we should implement this */
            /* Case: override type as nestedEnd node */
            dNode = createDNode(compiledNode);
        } else if (compiledNode.branchNode) {
            /* Case: recurse on branch node */
            dNode = createDNode(compiledNode, null);
            buildDAG(dNode, compiledNode, dTypes.branch, workflow);
        } else if (compiledNode.workflowNode) {
            /* Case: recurse on workflow node */
            const id = compiledNode.workflowNode.subWorkflowRef;
            const subworkflow = getSubWorkflowFromId(id, workflow);
            if (!isStartNode(root)) {
                dNode = createDNode(compiledNode, root);
            } else {
                /**
                 * @TODO may not need this else case
                 */
                dNode = createDNode(compiledNode, null);
            }
            buildDAG(dNode, subworkflow, dTypes.subworkflow, workflow);
        } else if (compiledNode.taskNode) {
            /* Case: build task node */
            const taskType = getTaskTypeFromCompiledNode(
                compiledNode.taskNode,
                workflow.tasks
            );
            dNode = createDNode(compiledNode, root, taskType);
        } else {
            /* Else: primary start/finish nodes */
            dNode = createDNode(compiledNode, root);
        }

        root.nodes.push(dNode);
    }
};

export const buildOutWorkflowEdges = (
    root,
    context: ConnectionSet,
    ingress,
    nodeMap
) => {
    const list = context.downstream[ingress].ids;
    for (let i = 0; i < list.length; i++) {
        const edge: dEdge = {
            sourceId: nodeMap[ingress].dNode.scopedId,
            targetId: nodeMap[list[i]].dNode.scopedId
        };
        root.edges.push(edge);
        if (context.downstream[list[i]]) {
            buildOutWorkflowEdges(root, context, list[i], nodeMap);
        }
    }
};

/**
 * Handles parsing CompiledWorkflow data objects
 *
 * @param root          Root node for the graph that will be rendered
 * @param context       The current workflow (could be child of main workflow)
 * @param type          Type (sub or primrary)
 * @param workflow      Main parent workflow
 */
export const parseWorkflow = (
    root,
    context: CompiledWorkflow,
    type: dTypes,
    workflow: CompiledWorkflowClosure
) => {
    /* Note: only Primary workflow is null, all others have root */
    let contextualRoot;
    if (root) {
        contextualRoot = root;
    } else {
        const primaryStart = createDNode({ id: startNodeId } as CompiledNode);
        contextualRoot = primaryStart;
    }

    /* Build Nodes */
    buildOutNodesFromContext(contextualRoot, context.template, type, workflow);

    const nodesList = context.template.nodes;
    const nodeMap = {};

    /* Create mapping of id => dNode for all child nodes of root to build edges */
    for (let i = 0; i < contextualRoot.nodes.length; i++) {
        const dNode = contextualRoot.nodes[i];
        nodeMap[dNode.id] = {
            dNode: dNode,
            compiledNode: nodesList[i]
        };
    }

    /* Build Edges */
    buildOutWorkflowEdges(
        contextualRoot,
        context.connections,
        startNodeId,
        nodeMap
    );
    return contextualRoot;
};

/**
 * Mutates root (if passed) by recursively rendering DAG of given context.
 *
 * @param root          Root node of DAG
 * @param graphType     DAG type (eg, branch, workflow)
 * @param context       Pointer to current context of response
 */
export const buildDAG = (
    root: dNode | null,
    context: any,
    graphType: dTypes,
    workflow: CompiledWorkflowClosure
) => {
    switch (graphType) {
        case dTypes.branch:
            parseBranch(root as dNode, context, workflow);
            break;
        case dTypes.subworkflow:
            parseWorkflow(root, context, graphType, workflow);
            break;
        case dTypes.primary:
            return parseWorkflow(root, context, graphType, workflow);
            break;
    }
};
