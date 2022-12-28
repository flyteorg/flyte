######################
Protocol Documentation
######################




.. _ref_flyteidl/plugins/array_job.proto:

flyteidl/plugins/array_job.proto
==================================================================





.. _ref_flyteidl.plugins.ArrayJob:

ArrayJob
------------------------------------------------------------------

Describes a job that can process independent pieces of data concurrently. Multiple copies of the runnable component
will be executed concurrently.



.. csv-table:: ArrayJob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "parallelism", ":ref:`ref_int64`", "", "Defines the minimum number of instances to bring up concurrently at any given point. Note that this is an optimistic restriction and that, due to network partitioning or other failures, the actual number of currently running instances might be more. This has to be a positive number if assigned. Default value is size."
   "size", ":ref:`ref_int64`", "", "Defines the number of instances to launch at most. This number should match the size of the input if the job requires processing of all input data. This has to be a positive number. In the case this is not defined, the back-end will determine the size at run-time by reading the inputs."
   "min_successes", ":ref:`ref_int64`", "", "An absolute number of the minimum number of successful completions of subtasks. As soon as this criteria is met, the array job will be marked as successful and outputs will be computed. This has to be a non-negative number if assigned. Default value is size (if specified)."
   "min_success_ratio", ":ref:`ref_float`", "", "If the array job size is not known beforehand, the min_success_ratio can instead be used to determine when an array job can be marked successful."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/dask.proto:

flyteidl/plugins/dask.proto
==================================================================





.. _ref_flyteidl.plugins.DaskCluster:

DaskCluster
------------------------------------------------------------------





.. csv-table:: DaskCluster type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "image", ":ref:`ref_string`", "", "Optional image to use for the scheduler as well as the default worker group. If unset, will use the default image."
   "nWorkers", ":ref:`ref_int32`", "", "Number of workers in the default worker group"
   "resources", ":ref:`ref_flyteidl.core.Resources`", "", "Resources assigned to the scheduler as well as all pods of the default worker group. As per https://kubernetes.dask.org/en/latest/kubecluster.html?highlight=limit#best-practices it is advised to only set limits. If requests are not explicitly set, the plugin will make sure to set requests==limits. The plugin sets ` --memory-limit` as well as `--nthreads` for the workers according to the limit."







.. _ref_flyteidl.plugins.DaskJob:

DaskJob
------------------------------------------------------------------

Custom Proto for Dask Plugin



.. csv-table:: DaskJob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "namespace", ":ref:`ref_string`", "", "Optional namespace to use for the dask pods. If none is given, the namespace of the Flyte task is used"
   "jobPodSpec", ":ref:`ref_flyteidl.plugins.JobPodSpec`", "", "Spec for the job pod"
   "cluster", ":ref:`ref_flyteidl.plugins.DaskCluster`", "", "Cluster"







.. _ref_flyteidl.plugins.JobPodSpec:

JobPodSpec
------------------------------------------------------------------

Specification for the job pod



.. csv-table:: JobPodSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "image", ":ref:`ref_string`", "", "Optional image to use. If unset, will use the default image."
   "resources", ":ref:`ref_flyteidl.core.Resources`", "", "Resources assigned to the job pod."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/mpi.proto:

flyteidl/plugins/mpi.proto
==================================================================





.. _ref_flyteidl.plugins.DistributedMPITrainingTask:

DistributedMPITrainingTask
------------------------------------------------------------------

MPI operator proposal https://github.com/kubeflow/community/blob/master/proposals/mpi-operator-proposal.md
Custom proto for plugin that enables distributed training using https://github.com/kubeflow/mpi-operator



.. csv-table:: DistributedMPITrainingTask type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "num_workers", ":ref:`ref_int32`", "", "number of worker spawned in the cluster for this job"
   "num_launcher_replicas", ":ref:`ref_int32`", "", "number of launcher replicas spawned in the cluster for this job The launcher pod invokes mpirun and communicates with worker pods through MPI."
   "slots", ":ref:`ref_int32`", "", "number of slots per worker used in hostfile. The available slots (GPUs) in each pod."






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/presto.proto:

flyteidl/plugins/presto.proto
==================================================================





.. _ref_flyteidl.plugins.PrestoQuery:

PrestoQuery
------------------------------------------------------------------

