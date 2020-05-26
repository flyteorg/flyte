import { WaitForData } from 'components/common';
import { useWorkflowNameList } from 'components/hooks/useNamedEntity';
import { SearchableWorkflowNameList } from 'components/Workflow/SearchableWorkflowNameList';
import { Admin } from 'flyteidl';
import {
    FilterOperationName,
    limits,
    SortDirection,
    workflowSortFields
} from 'models';
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
            },
            // Hide archived workflows from the list
            filter: [
                {
                    key: 'state',
                    operation: FilterOperationName.EQ,
                    value: Admin.NamedEntityState.NAMED_ENTITY_ACTIVE
                }
            ]
        }
    );

    return (
        <WaitForData {...workflowNames}>
            <SearchableWorkflowNameList names={workflowNames.value} />
        </WaitForData>
    );
};
