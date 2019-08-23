import * as React from 'react';

import { dateWithFromNow, protobufDurationToHMS } from 'common/formatters';
import { timestampToDate } from 'common/utils';

import { Execution } from 'models';

import { DetailsGroup } from 'components/common';

export const ExecutionExtendedDetails: React.FC<{ execution: Execution }> = ({
    execution
}) => {
    return (
        <DetailsGroup
            items={[
                {
                    name: 'started',
                    content: execution.closure.startedAt
                        ? dateWithFromNow(
                              timestampToDate(execution.closure.startedAt)
                          )
                        : ''
                },
                {
                    name: 'run time',
                    content: execution.closure.duration
                        ? protobufDurationToHMS(execution.closure.duration)
                        : ''
                },
                {
                    name: 'workflow version',
                    content: `${execution.spec.launchPlan.version}`
                }
            ]}
        />
    );
};
