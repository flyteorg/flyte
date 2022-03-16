import { Admin } from 'flyteidl';
import * as $protobuf from 'protobufjs';

/** Maps filter operations to the strings which should be used in queries to the Admin API */
export enum FilterOperationName {
  CONTAINS = 'contains',
  EQ = 'eq',
  GT = 'gt',
  GTE = 'gte',
  LT = 'lt',
  LTE = 'lte',
  NE = 'ne',
  VALUE_IN = 'value_in',
}

/** Represents a query filter for collection endpoints. Multiple filters can be combined into a single
 * query string.
 */

export type FilterOperationValue = string | number;
export type FilterOperationValueList = FilterOperationValue[];
export type FilterOperationValueGenerator = () => FilterOperationValue | FilterOperationValueList;
export interface FilterOperation {
  key: string;
  operation: FilterOperationName;
  value: FilterOperationValue | FilterOperationValueList | FilterOperationValueGenerator;
}

export type FilterOperationList = FilterOperation[];

export type SortDirection = Admin.Sort.Direction;
export const SortDirection = Admin.Sort.Direction;
export type Sort = RequiredNonNullable<Admin.ISort>;

/** Generic interface for any class which can be used to decode a protobuf message. */
export interface DecodableType<T> {
  decode(reader: $protobuf.Reader | Uint8Array, length?: number): T;
}

/** Generic interface for any class which can be used to encode a protobuf message. */
export interface EncodableType<T> {
  encode(
    // tslint:disable-next-line:no-any
    message: T,
    writer?: $protobuf.Writer,
  ): $protobuf.Writer;
}

/** Options for requests made to the Admin API.
 * @param data Specifies the data for a POST/PUT request, or query params to
 * append for GET requests.
 * @param filter A list of filter operations for use with collection endpoints.
 * @param limit Limits the records returned by a collection endpoint.
 * @param params A query params object to merge into the generated query string.
 * @param token An opaque token returned by calls to collection endpoints in the
 * case that more results are available. This should be passed in order to
 * retrieve the next page of results. It does not encode any other query
 * parameters, so those should always be passed.
 */
export interface RequestConfig {
  data?: unknown;
  filter?: FilterOperationList;
  limit?: number;
  params?: Record<string, any>;
  token?: string;
  sort?: Sort;
}

/** A generic shape for any response returned from a list endpoint */
export interface PaginatedEntityResponse<T> {
  entities: T[];
  token?: string;
}

/** A function that translates from an Admin.* entity to a local model type */
export type AdminEntityTransformer<T, TransformedType> = (message: T) => TransformedType;
