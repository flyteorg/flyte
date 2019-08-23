import { WaitForData } from 'components/common';
import { useWorkflowIds } from 'components/hooks';
import { SearchableWorkflowIdList } from 'components/Workflow/SearchableWorkflowIdList';
import { limits, SortDirection, workflowSortFields } from 'models';
import * as React from 'react';

export interface ProjectWorkflowsProps {
    projectId: string;
    domainId: string;
}

/** A listing of the Workflows registered for a project */
export const ProjectWorkflows: React.FC<ProjectWorkflowsProps> = ({
    domainId: domain,
    projectId: project
}) => {
    const workflowIds = useWorkflowIds(
        { domain, project },
        {
            limit: limits.NONE,
            sort: {
                direction: SortDirection.ASCENDING,
                key: workflowSortFields.name
            }
        }
    );

    return (
        <WaitForData {...workflowIds}>
            <SearchableWorkflowIdList workflowIds={workflowIds.value} />
        </WaitForData>
    );
};
