import * as React from 'react';

import { SectionHeader, WaitForData, withRouteParams } from 'components/common';
import { WorkflowExecutionsTable } from 'components/Executions/Tables/WorkflowExecutionsTable';
import { useWorkflowExecutions } from 'components/hooks';
import { SortDirection } from 'models/AdminEntity';
import { executionSortFields } from 'models/Execution';

export interface ProjectExecutionsRouteParams {
    projectId: string;
    domainId: string;
}

/** The tab/page content for viewing a project's executions */
export const ProjectExecutionsContainer: React.FC<ProjectExecutionsRouteParams> = ({
    projectId: project,
    domainId: domain
}) => {
    const sort = {
        key: executionSortFields.createdAt,
        direction: SortDirection.DESCENDING
    };
    const executions = useWorkflowExecutions({ domain, project }, { sort });

    return (
        <>
            <SectionHeader title="Executions" />
            <WaitForData {...executions}>
                <WorkflowExecutionsTable {...executions} />
            </WaitForData>
        </>
    );
};

export const ProjectExecutions = withRouteParams<ProjectExecutionsRouteParams>(
    ProjectExecutionsContainer
);
