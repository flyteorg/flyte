import { useQueryState } from 'components/hooks/useQueryState';
import { isEqual, mapValues, pickBy } from 'lodash';
import { FilterOperationName, FilterOperationValue } from 'models/AdminEntity/types';
import { useEffect, useState } from 'react';
import { FilterValue, MultiFilterState } from './types';
import { useFilterButtonState } from './useFilterButtonState';

function serializeForQueryState(values: any[]) {
  return values.join(';');
}
function deserializeFromQueryState(stateValue = '') {
  return stateValue.split(';');
}

interface MultiFilterStateArgs<FilterKey extends string, DataType extends FilterOperationValue> {
  options: Record<FilterKey, FilterValue<string, DataType>>;
  defaultValue: FilterKey[];
  filterKey: string;
  filterOperation?: FilterOperationName;
  label: string;
  listHeader: string;
  queryStateKey: string;
}

/** Maintains the state for a generic MultiSelectForm filter.
 * The generated `FilterOperation` will use the provided `key` and `operation` (
 * defaults to `VALUE_IN`).
 * The current state of the filter will be synced with the query string using the provided `queryStateKey` value.
 */
export function useMultiFilterState<
  FilterKey extends string,
  DataType extends FilterOperationValue,
>({
  options,
  defaultValue,
  filterKey,
  filterOperation = FilterOperationName.VALUE_IN,
  label,
  listHeader,
  queryStateKey,
}: MultiFilterStateArgs<FilterKey, DataType>): MultiFilterState<FilterKey, DataType> {
  const defaultMappedOptions = mapValues(options, (_, key) =>
    defaultValue.includes(key as FilterKey),
  ) as Record<FilterKey, boolean>;
  const { params, setQueryStateValue } = useQueryState<Record<string, string>>();
  const queryStateValue = params[queryStateKey];
  const parsedQueryState = deserializeFromQueryState(queryStateValue);
  const queryStateSelected = mapValues(options, (_, key) =>
    parsedQueryState.includes(key),
  ) as Record<FilterKey, boolean>;

  const [selected, setSelected] = useState({
    ...defaultMappedOptions,
    ...queryStateSelected,
  });
  const button = useFilterButtonState();

  const selectedKeys = Object.keys(pickBy(selected));
  const active = !isEqual(selectedKeys, defaultValue);

  const onChange = (selectedStates: Record<FilterKey, boolean>) => {
    setSelected(selectedStates);
  };

  const onReset = () => {
    setSelected(defaultMappedOptions);
    button.setOpen(false);
  };

  // Sync changes to selected value with the query state
  useEffect(() => {
    const queryValue = active ? serializeForQueryState(selectedKeys) : undefined;
    setQueryStateValue(queryStateKey, queryValue);
  }, [selectedKeys.join(), queryStateKey]);

  // Watch for external changes to query string and sync our value if
  // it's different
  useEffect(() => {
    if (queryStateValue) {
      if (!isEqual(selected, queryStateSelected)) {
        setSelected(queryStateSelected);
      }
    }
  }, [queryStateValue]);

  const getFilter = () => {
    // Collect the `data` property from all options for which we have
    // an active selected state
    const value = (Object.keys(options) as FilterKey[]).reduce<DataType[]>(
      (out, key) => (selected[key] ? [...out, options[key].data] : out),
      [],
    );
    return value.length
      ? [
          {
            value,
            key: filterKey,
            operation: filterOperation,
          },
        ]
      : [];
  };

  return {
    active,
    button,
    getFilter,
    label,
    listHeader,
    onChange,
    onReset,
    type: 'multi',
    values: Object.values(options),
    selectedStates: selected,
  };
}
