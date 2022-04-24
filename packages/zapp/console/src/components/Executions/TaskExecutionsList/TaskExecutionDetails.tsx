import { unknownValueString } from 'common/constants';
import { dateWithFromNow, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { DetailsGroup } from 'components/common/DetailsGroup';
import { TaskExecution } from 'models/Execution/types';
import * as React from 'react';

/** Renders the less important details for a `TaskExecution` as a `DetailsGroup`
 */
export const TaskExecutionDetails: React.FC<{
  taskExecution: TaskExecution;
}> = ({ taskExecution }) => {
  const labelWidthGridUnits = taskExecution.closure.startedAt ? 7 : 10;
  const detailItems = React.useMemo(() => {
    if (taskExecution.closure.startedAt) {
      return [
        {
          name: 'started',
          content: dateWithFromNow(timestampToDate(taskExecution.closure.startedAt)),
        },
        {
          name: 'run time',
          content: taskExecution.closure.duration
            ? protobufDurationToHMS(taskExecution.closure.duration)
            : unknownValueString,
        },
      ];
    } else {
      return [
        {
          name: 'last updated',
          content: taskExecution.closure.updatedAt
            ? dateWithFromNow(timestampToDate(taskExecution.closure.updatedAt))
            : unknownValueString,
        },
      ];
    }
  }, [taskExecution]);

  return (
    <section>
      <DetailsGroup labelWidthGridUnits={labelWidthGridUnits} items={detailItems} />
    </section>
  );
};
