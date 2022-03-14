import { InfiniteQueryObserverOptions, QueryObserverOptions } from 'react-query';

export enum QueryType {
  NodeExecutionDetails = 'NodeExecutionDetails',
  DynamicWorkflowFromNodeExecution = 'DynamicWorkflowFromNodeExecution',
  NodeExecution = 'nodeExecution',
  NodeExecutionList = 'nodeExecutionList',
  NodeExecutionChildList = 'nodeExecutionChildList',
  NodeExecutionTreeList = 'nodeExecutionTreeList',
  TaskExecution = 'taskExecution',
  TaskExecutionList = 'taskExecutionList',
  TaskExecutionChildList = 'taskExecutionChildList',
  TaskTemplate = 'taskTemplate',
  Workflow = 'workflow',
  WorkflowExecution = 'workflowExecution',
  WorkflowExecutionList = 'workflowExecutionList'
}

type QueryKeyArray = [QueryType, ...unknown[]];
export interface QueryInput<T> extends QueryObserverOptions<T, Error> {
  queryKey: QueryKeyArray;
  queryFn: QueryObserverOptions<T, Error>['queryFn'];
}

export interface InfiniteQueryInput<T> extends InfiniteQueryObserverOptions<InfiniteQueryPage<T>, Error> {
  queryKey: QueryKeyArray;
  queryFn: InfiniteQueryObserverOptions<InfiniteQueryPage<T>, Error>['queryFn'];
}

export interface InfiniteQueryPage<T> {
  data: T[];
  token?: string;
}
