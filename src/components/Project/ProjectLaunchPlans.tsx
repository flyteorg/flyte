import * as React from 'react';

import { SectionHeader, WaitForData, withRouteParams } from 'components/common';
import { useLaunchPlans } from 'components/hooks';
import { LaunchPlansTable } from 'components/Launch/LaunchPlansTable';

import { SortDirection } from 'models/AdminEntity';
import { launchSortFields } from 'models/Launch';

export interface ProjectLaunchPlansRouteParams {
    projectId: string;
    domainId: string;
}

/** The tab/page content for viewing a project's launch plans */
export const ProjectLaunchPlansContainer: React.FC<ProjectLaunchPlansRouteParams> = ({
    projectId: project,
    domainId: domain
}) => {
    const launchPlans = useLaunchPlans(
        { domain, project },
        {
            sort: {
                direction: SortDirection.DESCENDING,
                key: launchSortFields.createdAt
            }
        }
    );

    return (
        <>
            <SectionHeader title="Launch Plans" />
            <WaitForData {...launchPlans}>
                <LaunchPlansTable {...launchPlans} />
            </WaitForData>
        </>
    );
};

export const ProjectLaunchPlans = withRouteParams<
    ProjectLaunchPlansRouteParams
>(ProjectLaunchPlansContainer);
