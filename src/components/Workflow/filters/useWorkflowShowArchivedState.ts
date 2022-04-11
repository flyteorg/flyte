import { useState } from 'react';
import { FilterOperation, FilterOperationName } from 'models/AdminEntity/types';
import { WorkflowExecutionState } from 'models/Workflow/enums';

interface ArchiveFilterState {
  showArchived: boolean;
  setShowArchived: (newValue: boolean) => void;
  getFilter: () => FilterOperation;
}

/**
 *  Allows to filter by Archive state
 */
export function useWorkflowShowArchivedState(): ArchiveFilterState {
  const [showArchived, setShowArchived] = useState(false);

  // By default all values are returned with NAMED_ENTITY_ACTIVE state
  const getFilter = (): FilterOperation => {
    return {
      key: 'state',
      operation: FilterOperationName.EQ,
      value: showArchived
        ? WorkflowExecutionState.NAMED_ENTITY_ARCHIVED
        : WorkflowExecutionState.NAMED_ENTITY_ACTIVE,
    };
  };

  return {
    showArchived,
    setShowArchived,
    getFilter,
  };
}
