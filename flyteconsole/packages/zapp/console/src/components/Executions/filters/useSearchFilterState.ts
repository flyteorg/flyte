import { useQueryState } from 'components/hooks/useQueryState';
import { FilterOperationName } from 'models/AdminEntity/types';
import { useEffect, useState } from 'react';
import { SearchFilterState } from './types';
import { useFilterButtonState } from './useFilterButtonState';

interface SearchFilterStateArgs {
  defaultValue?: string;
  filterKey: string;
  filterOperation?: FilterOperationName;
  label: string;
  placeholder: string;
  queryStateKey: string;
}

/** Maintains the state for a `SearchInputForm` filter.
 * The generated `FilterOperation` will use the provided `key` and `operation`
 * (defaults to `EQ`)
 * The current search value will be synced to the query string using the
 * provided `queryStateKey` value.
 */
export function useSearchFilterState({
  defaultValue = '',
  filterKey,
  filterOperation = FilterOperationName.EQ,
  label,
  placeholder,
  queryStateKey,
}: SearchFilterStateArgs): SearchFilterState {
  const { params, setQueryStateValue } = useQueryState<Record<string, string>>();
  const queryStateValue = params[queryStateKey];

  const [value, setValue] = useState(defaultValue);
  const active = value !== defaultValue;

  const button = useFilterButtonState();
  const onChange = (newValue: string) => {
    setValue(newValue);
    // Automatically hide the form on submission of a new value
    button.setOpen(false);
  };

  const onReset = () => {
    setValue(defaultValue);
    button.setOpen(false);
  };

  useEffect(() => {
    const queryValue = value === defaultValue ? undefined : value;
    setQueryStateValue(queryStateKey, queryValue);
  }, [value, queryStateKey]);

  useEffect(() => {
    if (queryStateValue) {
      setValue(queryStateValue);
    }
  }, [queryStateValue]);

  const getFilter = () =>
    value
      ? [
          {
            value,
            key: filterKey,
            operation: filterOperation,
          },
        ]
      : [];

  return {
    active,
    button,
    getFilter,
    onChange,
    onReset,
    label,
    placeholder,
    value,
    type: 'search',
  };
}
