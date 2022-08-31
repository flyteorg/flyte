import { makeStyles } from '@material-ui/core';
import { DetailsPanel } from 'components/common/DetailsPanel';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { DataError } from 'components/Errors/DataError';
import { makeWorkflowQuery } from 'components/Workflow/workflowQueries';
import { WorkflowGraph } from 'components/WorkflowGraph/WorkflowGraph';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { NodeExecutionIdentifier } from 'models/Execution/types';
import { startNodeId, endNodeId } from 'models/Node/constants';
import { Workflow } from 'models/Workflow/types';
import * as React from 'react';
import { useContext, useEffect, useMemo, useState } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionsByIdContext } from '../contexts';
import { tabs } from './constants';
import { NodeExecutionDetailsPanelContent } from './NodeExecutionDetailsPanelContent';
import { NodeExecutionsTimelineContext } from './Timeline/context';
import { ExecutionTimeline } from './Timeline/ExecutionTimeline';
import { ExecutionTimelineFooter } from './Timeline/ExecutionTimelineFooter';
import { TimeZone } from './Timeline/helpers';
import { ScaleProvider } from './Timeline/scaleContext';

export interface ExecutionTabProps {
  tabType: string;
}

const useStyles = makeStyles(() => ({
  wrapper: {
    display: 'flex',
    flexDirection: 'column',
    flex: '1 1 100%',
  },
  container: {
    display: 'flex',
    flex: '1 1 0',
    overflowY: 'auto',
  },
}));

/** Contains the available ways to visualize the nodes of a WorkflowExecution */
export const ExecutionTab: React.FC<ExecutionTabProps> = ({ tabType }) => {
  const styles = useStyles();
  const queryClient = useQueryClient();
  const { workflowId } = useNodeExecutionContext();
  const workflowQuery = useQuery<Workflow, Error>(makeWorkflowQuery(queryClient, workflowId));

  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);

  // Note: flytegraph allows multiple selection, but we only support showing
  // a single item in the details panel
  const [selectedExecution, setSelectedExecution] = useState<NodeExecutionIdentifier | null>(
    selectedNodes.length
      ? nodeExecutionsById[selectedNodes[0]]
        ? nodeExecutionsById[selectedNodes[0]].id
        : {
            nodeId: selectedNodes[0],
            executionId: nodeExecutionsById[Object.keys(nodeExecutionsById)[0]].id.executionId,
          }
      : null,
  );

  const [selectedPhase, setSelectedPhase] = useState<TaskExecutionPhase | undefined>(undefined);
  const [isDetailsTabClosed, setIsDetailsTabClosed] = useState<boolean>(!selectedExecution);

  useEffect(() => {
    setIsDetailsTabClosed(!selectedExecution);
  }, [selectedExecution]);

  const onNodeSelectionChanged = (newSelection: string[]) => {
    const validSelection = newSelection.filter((nodeId) => {
      if (nodeId === startNodeId || nodeId === endNodeId) {
        return false;
      }
      return true;
    });
    setSelectedNodes(validSelection);
    const newSelectedExecution = validSelection.length
      ? nodeExecutionsById[validSelection[0]]
        ? nodeExecutionsById[validSelection[0]].id
        : {
            nodeId: validSelection[0],
            executionId: nodeExecutionsById[Object.keys(nodeExecutionsById)[0]].id.executionId,
          }
      : null;
    setSelectedExecution(newSelectedExecution);
  };

  const onCloseDetailsPanel = () => {
    setSelectedExecution(null);
    setSelectedPhase(undefined);
    setSelectedNodes([]);
  };

  const [chartTimezone, setChartTimezone] = useState(TimeZone.Local);

  const handleTimezoneChange = (tz) => setChartTimezone(tz);

  const timelineContext = useMemo(
    () => ({ selectedExecution, setSelectedExecution }),
    [selectedExecution, setSelectedExecution],
  );

  const renderGraph = (workflow: Workflow) => (
    <WorkflowGraph
      onNodeSelectionChanged={onNodeSelectionChanged}
      selectedPhase={selectedPhase}
      onPhaseSelectionChanged={setSelectedPhase}
      isDetailsTabClosed={isDetailsTabClosed}
      workflow={workflow}
    />
  );

  return (
    <ScaleProvider>
      {tabType === tabs.timeline.id && (
        <div className={styles.wrapper}>
          <div className={styles.container}>
            <NodeExecutionsTimelineContext.Provider value={timelineContext}>
              <ExecutionTimeline chartTimezone={chartTimezone} />;
            </NodeExecutionsTimelineContext.Provider>
          </div>
          <ExecutionTimelineFooter onTimezoneChange={handleTimezoneChange} />
        </div>
      )}
      {tabType === tabs.graph.id && (
        <WaitForQuery errorComponent={DataError} query={workflowQuery}>
          {renderGraph}
        </WaitForQuery>
      )}
      {/* Side panel, shows information for specific node */}
      <DetailsPanel open={!isDetailsTabClosed} onClose={onCloseDetailsPanel}>
        {!isDetailsTabClosed && selectedExecution && (
          <NodeExecutionDetailsPanelContent
            onClose={onCloseDetailsPanel}
            phase={selectedPhase}
            nodeExecutionId={selectedExecution}
          />
        )}
      </DetailsPanel>
    </ScaleProvider>
  );
};
