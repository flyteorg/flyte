import { APIContextValue, useAPIContext } from 'components/data/apiContext';
import { FilterOperationName, SortDirection } from 'models/AdminEntity/types';
import { Identifier, NamedEntityIdentifier } from 'models/Common/types';
import { taskSortFields } from 'models/Task/constants';
import { Task } from 'models/Task/types';
import { useMemo, useState } from 'react';
import { SearchableSelectorOption } from './SearchableSelector';
import { TaskSourceSelectorState } from './types';
import { useVersionSelectorOptions } from './useVersionSelectorOptions';
import { versionsToSearchableSelectorOptions } from './utils';

function generateFetchSearchResults({ listTasks }: APIContextValue, taskId: NamedEntityIdentifier) {
  return async (query: string) => {
    const { project, domain, name } = taskId;
    const { entities: tasks } = await listTasks(
      { project, domain, name },
      {
        filter: [
          {
            key: 'version',
            operation: FilterOperationName.CONTAINS,
            value: query,
          },
        ],
        sort: {
          key: taskSortFields.createdAt,
          direction: SortDirection.DESCENDING,
        },
      },
    );
    return versionsToSearchableSelectorOptions(tasks);
  };
}

interface UseTaskSourceSelectorStateArgs {
  /** The parent task for which we are selecting a version. */
  sourceId: NamedEntityIdentifier;
  /** The currently selected Task version. */
  taskVersion?: Identifier;
  /** The list of options to show for the Task selector. */
  taskVersionOptions: Task[];
  /** Callback fired when a task has been selected. */
  selectTaskVersion(task: Identifier): void;
}

/** Generates state for the version selector rendered when using a task
 * as a source in the Launch form.
 */
export function useTaskSourceSelectorState({
  sourceId,
  taskVersion,
  taskVersionOptions,
  selectTaskVersion,
}: UseTaskSourceSelectorStateArgs): TaskSourceSelectorState {
  const apiContext = useAPIContext();
  const taskSelectorOptions = useVersionSelectorOptions(taskVersionOptions);
  const [taskVersionSearchOptions, setTaskVersionSearchOptions] = useState<
    SearchableSelectorOption<Identifier>[]
  >([]);

  const selectedTask = useMemo(() => {
    if (!taskVersion) {
      return undefined;
    }
    // Search both the default and search results to match our selected task
    // with the correct SearchableSelector item.
    return [...taskSelectorOptions, ...taskVersionSearchOptions].find(
      (option) => option.id === taskVersion.version,
    );
  }, [taskVersion, taskVersionOptions]);

  const onSelectTaskVersion = useMemo(
    () =>
      ({ data }: SearchableSelectorOption<Identifier>) =>
        selectTaskVersion(data),
    [selectTaskVersion],
  );

  const fetchSearchResults = useMemo(() => {
    const doFetch = generateFetchSearchResults(apiContext, sourceId);
    return async (query: string) => {
      const results = await doFetch(query);
      setTaskVersionSearchOptions(results);
      return results;
    };
  }, [apiContext, sourceId, setTaskVersionSearchOptions]);

  return {
    fetchSearchResults,
    onSelectTaskVersion,
    selectedTask,
    taskSelectorOptions,
  };
}
