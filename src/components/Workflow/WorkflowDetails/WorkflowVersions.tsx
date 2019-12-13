import * as React from 'react';

import { WaitForData, withRouteParams } from 'components/common';
import { useWorkflows } from 'components/hooks';

import { SortDirection, workflowSortFields } from 'models';

import { WorkflowsTable } from '../WorkflowsTable';

export interface WorkflowVersionsRouteParams {
    projectId: string;
    domainId: string;
    workflowName: string;
}

/** The tab/page content for viewing a workflow's versions */
export const WorkflowVersionsContainer: React.FC<WorkflowVersionsRouteParams> = ({
    projectId: project,
    domainId: domain,
    workflowName: name
}) => {
    const workflows = useWorkflows(
        { domain, name, project },
        {
            sort: {
                direction: SortDirection.DESCENDING,
                key: workflowSortFields.createdAt
            }
        }
    );

    return (
        <WaitForData {...workflows}>
            <WorkflowsTable {...workflows} />
        </WaitForData>
    );
};

export const WorkflowVersions = withRouteParams<WorkflowVersionsRouteParams>(
    WorkflowVersionsContainer
);
