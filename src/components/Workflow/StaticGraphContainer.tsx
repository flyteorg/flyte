import * as React from 'react';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import { useQuery, useQueryClient } from 'react-query';
import { makeWorkflowQuery } from './workflowQueries';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { transformerWorkflowToDAG } from 'components/WorkflowGraph/transformerWorkflowToDAG';
import { ReactFlowWrapper } from 'components/flytegraph/ReactFlow/ReactFlowWrapper';
import { ConvertFlyteDagToReactFlows } from 'components/flytegraph/ReactFlow/transformerDAGToReactFlow';
import { dNode } from 'models/Graph/types';
import { getRFBackground } from 'components/flytegraph/ReactFlow/utils';
import {
    ConvertDagProps,
    RFGraphTypes,
    RFWrapperProps
} from 'components/flytegraph/ReactFlow/types';

export const renderStaticGraph = props => {
    const workflow = props.closure.compiledWorkflow;
    const version = props.id.version;

    const dag: dNode = transformerWorkflowToDAG(workflow);
    const rfGraphJson = ConvertFlyteDagToReactFlows({
        root: dag,
        maxRenderDepth: 0,
        isStaticGraph: true
    } as ConvertDagProps);
    const backgroundStyle = getRFBackground().static;
    const ReactFlowProps: RFWrapperProps = {
        backgroundStyle,
        rfGraphJson: rfGraphJson,
        type: RFGraphTypes.static,
        version: version
    };
    return <ReactFlowWrapper {...ReactFlowProps} />;
};

export interface StaticGraphContainerProps {
    workflowId: WorkflowId;
}

export const StaticGraphContainer: React.FC<StaticGraphContainerProps> = ({
    workflowId
}) => {
    const containerStyle = {
        width: '100%',
        height: '30%',
        maxHeight: '400px',
        minHeight: '220px'
    };
    const workflowQuery = useQuery<Workflow, Error>(
        makeWorkflowQuery(useQueryClient(), workflowId)
    );

    return (
        <div style={containerStyle}>
            <WaitForQuery query={workflowQuery} errorComponent={DataError}>
                {renderStaticGraph}
            </WaitForQuery>
        </div>
    );
};
