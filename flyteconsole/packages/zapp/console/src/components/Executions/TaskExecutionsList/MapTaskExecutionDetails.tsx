import * as React from 'react';
import { MapTaskExecution } from 'models/Execution/types';
import { TaskExecutionPhase } from 'models/Execution/enums';
import { PanelSection } from 'components/common/PanelSection';
import { formatRetryAttempt, getTaskRetryAtemptsForIndex } from './utils';
import { TaskExecutionLogsCard } from './TaskExecutionLogsCard';

interface MapTaskExecutionDetailsProps {
  taskExecution: MapTaskExecution;
}

/** Renders an individual map task execution attempts as part of a list */
export const MapTaskExecutionDetails: React.FC<MapTaskExecutionDetailsProps> = ({
  taskExecution,
}) => {
  const {
    closure: { metadata },
    taskIndex,
  } = taskExecution;

  const filteredResources = getTaskRetryAtemptsForIndex(
    metadata?.externalResources ?? [],
    taskIndex,
  );

  return (
    <PanelSection>
      {filteredResources.map((item) => {
        const attempt = item.retryAttempt ?? 0;
        const headerText = formatRetryAttempt(attempt);

        return (
          <div key={`card-${attempt}`}>
            <TaskExecutionLogsCard
              taskExecution={taskExecution}
              headerText={headerText}
              phase={item.phase ?? TaskExecutionPhase.UNDEFINED}
              logs={item.logs ?? []}
            />
          </div>
        );
      })}
    </PanelSection>
  );
};
