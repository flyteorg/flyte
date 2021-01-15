import { NonIdealState } from 'components/common/NonIdealState';
import { Graph } from 'components/flytegraph/Graph';
import { NodeRenderer } from 'components/flytegraph/types';
import { keyBy } from 'lodash';
import { convertFlyteGraphToDAG } from 'models/Graph/convertFlyteGraphToDAG';
import { DAGNode } from 'models/Graph/types';
import { CompiledNode } from 'models/Node/types';
import { Workflow } from 'models/Workflow/types';
import * as React from 'react';

export interface WorkflowGraphProps {
    nodeRenderer: NodeRenderer<DAGNode>;
    onNodeSelectionChanged: (selectedNodes: string[]) => void;
    selectedNodes?: string[];
    workflow: Workflow;
}

interface WorkflowGraphState {
    dag: DAGNode[];
    nodesById: Record<string, CompiledNode>;
    error?: Error;
}

interface PrepareDAGResult {
    dag: DAGNode[];
    error?: Error;
    nodesById: Record<string, CompiledNode>;
}

function workflowToDag(workflow: Workflow): PrepareDAGResult {
    try {
        if (!workflow.closure) {
            throw new Error('Workflow has no closure');
        }
        if (!workflow.closure.compiledWorkflow) {
            throw new Error('Workflow closure missing a compiled workflow');
        }
        const { compiledWorkflow } = workflow.closure;
        const nodesById = keyBy(compiledWorkflow.primary.template.nodes, 'id');
        const dag = convertFlyteGraphToDAG(compiledWorkflow);
        return { dag, nodesById };
    } catch (e) {
        return {
            dag: [],
            error: e,
            nodesById: {}
        };
    }
}

/** Uses flytegraph to render a graph representation of a workflow closure,
 * pipes node details about selected nodes into the DetailsPanel
 */
export class WorkflowGraph extends React.Component<
    WorkflowGraphProps,
    WorkflowGraphState
> {
    constructor(props: WorkflowGraphProps) {
        super(props);
        const { dag, error, nodesById } = workflowToDag(this.props.workflow);
        this.state = { dag, error, nodesById };
    }

    render() {
        const { dag, error } = this.state;
        const {
            nodeRenderer,
            onNodeSelectionChanged,
            selectedNodes
        } = this.props;
        if (error) {
            return (
                <NonIdealState
                    title="Cannot render Workflow graph"
                    description={error.message}
                />
            );
        }

        return (
            <>
                <Graph
                    data={dag}
                    nodeRenderer={nodeRenderer}
                    onNodeSelectionChanged={onNodeSelectionChanged}
                    selectedNodes={selectedNodes}
                />
            </>
        );
    }
}
