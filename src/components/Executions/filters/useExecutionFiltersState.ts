import { compact, flatMap } from 'lodash';
import { FilterOperation } from 'models/AdminEntity/types';
import { filterLabels } from './constants';
import { durationFilters } from './durationFilters';
import {
  nodeExecutionStartTimeFilters,
  workflowExecutionStartTimeFilters,
} from './startTimeFilters';
import { nodeExecutionStatusFilters, workflowExecutionStatusFilters } from './statusFilters';
import { FilterState } from './types';
import { useMultiFilterState } from './useMultiFilterState';
import { useSearchFilterState } from './useSearchFilterState';
import { useSingleFilterState } from './useSingleFilterState';
import { useCurrentUserOnlyFilterState } from './useCurrentUserOnlyFilterState';

export interface ExecutionFiltersState {
  appliedFilters: FilterOperation[];
  filters: FilterState[];
}

/** Maintains state to be used with a `ExecutionFilters` component. This state
 * combines multiple generic filter states (SingleFilterState, MultiFilterState,
 * SearchInputState) as well as an `appliedFilters` list which can be passed
 * directly in a `RequestConfig`.
 */
export function useExecutionFiltersState(filters: FilterState[]): ExecutionFiltersState {
  const appliedFilters = compact(flatMap(filters, ({ getFilter }) => getFilter()));

  return { filters, appliedFilters };
}

export function useWorkflowExecutionFiltersState() {
  return useExecutionFiltersState([
    useMultiFilterState({
      options: workflowExecutionStatusFilters,
      defaultValue: [],
      filterKey: 'phase',
      label: filterLabels.status,
      listHeader: 'Filter By',
      queryStateKey: 'status',
    }),
    useSearchFilterState({
      filterKey: 'workflow.version',
      label: filterLabels.version,
      placeholder: 'Enter Version String',
      queryStateKey: 'version',
    }),
    useSingleFilterState({
      options: workflowExecutionStartTimeFilters,
      defaultValue: workflowExecutionStartTimeFilters.all,
      label: filterLabels.startTime,
      queryStateKey: 'startTime',
    }),
    useSingleFilterState({
      options: durationFilters,
      defaultValue: durationFilters.all,
      label: filterLabels.duration,
      queryStateKey: 'duration',
    }),
    useCurrentUserOnlyFilterState(),
  ]);
}

export function useNodeExecutionFiltersState() {
  return useExecutionFiltersState([
    useMultiFilterState({
      options: nodeExecutionStatusFilters,
      defaultValue: [],
      filterKey: 'phase',
      label: filterLabels.status,
      listHeader: 'Filter By',
      queryStateKey: 'status',
    }),
    useSingleFilterState({
      options: nodeExecutionStartTimeFilters,
      defaultValue: nodeExecutionStartTimeFilters.all,
      label: filterLabels.startTime,
      queryStateKey: 'startTime',
    }),
    useSingleFilterState({
      options: durationFilters,
      defaultValue: durationFilters.all,
      label: filterLabels.duration,
      queryStateKey: 'duration',
    }),
  ]);
}