This message works with the 'presto' task type in the SDK and is the object that will be in the 'custom' field
of a Presto task's TaskTemplate



.. csv-table:: PrestoQuery type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "routing_group", ":ref:`ref_string`", "", ""
   "catalog", ":ref:`ref_string`", "", ""
   "schema", ":ref:`ref_string`", "", ""
   "statement", ":ref:`ref_string`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/pytorch.proto:

flyteidl/plugins/pytorch.proto
==================================================================





.. _ref_flyteidl.plugins.DistributedPyTorchTrainingTask:

DistributedPyTorchTrainingTask
------------------------------------------------------------------

Custom proto for plugin that enables distributed training using https://github.com/kubeflow/pytorch-operator



.. csv-table:: DistributedPyTorchTrainingTask type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workers", ":ref:`ref_int32`", "", "number of worker replicas spawned in the cluster for this job"






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/qubole.proto:

flyteidl/plugins/qubole.proto
==================================================================





.. _ref_flyteidl.plugins.HiveQuery:

HiveQuery
------------------------------------------------------------------

Defines a query to execute on a hive cluster.



.. csv-table:: HiveQuery type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "query", ":ref:`ref_string`", "", ""
   "timeout_sec", ":ref:`ref_uint32`", "", ""
   "retryCount", ":ref:`ref_uint32`", "", ""







.. _ref_flyteidl.plugins.HiveQueryCollection:

HiveQueryCollection
------------------------------------------------------------------

Defines a collection of hive queries.



.. csv-table:: HiveQueryCollection type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "queries", ":ref:`ref_flyteidl.plugins.HiveQuery`", "repeated", ""







.. _ref_flyteidl.plugins.QuboleHiveJob:

QuboleHiveJob
------------------------------------------------------------------

This message works with the 'hive' task type in the SDK and is the object that will be in the 'custom' field
of a hive task's TaskTemplate



.. csv-table:: QuboleHiveJob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "cluster_label", ":ref:`ref_string`", "", ""
   "query_collection", ":ref:`ref_flyteidl.plugins.HiveQueryCollection`", "", "**Deprecated.** "
   "tags", ":ref:`ref_string`", "repeated", ""
   "query", ":ref:`ref_flyteidl.plugins.HiveQuery`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/ray.proto:

flyteidl/plugins/ray.proto
==================================================================





.. _ref_flyteidl.plugins.HeadGroupSpec:

HeadGroupSpec
------------------------------------------------------------------

HeadGroupSpec are the spec for the head pod



.. csv-table:: HeadGroupSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "ray_start_params", ":ref:`ref_flyteidl.plugins.HeadGroupSpec.RayStartParamsEntry`", "repeated", "Optional. RayStartParams are the params of the start command: address, object-store-memory. Refer to https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start"







.. _ref_flyteidl.plugins.HeadGroupSpec.RayStartParamsEntry:

HeadGroupSpec.RayStartParamsEntry
------------------------------------------------------------------





.. csv-table:: HeadGroupSpec.RayStartParamsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.plugins.RayCluster:

RayCluster
------------------------------------------------------------------

Define Ray cluster defines the desired state of RayCluster



.. csv-table:: RayCluster type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "head_group_spec", ":ref:`ref_flyteidl.plugins.HeadGroupSpec`", "", "HeadGroupSpecs are the spec for the head pod"
   "worker_group_spec", ":ref:`ref_flyteidl.plugins.WorkerGroupSpec`", "repeated", "WorkerGroupSpecs are the specs for the worker pods"







.. _ref_flyteidl.plugins.RayJob:

RayJob
------------------------------------------------------------------

RayJobSpec defines the desired state of RayJob



.. csv-table:: RayJob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "ray_cluster", ":ref:`ref_flyteidl.plugins.RayCluster`", "", "RayClusterSpec is the cluster template to run the job"
   "runtime_env", ":ref:`ref_string`", "", "runtime_env is base64 encoded. Ray runtime environments: https://docs.ray.io/en/latest/ray-core/handling-dependencies.html#runtime-environments"







.. _ref_flyteidl.plugins.WorkerGroupSpec:

WorkerGroupSpec
------------------------------------------------------------------

WorkerGroupSpec are the specs for the worker pods



