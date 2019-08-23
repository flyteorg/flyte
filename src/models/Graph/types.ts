import { NodeExecution } from 'models/Execution';
import { TaskTemplate } from 'models/Task';

/** A flyte-graph compatible node representation which also includes all of the
 * additional task data needed for our custom rendering
 */
export interface DAGNode {
    execution?: NodeExecution;
    id: string;
    parentIds?: string[];
    taskTemplate?: TaskTemplate;
}
