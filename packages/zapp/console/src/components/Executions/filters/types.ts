import { FilterOperation, FilterOperationList } from 'models/AdminEntity/types';

export interface FilterValue<FilterKey extends string, DataType = FilterOperationList> {
  label: string;
  value: FilterKey;
  data: DataType;
}

export type FilterMap<FilterKey extends string, DataType = FilterOperationList> = Record<
  FilterKey,
  FilterValue<FilterKey, DataType>
>;

export interface FilterButtonState {
  open: boolean;
  setOpen: (open: boolean) => void;
  onClick: () => void;
}

export type FilterStateType = 'single' | 'multi' | 'search' | 'boolean';

export interface FilterState {
  active: boolean;
  button: FilterButtonState;
  label: string;
  type: FilterStateType;
  getFilter: () => FilterOperation[];
  onReset: () => void;
  onChange?: (value) => void;
}

export interface SingleFilterState<FilterKey extends string> extends FilterState {
  onChange: (newValue: string) => void;
  selectedValue: FilterKey;
  type: 'single';
  values: FilterValue<FilterKey>[];
}

export interface SearchFilterState extends FilterState {
  onChange: (newValue: string) => void;
  placeholder: string;
  type: 'search';
  value: string;
}

export interface MultiFilterState<FilterKey extends string, DataType> extends FilterState {
  listHeader: string;
  onChange: (selectedStates: Record<FilterKey, boolean>) => void;
  selectedStates: Record<FilterKey, boolean>;
  type: 'multi';
  values: FilterValue<FilterKey, DataType>[];
}

export interface BooleanFilterState extends FilterState {
  setActive: (active: boolean) => void;
  type: 'boolean';
}
