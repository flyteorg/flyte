import * as React from 'react';

import { WaitForData, withRouteParams } from 'components/common';
import { useLaunchPlans } from 'components/hooks';
import { LaunchPlansTable } from 'components/Launch/LaunchPlansTable';

import { launchSortFields, SortDirection } from 'models';

export interface WorkflowLaunchPlansRouteParams {
    projectId: string;
    domainId: string;
    workflowName: string;
}

/** The tab/page content for viewing a workflow's launch plans */
export const WorkflowLaunchPlansContainer: React.FC<WorkflowLaunchPlansRouteParams> = ({
    projectId: project,
    domainId: domain,
    workflowName: name
}) => {
    const launchPlans = useLaunchPlans(
        { domain, name, project },
        {
            sort: {
                direction: SortDirection.DESCENDING,
                key: launchSortFields.createdAt
            }
        }
    );

    return (
        <WaitForData {...launchPlans}>
            <LaunchPlansTable {...launchPlans} />
        </WaitForData>
    );
};

export const WorkflowLaunchPlans = withRouteParams<
    WorkflowLaunchPlansRouteParams
>(WorkflowLaunchPlansContainer);
