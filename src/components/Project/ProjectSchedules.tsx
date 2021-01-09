import { SectionHeader, WaitForData, withRouteParams } from 'components/common';
import { useLaunchPlans } from 'components/hooks';
import { isLoadingState } from 'components/hooks/fetchMachine';
import { SchedulesTable } from 'components/Launch/SchedulesTable';
import { SortDirection } from 'models/AdminEntity';
import { launchSortFields } from 'models/Launch';
import * as React from 'react';

export interface ProjectSchedulesRouteParams {
    projectId: string;
    domainId: string;
}

/** The tab/page content for viewing a project's schedules */
export const ProjectSchedulesContainer: React.FC<ProjectSchedulesRouteParams> = ({
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
            <SectionHeader title="Schedules" />
            <WaitForData {...launchPlans}>
                <SchedulesTable
                    {...launchPlans}
                    isFetching={isLoadingState(launchPlans.state)}
                />
            </WaitForData>
        </>
    );
};

export const ProjectSchedules = withRouteParams<ProjectSchedulesRouteParams>(
    ProjectSchedulesContainer
);
