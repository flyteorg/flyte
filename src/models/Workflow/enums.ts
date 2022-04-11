import { Admin } from 'flyteidl';

/** These enums are only aliased and exported from this file. They should
 * be imported directly from here to avoid runtime errors when TS processes
 * modules individually (such as when running with ts-jest)
 */

/* It's an ENUM exports, and as such need to be exported as both type and const value */
/* eslint-disable @typescript-eslint/no-redeclare */
export type WorkflowExecutionState = Admin.NamedEntityState;
export const WorkflowExecutionState = Admin.NamedEntityState;
