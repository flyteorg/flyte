import { FilterOperationName } from 'models/AdminEntity/types';
import { FilterMap } from './types';

const key = 'duration';

export type DurationFilterKey =
  | 'all'
  | 'lt1hour'
  | 'gt1hour'
  | '2to3hours'
  | '3to4hours'
  | 'gte4hours';

/** A set of WorkflowExecution duration filters to be consumed by a SingleFilterState.
 * Duration values are in seconds.
 */
export const durationFilters: FilterMap<DurationFilterKey> = {
  all: {
    label: 'All Durations',
    value: 'all',
    data: [],
  },
  lt1hour: {
    label: 'Less than 1 hour',
    value: 'lt1hour',
    data: [
      {
        key,
        operation: FilterOperationName.LT,
        value: 3600,
      },
    ],
  },
  gt1hour: {
    label: 'More than 1 hour',
    value: 'gt1hour',
    data: [
      {
        key,
        operation: FilterOperationName.GTE,
        value: 3600,
      },
    ],
  },
  '2to3hours': {
    label: '2 hours',
    value: '2to3hours',
    data: [
      {
        key,
        operation: FilterOperationName.GTE,
        value: 3600 * 2,
      },
      {
        key,
        operation: FilterOperationName.LT,
        value: 3600 * 3,
      },
    ],
  },
  '3to4hours': {
    label: '3 hours',
    value: '3to4hours',
    data: [
      {
        key,
        operation: FilterOperationName.GTE,
        value: 3600 * 3,
      },
      {
        key,
        operation: FilterOperationName.LT,
        value: 3600 * 4,
      },
    ],
  },
  gte4hours: {
    label: 'More than 3 hours',
    value: 'gte4hours',
    data: [
      {
        key,
        operation: FilterOperationName.GTE,
        value: 3600 * 4,
      },
    ],
  },
};
