import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { makeWorkflowQuery } from 'components/Workflow/workflowQueries';
import { Workflow } from 'models/Workflow/types';
import * as React from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { NodeExecution } from 'models/Execution/types';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { ScaleProvider } from './Timeline/scaleContext';
import { ExecutionTabContent } from './ExecutionTabContent';

interface ExecutionTabProps {
  tabType: string;
  filteredNodeExecutions?: NodeExecution[];
}

/** Contains the available ways to visualize the nodes of a WorkflowExecution */
export const ExecutionTab: React.FC<ExecutionTabProps> = ({ tabType, filteredNodeExecutions }) => {
  const queryClient = useQueryClient();
  const { workflowId } = useNodeExecutionContext();
  const workflowQuery = useQuery<Workflow, Error>(makeWorkflowQuery(queryClient, workflowId));

  return (
    <ScaleProvider>
      <WaitForQuery errorComponent={DataError} query={workflowQuery}>
        {() => (
          <ExecutionTabContent tabType={tabType} filteredNodeExecutions={filteredNodeExecutions} />
        )}
      </WaitForQuery>
    </ScaleProvider>
  );
};
