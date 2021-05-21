# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [flyteidl/plugins/array_job.proto](#flyteidl/plugins/array_job.proto)
    - [ArrayJob](#flyteidl.plugins.ArrayJob)
  
- [flyteidl/plugins/presto.proto](#flyteidl/plugins/presto.proto)
    - [PrestoQuery](#flyteidl.plugins.PrestoQuery)
  
- [flyteidl/plugins/pytorch.proto](#flyteidl/plugins/pytorch.proto)
    - [DistributedPyTorchTrainingTask](#flyteidl.plugins.DistributedPyTorchTrainingTask)
  
- [flyteidl/plugins/qubole.proto](#flyteidl/plugins/qubole.proto)
    - [HiveQuery](#flyteidl.plugins.HiveQuery)
    - [HiveQueryCollection](#flyteidl.plugins.HiveQueryCollection)
    - [QuboleHiveJob](#flyteidl.plugins.QuboleHiveJob)
  
- [flyteidl/plugins/sidecar.proto](#flyteidl/plugins/sidecar.proto)
    - [SidecarJob](#flyteidl.plugins.SidecarJob)
  
- [flyteidl/plugins/spark.proto](#flyteidl/plugins/spark.proto)
    - [SparkApplication](#flyteidl.plugins.SparkApplication)
    - [SparkJob](#flyteidl.plugins.SparkJob)
    - [SparkJob.HadoopConfEntry](#flyteidl.plugins.SparkJob.HadoopConfEntry)
    - [SparkJob.SparkConfEntry](#flyteidl.plugins.SparkJob.SparkConfEntry)
  
    - [SparkApplication.Type](#flyteidl.plugins.SparkApplication.Type)
  
- [flyteidl/plugins/tensorflow.proto](#flyteidl/plugins/tensorflow.proto)
    - [DistributedTensorflowTrainingTask](#flyteidl.plugins.DistributedTensorflowTrainingTask)
  
- [flyteidl/plugins/waitable.proto](#flyteidl/plugins/waitable.proto)
    - [Waitable](#flyteidl.plugins.Waitable)
  
- [Scalar Value Types](#scalar-value-types)



<a name="flyteidl/plugins/array_job.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/array_job.proto



<a name="flyteidl.plugins.ArrayJob"></a>

### ArrayJob
Describes a job that can process independent pieces of data concurrently. Multiple copies of the runnable component
will be executed concurrently.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| parallelism | [int64](#int64) |  | Defines the minimum number of instances to bring up concurrently at any given point. Note that this is an optimistic restriction and that, due to network partitioning or other failures, the actual number of currently running instances might be more. This has to be a positive number if assigned. Default value is size. |
| size | [int64](#int64) |  | Defines the number of instances to launch at most. This number should match the size of the input if the job requires processing of all input data. This has to be a positive number. In the case this is not defined, the back-end will determine the size at run-time by reading the inputs. |
| min_successes | [int64](#int64) |  | An absolute number of the minimum number of successful completions of subtasks. As soon as this criteria is met, the array job will be marked as successful and outputs will be computed. This has to be a non-negative number if assigned. Default value is size (if specified). |
| min_success_ratio | [float](#float) |  | If the array job size is not known beforehand, the min_success_ratio can instead be used to determine when an array job can be marked successful. |





 

 

 

 



<a name="flyteidl/plugins/presto.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/presto.proto



<a name="flyteidl.plugins.PrestoQuery"></a>

### PrestoQuery
This message works with the &#39;presto&#39; task type in the SDK and is the object that will be in the &#39;custom&#39; field
of a Presto task&#39;s TaskTemplate


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| routing_group | [string](#string) |  |  |
| catalog | [string](#string) |  |  |
| schema | [string](#string) |  |  |
| statement | [string](#string) |  |  |





 

 

 

 



<a name="flyteidl/plugins/pytorch.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/pytorch.proto



<a name="flyteidl.plugins.DistributedPyTorchTrainingTask"></a>

### DistributedPyTorchTrainingTask
Custom proto for plugin that enables distributed training using https://github.com/kubeflow/pytorch-operator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workers | [int32](#int32) |  | number of worker replicas spawned in the cluster for this job |





 

 

 

 



<a name="flyteidl/plugins/qubole.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/qubole.proto



<a name="flyteidl.plugins.HiveQuery"></a>

### HiveQuery
Defines a query to execute on a hive cluster.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| query | [string](#string) |  |  |
| timeout_sec | [uint32](#uint32) |  |  |
| retryCount | [uint32](#uint32) |  |  |






<a name="flyteidl.plugins.HiveQueryCollection"></a>

### HiveQueryCollection
Defines a collection of hive queries.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| queries | [HiveQuery](#flyteidl.plugins.HiveQuery) | repeated |  |






<a name="flyteidl.plugins.QuboleHiveJob"></a>

### QuboleHiveJob
This message works with the &#39;hive&#39; task type in the SDK and is the object that will be in the &#39;custom&#39; field
of a hive task&#39;s TaskTemplate


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cluster_label | [string](#string) |  |  |
| query_collection | [HiveQueryCollection](#flyteidl.plugins.HiveQueryCollection) |  | **Deprecated.**  |
| tags | [string](#string) | repeated |  |
| query | [HiveQuery](#flyteidl.plugins.HiveQuery) |  |  |





 

 

 

 



<a name="flyteidl/plugins/sidecar.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/sidecar.proto



<a name="flyteidl.plugins.SidecarJob"></a>

### SidecarJob
A sidecar job brings up the desired pod_spec.
The plugin executor is responsible for keeping the pod alive until the primary container terminates
or the task itself times out.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pod_spec | [k8s.io.api.core.v1.PodSpec](#k8s.io.api.core.v1.PodSpec) |  |  |
| primary_container_name | [string](#string) |  |  |





 

 

 

 



<a name="flyteidl/plugins/spark.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/spark.proto



<a name="flyteidl.plugins.SparkApplication"></a>

### SparkApplication







<a name="flyteidl.plugins.SparkJob"></a>

### SparkJob
Custom Proto for Spark Plugin.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| applicationType | [SparkApplication.Type](#flyteidl.plugins.SparkApplication.Type) |  |  |
| mainApplicationFile | [string](#string) |  |  |
| mainClass | [string](#string) |  |  |
| sparkConf | [SparkJob.SparkConfEntry](#flyteidl.plugins.SparkJob.SparkConfEntry) | repeated |  |
| hadoopConf | [SparkJob.HadoopConfEntry](#flyteidl.plugins.SparkJob.HadoopConfEntry) | repeated |  |
| executorPath | [string](#string) |  | Executor path for Python jobs. |






<a name="flyteidl.plugins.SparkJob.HadoopConfEntry"></a>

### SparkJob.HadoopConfEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="flyteidl.plugins.SparkJob.SparkConfEntry"></a>

### SparkJob.SparkConfEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="flyteidl.plugins.SparkApplication.Type"></a>

### SparkApplication.Type


| Name | Number | Description |
| ---- | ------ | ----------- |
| PYTHON | 0 |  |
| JAVA | 1 |  |
| SCALA | 2 |  |
| R | 3 |  |


 

 

 



<a name="flyteidl/plugins/tensorflow.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/tensorflow.proto



<a name="flyteidl.plugins.DistributedTensorflowTrainingTask"></a>

### DistributedTensorflowTrainingTask
Custom proto for plugin that enables distributed training using https://github.com/kubeflow/tf-operator


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workers | [int32](#int32) |  | number of worker, ps, chief replicas spawned in the cluster for this job |
| ps_replicas | [int32](#int32) |  | PS -&gt; Parameter server |
| chief_replicas | [int32](#int32) |  |  |





 

 

 

 



<a name="flyteidl/plugins/waitable.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## flyteidl/plugins/waitable.proto



<a name="flyteidl.plugins.Waitable"></a>

### Waitable
Represents an Execution that was launched and could be waited on.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| wf_exec_id | [flyteidl.core.WorkflowExecutionIdentifier](#flyteidl.core.WorkflowExecutionIdentifier) |  |  |
| phase | [flyteidl.core.WorkflowExecution.Phase](#flyteidl.core.WorkflowExecution.Phase) |  |  |
| workflow_id | [string](#string) |  |  |





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

