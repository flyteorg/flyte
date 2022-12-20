/** Defines the values of `TaskTemplate.type` for each of the known
 * first-class task types in Flyte
 */
export enum TaskType {
  ARRAY = 'container_array',
  BATCH_HIVE = 'batch_hive',
  DYNAMIC = 'dynamic-task',
  HIVE = 'hive',
  PYTHON = 'python-task',
  SIDECAR = 'sidecar',
  SPARK = 'spark',
  UNKNOWN = 'unknown',
  WAITABLE = 'waitable',
  MPI = 'mpi',

  // plugins
  ARRAY_AWS = 'aws-batch',
  ARRAY_K8S = 'k8s-array',
}

export const taskSortFields = {
  createdAt: 'created_at',
  name: 'name',
};
