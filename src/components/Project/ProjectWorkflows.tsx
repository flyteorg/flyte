import { WaitForData } from 'components/common';
import { useWorkflowNameList } from 'components/hooks/useNamedEntity';
import { SearchableWorkflowNameList } from 'components/Workflow/SearchableWorkflowNameList';
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
    const workflowNames = useWorkflowNameList(
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
        <WaitForData {...workflowNames}>
            <SearchableWorkflowNameList workflowNames={workflowNames.value} />
        </WaitForData>
    );
};
