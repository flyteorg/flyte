import { WaitForData } from 'components/common/WaitForData';
import { useWorkflowShowArchivedState } from 'components/Workflow/filters/useWorkflowShowArchivedState';
import { SearchableWorkflowNameList } from 'components/Workflow/SearchableWorkflowNameList';
import { limits } from 'models/AdminEntity/constants';
import { SortDirection } from 'models/AdminEntity/types';
import { workflowSortFields } from 'models/Workflow/constants';
import * as React from 'react';
import { useWorkflowInfoList } from '../Workflow/useWorkflowInfoList';

export interface ProjectWorkflowsProps {
  projectId: string;
  domainId: string;
}

const DEFAULT_SORT = {
  direction: SortDirection.ASCENDING,
  key: workflowSortFields.name,
};

/** A listing of the Workflows registered for a project */
export const ProjectWorkflows: React.FC<ProjectWorkflowsProps> = ({
  domainId: domain,
  projectId: project,
}) => {
  const archivedFilter = useWorkflowShowArchivedState();
  const workflows = useWorkflowInfoList(
    { domain, project },
    {
      limit: limits.NONE,
      sort: DEFAULT_SORT,
      filter: [archivedFilter.getFilter()],
    },
  );

  return (
    <WaitForData {...workflows}>
      <SearchableWorkflowNameList
        workflows={workflows.value}
        showArchived={archivedFilter.showArchived}
        onArchiveFilterChange={archivedFilter.setShowArchived}
      />
    </WaitForData>
  );
};
