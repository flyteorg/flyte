import * as React from 'react';
import { makeStyles } from '@material-ui/core';
import { NodeExecutionIdentifier } from 'models/Execution/types';
import { DetailsPanel } from 'components/common/DetailsPanel';
import { useMemo, useState } from 'react';
import { NodeExecutionDetailsPanelContent } from '../NodeExecutionDetailsPanelContent';
import { NodeExecutionsTimelineContext } from './context';
import { ExecutionTimelineFooter } from './ExecutionTimelineFooter';
import { ExecutionTimeline } from './ExecutionTimeline';
import { TimeZone } from './helpers';
import { ScaleProvider } from './scaleContext';

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

export const ExecutionNodesTimeline = () => {
  const styles = useStyles();

  const [selectedExecution, setSelectedExecution] = useState<NodeExecutionIdentifier | null>(null);
  const [chartTimezone, setChartTimezone] = useState(TimeZone.Local);

  const onCloseDetailsPanel = () => setSelectedExecution(null);
  const handleTimezoneChange = (tz) => setChartTimezone(tz);

  const timelineContext = useMemo(
    () => ({ selectedExecution, setSelectedExecution }),
    [selectedExecution, setSelectedExecution],
  );

  return (
    <ScaleProvider>
      <div className={styles.wrapper}>
        <div className={styles.container}>
          <NodeExecutionsTimelineContext.Provider value={timelineContext}>
            <ExecutionTimeline chartTimezone={chartTimezone} />;
          </NodeExecutionsTimelineContext.Provider>
        </div>
        <ExecutionTimelineFooter onTimezoneChange={handleTimezoneChange} />
      </div>

      {/* Side panel, shows information for specific node */}
      <DetailsPanel open={selectedExecution !== null} onClose={onCloseDetailsPanel}>
        {selectedExecution && (
          <NodeExecutionDetailsPanelContent
            onClose={onCloseDetailsPanel}
            nodeExecutionId={selectedExecution}
          />
        )}
      </DetailsPanel>
    </ScaleProvider>
  );
};
