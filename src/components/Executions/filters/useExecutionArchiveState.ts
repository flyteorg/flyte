import { useState } from 'react';
import { FilterOperation, FilterOperationName } from 'models/AdminEntity/types';
import { ExecutionState } from 'models/Execution/enums';

interface ArchiveFilterState {
    showArchived: boolean;
    setShowArchived: (newValue: boolean) => void;
    getFilter: () => FilterOperation | null;
}

/**
 *  Allows to filter by Archive state
 */
export function useExecutionShowArchivedState(): ArchiveFilterState {
    const [showArchived, setShowArchived] = useState(false);

    // By default all values are returned with EXECUTION_ACTIVE state,
    // so filter need to be applied only for ARCHIVED executions
    const getFilter = (): FilterOperation | null => {
        if (!showArchived) {
            return null;
        }

        return {
            key: 'state',
            operation: FilterOperationName.EQ,
            value: ExecutionState.EXECUTION_ARCHIVED
        };
    };

    return {
        showArchived,
        setShowArchived,
        getFilter
    };
}
