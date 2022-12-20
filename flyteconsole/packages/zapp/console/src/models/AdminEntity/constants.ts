import { RequestConfig } from './types';

export const apiPrefix = '/api/v1';

export const limits = {
  DEFAULT: 25,
  /** The admin API requires a limit value for all list endpoints, but does not
   * have the concept of returning unlimited results. For the few use cases that
   * we *want* to list all results, this value can be used as `limit` in a
   * `RequestConfig`. Use with caution, as this could result in a large query
   * response.
   */
  NONE: 10000,
};

export const sortQueryKeys = {
  direction: 'sort_by.direction',
  key: 'sort_by.key',
};

/** Sane values to be used as a basis for any endpoints returning paginated results */
export const defaultPaginationConfig: RequestConfig = {
  limit: limits.DEFAULT,
};

/** For listing execution children, we generally do *not* want multiple pages */
export const defaultListExecutionChildrenConfig: RequestConfig = {
  limit: limits.NONE,
};
