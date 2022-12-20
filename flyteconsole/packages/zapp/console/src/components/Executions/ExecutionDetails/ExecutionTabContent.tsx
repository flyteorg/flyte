import { makeStyles } from '@material-ui/core';
import { DetailsPanel } from 'components/common/DetailsPanel';
import { makeNodeExecutionDynamicWorkflowQuery } from 'components/Workflow/workflowQueries';
import { WorkflowGraph } from 'components/WorkflowGraph/WorkflowGraph';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { NodeExecution, NodeExecutionIdentifier } from 'models/Execution/types';
import { startNodeId, endNodeId } from 'models/Node/constants';
import * as React from 'react';
import { transformerWorkflowToDag } from 'components/WorkflowGraph/transformerWorkflowToDag';
import { checkForDynamicExecutions } from 'components/common/utils';
import { dNode } from 'models/Graph/types';
import { useContext, useEffect, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import {
  FilterOperation,
  FilterOperationName,
  FilterOperationValueList,
} from 'models/AdminEntity/types';
import { isEqual } from 'lodash';
import { useNodeExecutionContext } from '../contextProvider/NodeExecutionDetails';
import { NodeExecutionsByIdContext } from '../contexts';
import { NodeExecutionsTable } from '../Tables/NodeExecutionsTable';
import { tabs } from './constants';
import { NodeExecutionDetailsPanelContent } from './NodeExecutionDetailsPanelContent';
import { ExecutionTimeline } from './Timeline/ExecutionTimeline';
import { ExecutionTimelineFooter } from './Timeline/ExecutionTimelineFooter';
import { convertToPlainNodes, TimeZone } from './Timeline/helpers';
import { DetailsPanelContext } from './DetailsPanelContext';
import { useNodeExecutionFiltersState } from '../filters/useExecutionFiltersState';
import { nodeExecutionPhaseConstants } from '../constants';

interface ExecutionTabContentProps {
  tabType: string;
  filteredNodeExecutions?: NodeExecution[];
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

const executionMatchesPhaseFilter = (
  nodeExecution: NodeExecution,
  { key, value, operation }: FilterOperation,
) => {
  if (key === 'phase' && operation === FilterOperationName.VALUE_IN) {
    // default to UNKNOWN phase if the field does not exist on a closure
    const itemValue =
      nodeExecutionPhaseConstants[nodeExecution?.closure[key]]?.value ??
      nodeExecutionPhaseConstants[0].value;
    // phase check filters always return values in an array
    const valuesArray = value as FilterOperationValueList;
    return valuesArray.includes(itemValue);
  }
  return false;
};

export const ExecutionTabContent: React.FC<ExecutionTabContentProps> = ({
  tabType,
  filteredNodeExecutions,
}) => {
  const styles = useStyles();
  const { compiledWorkflowClosure } = useNodeExecutionContext();
  const { appliedFilters } = useNodeExecutionFiltersState();
  const nodeExecutionsById = useContext(NodeExecutionsByIdContext);

  const { dag, staticExecutionIdsMap, error } = compiledWorkflowClosure
    ? transformerWorkflowToDag(compiledWorkflowClosure)
    : { dag: {}, staticExecutionIdsMap: {}, error: null };
  const dynamicParents = checkForDynamicExecutions(nodeExecutionsById, staticExecutionIdsMap);
  const { data: dynamicWorkflows } = useQuery(
    makeNodeExecutionDynamicWorkflowQuery(dynamicParents),
  );
  const [initialNodes, setInitialNodes] = useState<dNode[]>([]);
  const [initialFilteredNodes, setInitialFilteredNodes] = useState<dNode[] | undefined>(undefined);
  const [mergedDag, setMergedDag] = useState(null);
  const [filters, setFilters] = useState<FilterOperation[]>(appliedFilters);
  const [isFiltersChanged, setIsFiltersChanged] = useState<boolean>(false);

  useEffect(() => {
    const nodes: dNode[] = compiledWorkflowClosure
      ? transformerWorkflowToDag(compiledWorkflowClosure, dynamicWorkflows).dag.nodes
      : [];
    // we remove start/end node info in the root dNode list during first assignment
    const plainNodes = convertToPlainNodes(nodes);

    let newMergedDag = dag;

    for (const dynamicId in dynamicWorkflows) {
      if (staticExecutionIdsMap[dynamicId]) {
        if (compiledWorkflowClosure) {
          const dynamicWorkflow = transformerWorkflowToDag(
            compiledWorkflowClosure,
            dynamicWorkflows,
          );
          newMergedDag = dynamicWorkflow.dag;
        }
      }
    }
    setMergedDag(newMergedDag);
    setInitialNodes(plainNodes);
  }, [compiledWorkflowClosure, dynamicWorkflows]);

  useEffect(() => {
    if (!isEqual(filters, appliedFilters)) {
      setFilters(appliedFilters);
      setIsFiltersChanged(true);
    } else {
      setIsFiltersChanged(false);
    }
  }, [appliedFilters]);

  useEffect(() => {
    if (appliedFilters.length > 0) {
      // if filter was apllied, but filteredNodeExecutions is empty, we only appliied Phase filter,
      // and need to clear out items manually
      if (!filteredNodeExecutions) {
        const filteredNodes = initialNodes.filter((node) =>
          executionMatchesPhaseFilter(nodeExecutionsById[node.scopedId], appliedFilters[0]),
        );
        setInitialFilteredNodes(filteredNodes);
      } else {
        const filteredNodes = initialNodes.filter((node: dNode) =>
          filteredNodeExecutions.find(
            (execution: NodeExecution) => execution.scopedId === node.scopedId,
          ),
        );
        setInitialFilteredNodes(filteredNodes);
      }
    }
  }, [initialNodes, filteredNodeExecutions, isFiltersChanged]);

  const [selectedNodes, setSelectedNodes] = useState<string[]>([]);

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

  const onCloseDetailsPanel = () => {
    setSelectedExecution(null);
    setSelectedPhase(undefined);
    setSelectedNodes([]);
  };

  const [chartTimezone, setChartTimezone] = useState(TimeZone.Local);

  const handleTimezoneChange = (tz) => setChartTimezone(tz);

  const detailsPanelContext = useMemo(
    () => ({ selectedExecution, setSelectedExecution }),
    [selectedExecution, setSelectedExecution],
  );

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

  const renderContent = () => {
    switch (tabType) {
      case tabs.nodes.id:
        return (
          <NodeExecutionsTable initialNodes={initialNodes} filteredNodes={initialFilteredNodes} />
        );
      case tabs.graph.id:
        return (
          <WorkflowGraph
            mergedDag={mergedDag}
            error={error}
            dynamicWorkflows={dynamicWorkflows}
            initialNodes={initialNodes}
            onNodeSelectionChanged={onNodeSelectionChanged}
            selectedPhase={selectedPhase}
            onPhaseSelectionChanged={setSelectedPhase}
            isDetailsTabClosed={isDetailsTabClosed}
          />
        );
      case tabs.timeline.id:
        return (
          <div className={styles.wrapper}>
            <div className={styles.container}>
              <ExecutionTimeline chartTimezone={chartTimezone} initialNodes={initialNodes} />
            </div>
            <ExecutionTimelineFooter onTimezoneChange={handleTimezoneChange} />
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <>
      <DetailsPanelContext.Provider value={detailsPanelContext}>
        {renderContent()}
      </DetailsPanelContext.Provider>
      {/* Side panel, shows information for specific node */}
      <DetailsPanel open={!isDetailsTabClosed} onClose={onCloseDetailsPanel}>
        {!isDetailsTabClosed && selectedExecution && (
          <NodeExecutionDetailsPanelContent
            onClose={onCloseDetailsPanel}
            taskPhase={selectedPhase ?? TaskExecutionPhase.UNDEFINED}
            nodeExecutionId={selectedExecution}
          />
        )}
      </DetailsPanel>
    </>
  );
};
