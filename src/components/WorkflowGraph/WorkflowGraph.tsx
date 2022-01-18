import { transformerWorkflowToDag } from './transformerWorkflowToDag';
import { dNode } from 'models/Graph/types';
import { Workflow } from 'models/Workflow/types';
import * as React from 'react';
import ReactFlowGraphComponent from 'components/flytegraph/ReactFlow/ReactFlowGraphComponent';
import { Error } from 'models/Common/types';

export interface WorkflowGraphProps {
    onNodeSelectionChanged: (selectedNodes: string[]) => void;
    selectedNodes?: string[];
    workflow: Workflow;
    nodeExecutionsById?: any;
}

interface WorkflowGraphState {
    dag: dNode | null;
    error?: Error;
}

interface PrepareDAGResult {
    dag: dNode | null;
    error?: Error;
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
        const dag: dNode = transformerWorkflowToDag(compiledWorkflow);
        return { dag };
    } catch (e) {
        return {
            dag: null,
            error: e as Error
        };
    }
}

export class WorkflowGraph extends React.Component<
    WorkflowGraphProps,
    WorkflowGraphState
> {
    constructor(props) {
        super(props);
        const { dag, error } = workflowToDag(this.props.workflow);
        this.state = { dag, error };
    }

    render() {
        const { dag } = this.state;
        const { onNodeSelectionChanged, nodeExecutionsById } = this.props;

        return (
            <ReactFlowGraphComponent
                nodeExecutionsById={nodeExecutionsById}
                data={dag}
                onNodeSelectionChanged={onNodeSelectionChanged}
            />
        );
    }
}
