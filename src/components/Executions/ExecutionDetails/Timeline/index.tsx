import * as React from 'react';
import { makeStyles } from '@material-ui/core';
import { NodeExecution, NodeExecutionIdentifier } from 'models/Execution/types';
import { WaitForQuery } from 'components/common/WaitForQuery';
import { NodeExecutionsRequestConfigContext } from 'components/Executions/contexts';
import { useAllTreeNodeExecutionGroupsQuery } from 'components/Executions/nodeExecutionQueries';
import { DataError } from 'components/Errors/DataError';
import { DetailsPanel } from 'components/common/DetailsPanel';
import { NodeExecutionDetailsPanelContent } from '../NodeExecutionDetailsPanelContent';
import { NodeExecutionsTimelineContext } from './context';
import { ExecutionTimelineFooter } from './ExecutionTimelineFooter';
import { ExecutionTimeline } from './ExecutionTimeline';
import { TimeZone } from './helpers';

const useStyles = makeStyles(() => ({
  wrapper: {
    display: 'flex',
    flexDirection: 'column',
    flex: '1 1 100%'
  },
  container: {
    display: 'flex',
    flex: '1 1 0',
    overflowY: 'auto'
  }
}));

interface TimelineProps {
  nodeExecutions: NodeExecution[];
}

export const ExecutionNodesTimeline = (props: TimelineProps) => {
  const styles = useStyles();

  const [selectedExecution, setSelectedExecution] = React.useState<NodeExecutionIdentifier | null>(null);
  const [chartTimeInterval, setChartTimeInterval] = React.useState(12); //narusina - should use 1-6 point system instead of real numbers
  const [chartTimezone, setChartTimezone] = React.useState(TimeZone.Local);

  const onCloseDetailsPanel = () => setSelectedExecution(null);
  const handleTimeIntervalChange = interval => setChartTimeInterval(interval);
  const handleTimezoneChange = tz => setChartTimezone(tz);

  const requestConfig = React.useContext(NodeExecutionsRequestConfigContext);
  const childGroupsQuery = useAllTreeNodeExecutionGroupsQuery(props.nodeExecutions, requestConfig);

  const timelineContext = React.useMemo(() => ({ selectedExecution, setSelectedExecution }), [
    selectedExecution,
    setSelectedExecution
  ]);

  const renderExecutionsTimeline = (nodeExecutions: NodeExecution[]) => {
    console.log(`!!! NODE: ${nodeExecutions.length}`);
    return (
      <ExecutionTimeline
        nodeExecutions={nodeExecutions}
        chartTimeInterval={chartTimeInterval}
        chartTimezone={chartTimezone}
      />
    );
  };

  return (
    <>
      <div className={styles.wrapper}>
        <div className={styles.container}>
          <NodeExecutionsTimelineContext.Provider value={timelineContext}>
            <WaitForQuery errorComponent={DataError} query={childGroupsQuery}>
              {renderExecutionsTimeline}
            </WaitForQuery>
          </NodeExecutionsTimelineContext.Provider>
        </div>
        <ExecutionTimelineFooter
          maxTime={120} // narusina - this time should depend on longest execution, currently always cupped to 2 min
          onTimeIntervalChange={handleTimeIntervalChange}
          onTimezoneChange={handleTimezoneChange}
        />
      </div>

      {/* Side panel, shows information for specific node */}
      <DetailsPanel open={selectedExecution !== null} onClose={onCloseDetailsPanel}>
        {selectedExecution && (
          <NodeExecutionDetailsPanelContent onClose={onCloseDetailsPanel} nodeExecutionId={selectedExecution} />
        )}
      </DetailsPanel>
    </>
  );
};
