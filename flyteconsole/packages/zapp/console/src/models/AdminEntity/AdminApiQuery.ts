import {
  FilterOperation,
  FilterOperationList,
  FilterOperationName,
  Sort,
  SortDirection,
} from './types';

import { sortQueryKeys } from './constants';

interface QueryConfig {
  filter?: FilterOperationList;
  limit?: number;
  sort?: Sort;
  token?: string;
}

type ConverterFnValue = string | number | (string | number)[];
type ConverterFn = (key: string, value: ConverterFnValue) => string;

function formatValueInQuery(values: ConverterFnValue) {
  return Array.isArray(values) ? values.join(';') : values;
}

const converters: { [k in FilterOperationName]: ConverterFn } = {
  [FilterOperationName.CONTAINS]: (key, value) => `contains(${key},${value})`,
  [FilterOperationName.EQ]: (key, value) => `eq(${key},${value})`,
  [FilterOperationName.GT]: (key, value) => `gt(${key},${value})`,
  [FilterOperationName.GTE]: (key, value) => `gte(${key},${value})`,
  [FilterOperationName.LT]: (key, value) => `lt(${key},${value})`,
  [FilterOperationName.LTE]: (key, value) => `lte(${key},${value})`,
  [FilterOperationName.NE]: (key, value) => `ne(${key},${value})`,
  [FilterOperationName.VALUE_IN]: (key, value) => `value_in(${key},${formatValueInQuery(value)})`,
};

function generateOperation({ key, operation, value }: FilterOperation) {
  const converter = converters[operation];
  if (!converter) {
    throw new Error(`AdminApiQuery does not implement ${operation}`);
  }
  const finalValue = typeof value === 'function' ? value() : value;
  return converter(key, finalValue);
}

const filterSeparator = '+';

function generateFilterQuery(filter: FilterOperationList) {
  return filter.map(generateOperation).join(filterSeparator);
}

function generateSortQuery(sort: Sort) {
  return {
    // Mapping from enum value to string name (ASECNDING/DESCENDING)
    [sortQueryKeys.direction]: SortDirection[sort.direction],
    [sortQueryKeys.key]: sort.key,
  };
}

/** Given a request config object, will generate any applicable query string
 * parameters needed to filter/sort/paginate results from the Admin API.
 */
export function generateAdminApiQuery(config: QueryConfig = { filter: [] }) {
  let filters;
  let sort;
  if (config.filter) {
    filters = generateFilterQuery(config.filter);
  }
  if (config.sort) {
    sort = generateSortQuery(config.sort);
  }

  return {
    filters,
    token: config.token,
    limit: config.limit,
    ...sort,
  };
}
