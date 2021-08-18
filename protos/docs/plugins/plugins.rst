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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




.. _ref_flyteidl/plugins/sidecar.proto:

flyteidl/plugins/sidecar.proto
==================================================================





.. _ref_flyteidl.plugins.SidecarJob:

SidecarJob
------------------------------------------------------------------

A sidecar job brings up the desired pod_spec.
The plugin executor is responsible for keeping the pod alive until the primary container terminates
or the task itself times out.



.. csv-table:: SidecarJob type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "pod_spec", ":ref:`ref_k8s.io.api.core.v1.PodSpec`", "", ""
   "primary_container_name", ":ref:`ref_string`", "", ""
   "annotations", ":ref:`ref_flyteidl.plugins.SidecarJob.AnnotationsEntry`", "repeated", "Pod annotations"
   "labels", ":ref:`ref_flyteidl.plugins.SidecarJob.LabelsEntry`", "repeated", "Pod labels"







.. _ref_flyteidl.plugins.SidecarJob.AnnotationsEntry:

SidecarJob.AnnotationsEntry
------------------------------------------------------------------





.. csv-table:: SidecarJob.AnnotationsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""







.. _ref_flyteidl.plugins.SidecarJob.LabelsEntry:

SidecarJob.LabelsEntry
------------------------------------------------------------------





.. csv-table:: SidecarJob.LabelsEntry type fields
   :header: "Field", "Type", "Label", "Description"
   :widths: auto

   "key", ":ref:`ref_string`", "", ""
   "value", ":ref:`ref_string`", "", ""





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




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





 <!-- end messages -->



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

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->




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





 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->