.. csv-table:: WorkerGroupSpec type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "group_name", ":ref:`ref_string`", "", "Required. RayCluster can have multiple worker groups, and it distinguishes them by name"
   "replicas", ":ref:`ref_int32`", "", "Required. Desired replicas of the worker group. Defaults to 1."
   "min_replicas", ":ref:`ref_int32`", "", "Optional. Min replicas of the worker group. MinReplicas defaults to 1."
   "max_replicas", ":ref:`ref_int32`", "", "Optional. Max replicas of the worker group. MaxReplicas defaults to maxInt32"
   "ray_start_params", ":ref:`ref_flyteidl.plugins.WorkerGroupSpec.RayStartParamsEntry`", "repeated", "Optional. RayStartParams are the params of the start command: address, object-store-memory. Refer to https://docs.ray.io/en/latest/ray-core/package-ref.html#ray-start"







.. _ref_flyteidl.plugins.WorkerGroupSpec.RayStartParamsEntry:

WorkerGroupSpec.RayStartParamsEntry
------------------------------------------------------------------





.. csv-table:: WorkerGroupSpec.RayStartParamsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/spark.proto:

flyteidl/plugins/spark.proto
==================================================================





.. _ref_flyteidl.plugins.SparkApplication:

SparkApplication
------------------------------------------------------------------










.. _ref_flyteidl.plugins.SparkJob:

SparkJob
------------------------------------------------------------------

Custom Proto for Spark Plugin.



.. csv-table:: SparkJob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "applicationType", ":ref:`ref_flyteidl.plugins.SparkApplication.Type`", "", ""
   "mainApplicationFile", ":ref:`ref_string`", "", ""
   "mainClass", ":ref:`ref_string`", "", ""
   "sparkConf", ":ref:`ref_flyteidl.plugins.SparkJob.SparkConfEntry`", "repeated", ""
   "hadoopConf", ":ref:`ref_flyteidl.plugins.SparkJob.HadoopConfEntry`", "repeated", ""
   "executorPath", ":ref:`ref_string`", "", "Executor path for Python jobs."
   "databricksConf", ":ref:`ref_string`", "", "databricksConf is base64 encoded string which stores databricks job configuration. Config structure can be found here. https://docs.databricks.com/dev-tools/api/2.0/jobs.html#request-structure The config is automatically encoded by flytekit, and decoded in the propeller."







.. _ref_flyteidl.plugins.SparkJob.HadoopConfEntry:

SparkJob.HadoopConfEntry
------------------------------------------------------------------





.. csv-table:: SparkJob.HadoopConfEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.plugins.SparkJob.SparkConfEntry:

SparkJob.SparkConfEntry
------------------------------------------------------------------





.. csv-table:: SparkJob.SparkConfEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""






..
   end messages



.. _ref_flyteidl.plugins.SparkApplication.Type:

SparkApplication.Type
------------------------------------------------------------------



.. csv-table:: Enum SparkApplication.Type values
   :header: "Name", "Number", "Description"
   :widths: auto

   "PYTHON", "0", ""
   "JAVA", "1", ""
   "SCALA", "2", ""
   "R", "3", ""


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/tensorflow.proto:

flyteidl/plugins/tensorflow.proto
==================================================================





.. _ref_flyteidl.plugins.DistributedTensorflowTrainingTask:

DistributedTensorflowTrainingTask
------------------------------------------------------------------

Custom proto for plugin that enables distributed training using https://github.com/kubeflow/tf-operator



.. csv-table:: DistributedTensorflowTrainingTask type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "workers", ":ref:`ref_int32`", "", "number of worker, ps, chief replicas spawned in the cluster for this job"
   "ps_replicas", ":ref:`ref_int32`", "", "PS -> Parameter server"
   "chief_replicas", ":ref:`ref_int32`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services




.. _ref_flyteidl/plugins/waitable.proto:

flyteidl/plugins/waitable.proto
==================================================================





.. _ref_flyteidl.plugins.Waitable:

Waitable
------------------------------------------------------------------

Represents an Execution that was launched and could be waited on.



.. csv-table:: Waitable type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "wf_exec_id", ":ref:`ref_flyteidl.core.WorkflowExecutionIdentifier`", "", ""
   "phase", ":ref:`ref_flyteidl.core.WorkflowExecution.Phase`", "", ""
   "workflow_id", ":ref:`ref_string`", "", ""






..
   end messages


..
   end enums


..
   end HasExtensions


..
   end services


