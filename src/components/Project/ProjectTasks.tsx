import { WaitForData } from 'components/common';
import { useTaskNameList } from 'components/hooks/useNamedEntity';
import { SearchableTaskNameList } from 'components/Task/SearchableTaskNameList';
import { limits, SortDirection, workflowSortFields } from 'models';
import * as React from 'react';

export interface ProjectTasksProps {
    projectId: string;
    domainId: string;
}

/** A listing of the Tasks registered for a project */
export const ProjectTasks: React.FC<ProjectTasksProps> = ({
    domainId: domain,
    projectId: project
}) => {
    const taskNames = useTaskNameList(
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
        <WaitForData {...taskNames}>
            <SearchableTaskNameList names={taskNames.value} />
        </WaitForData>
    );
};
