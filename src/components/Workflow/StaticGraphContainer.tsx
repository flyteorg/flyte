import * as React from 'react';
import { Workflow, WorkflowId } from 'models/Workflow/types';
import { useQuery, useQueryClient } from 'react-query';
import { makeWorkflowQuery } from './workflowQueries';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { transformerWorkflowToDag } from 'components/WorkflowGraph/transformerWorkflowToDag';
import { ReactFlowWrapper } from 'components/flytegraph/ReactFlow/ReactFlowWrapper';
import { ConvertFlyteDagToReactFlows } from 'components/flytegraph/ReactFlow/transformDAGToReactFlowV2';
import { getRFBackground } from 'components/flytegraph/ReactFlow/utils';
import { ConvertDagProps, RFWrapperProps } from 'components/flytegraph/ReactFlow/types';

export const renderStaticGraph = props => {
  const workflow = props.closure.compiledWorkflow;
  const { dag } = transformerWorkflowToDag(workflow);
  const rfGraphJson = ConvertFlyteDagToReactFlows({
    root: dag,
    maxRenderDepth: 0,
    currentNestedView: [],
    isStaticGraph: true
  } as ConvertDagProps);

  const backgroundStyle = getRFBackground().static;
  const ReactFlowProps: RFWrapperProps = {
    backgroundStyle,
    rfGraphJson: rfGraphJson,
    currentNestedView: []
  };
  return <ReactFlowWrapper {...ReactFlowProps} />;
};

export interface StaticGraphContainerProps {
  workflowId: WorkflowId;
}

export const StaticGraphContainer: React.FC<StaticGraphContainerProps> = ({ workflowId }) => {
  const containerStyle: React.CSSProperties = {
    display: 'flex',
    width: '100%'
  };
  const workflowQuery = useQuery<Workflow, Error>(makeWorkflowQuery(useQueryClient(), workflowId));

  return (
    <div style={containerStyle}>
      <WaitForQuery query={workflowQuery} errorComponent={DataError}>
        {renderStaticGraph}
      </WaitForQuery>
    </div>
  );
};
