import { Admin } from 'flyteidl';

/** These enums are only aliased and exported from this file. They should
 * be imported directly from here to avoid runtime errors when TS processes
 * modules individually (such as when running with ts-jest)
 */

/* eslint-disable @typescript-eslint/no-redeclare */
export type NamedEntityState = Admin.NamedEntityState;
export const NamedEntityState = Admin.NamedEntityState;
