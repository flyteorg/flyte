import { WaitForData } from 'components/common/WaitForData';
import { useTaskNameList } from 'components/hooks/useNamedEntity';
import { SearchableTaskNameList } from 'components/Task/SearchableTaskNameList';
import { useTaskShowArchivedState } from 'components/Task/useTaskShowArchivedState';
import { limits } from 'models/AdminEntity/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { taskSortFields } from 'models/Task/constants';
import * as React from 'react';

export interface ProjectTasksProps {
  projectId: string;
  domainId: string;
}

const DEFAULT_SORT = {
  direction: SortDirection.ASCENDING,
  key: taskSortFields.name,
};

/** A listing of the Tasks registered for a project */
export const ProjectTasks: React.FC<ProjectTasksProps> = ({
  domainId: domain,
  projectId: project,
}) => {
  const archivedFilter = useTaskShowArchivedState();

  const taskNames = useTaskNameList(
    { domain, project },
    {
      limit: limits.NONE,
      sort: DEFAULT_SORT,
      filter: [archivedFilter.getFilter()],
    },
  );

  return (
    <WaitForData {...taskNames}>
      <SearchableTaskNameList
        names={taskNames.value}
        showArchived={archivedFilter.showArchived}
        onArchiveFilterChange={archivedFilter.setShowArchived}
      />
    </WaitForData>
  );
};
