import { unknownValueString } from 'common/constants';
import { dateWithFromNow, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';
import { DetailsGroup } from 'components/common';
import { TaskExecution } from 'models';
import * as React from 'react';

/** Renders the less important details for a `TaskExecution` as a `DetailsGroup`
 */
export const TaskExecutionDetails: React.FC<{
    taskExecution: TaskExecution;
}> = ({ taskExecution }) => {
    return (
        <section>
            <DetailsGroup
                labelWidthGridUnits={7}
                items={[
                    {
                        name: 'started',
                        content: taskExecution.closure.startedAt
                            ? dateWithFromNow(
                                  timestampToDate(
                                      taskExecution.closure.startedAt
                                  )
                              )
                            : unknownValueString
                    },
                    {
                        name: 'run time',
                        content: taskExecution.closure.duration
                            ? protobufDurationToHMS(
                                  taskExecution.closure.duration
                              )
                            : unknownValueString
                    }
                ]}
            />
        </section>
    );
};
