import { WaitForData } from 'components/common/WaitForData';
import { useTaskNameList } from 'components/hooks/useNamedEntity';
import { SearchableTaskNameList } from 'components/Task/SearchableTaskNameList';
import { limits } from 'models/AdminEntity/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { taskSortFields } from 'models/Task/constants';
import * as React from 'react';

export interface ProjectTasksProps {
  projectId: string;
  domainId: string;
}

/** A listing of the Tasks registered for a project */
export const ProjectTasks: React.FC<ProjectTasksProps> = ({
  domainId: domain,
  projectId: project,
}) => {
  const taskNames = useTaskNameList(
    { domain, project },
    {
      limit: limits.NONE,
      sort: {
        direction: SortDirection.ASCENDING,
        key: taskSortFields.name,
      },
    },
  );

  return (
    <WaitForData {...taskNames}>
      <SearchableTaskNameList names={taskNames.value} />
    </WaitForData>
  );
};
