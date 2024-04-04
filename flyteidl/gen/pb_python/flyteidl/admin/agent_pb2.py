# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/admin/agent.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2
from flyteidl.core import tasks_pb2 as flyteidl_dot_core_dot_tasks__pb2
from flyteidl.core import workflow_pb2 as flyteidl_dot_core_dot_workflow__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import execution_pb2 as flyteidl_dot_core_dot_execution__pb2
from flyteidl.core import metrics_pb2 as flyteidl_dot_core_dot_metrics__pb2
from flyteidl.core import security_pb2 as flyteidl_dot_core_dot_security__pb2
from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1a\x66lyteidl/admin/agent.proto\x12\x0e\x66lyteidl.admin\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x19\x66lyteidl/core/tasks.proto\x1a\x1c\x66lyteidl/core/workflow.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1d\x66lyteidl/core/execution.proto\x1a\x1b\x66lyteidl/core/metrics.proto\x1a\x1c\x66lyteidl/core/security.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1cgoogle/protobuf/struct.proto\"\x9e\x07\n\x15TaskExecutionMetadata\x12R\n\x11task_execution_id\x18\x01 \x01(\x0b\x32&.flyteidl.core.TaskExecutionIdentifierR\x0ftaskExecutionId\x12\x1c\n\tnamespace\x18\x02 \x01(\tR\tnamespace\x12I\n\x06labels\x18\x03 \x03(\x0b\x32\x31.flyteidl.admin.TaskExecutionMetadata.LabelsEntryR\x06labels\x12X\n\x0b\x61nnotations\x18\x04 \x03(\x0b\x32\x36.flyteidl.admin.TaskExecutionMetadata.AnnotationsEntryR\x0b\x61nnotations\x12.\n\x13k8s_service_account\x18\x05 \x01(\tR\x11k8sServiceAccount\x12t\n\x15\x65nvironment_variables\x18\x06 \x03(\x0b\x32?.flyteidl.admin.TaskExecutionMetadata.EnvironmentVariablesEntryR\x14\x65nvironmentVariables\x12!\n\x0cmax_attempts\x18\x07 \x01(\x05R\x0bmaxAttempts\x12$\n\rinterruptible\x18\x08 \x01(\x08R\rinterruptible\x12\x46\n\x1finterruptible_failure_threshold\x18\t \x01(\x05R\x1dinterruptibleFailureThreshold\x12>\n\toverrides\x18\n \x01(\x0b\x32 .flyteidl.core.TaskNodeOverridesR\toverrides\x12\x33\n\x08identity\x18\x0b \x01(\x0b\x32\x17.flyteidl.core.IdentityR\x08identity\x1a\x39\n\x0bLabelsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x1a>\n\x10\x41nnotationsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x1aG\n\x19\x45nvironmentVariablesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"\x1e\n\x06Secret\x12\x14\n\x05value\x18\x01 \x01(\tR\x05value\"\xb5\x02\n\x11\x43reateTaskRequest\x12\x31\n\x06inputs\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x06inputs\x12\x37\n\x08template\x18\x02 \x01(\x0b\x32\x1b.flyteidl.core.TaskTemplateR\x08template\x12#\n\routput_prefix\x18\x03 \x01(\tR\x0coutputPrefix\x12]\n\x17task_execution_metadata\x18\x04 \x01(\x0b\x32%.flyteidl.admin.TaskExecutionMetadataR\x15taskExecutionMetadata\x12\x30\n\x07secrets\x18\x05 \x03(\x0b\x32\x16.flyteidl.admin.SecretR\x07secrets\"9\n\x12\x43reateTaskResponse\x12#\n\rresource_meta\x18\x01 \x01(\x0cR\x0cresourceMeta\"\xb9\x02\n\x13\x43reateRequestHeader\x12\x37\n\x08template\x18\x01 \x01(\x0b\x32\x1b.flyteidl.core.TaskTemplateR\x08template\x12#\n\routput_prefix\x18\x02 \x01(\tR\x0coutputPrefix\x12]\n\x17task_execution_metadata\x18\x03 \x01(\x0b\x32%.flyteidl.admin.TaskExecutionMetadataR\x15taskExecutionMetadata\x12\x33\n\x16max_dataset_size_bytes\x18\x04 \x01(\x03R\x13maxDatasetSizeBytes\x12\x30\n\x07secrets\x18\x05 \x03(\x0b\x32\x16.flyteidl.admin.SecretR\x07secrets\"\x94\x01\n\x16\x45xecuteTaskSyncRequest\x12=\n\x06header\x18\x01 \x01(\x0b\x32#.flyteidl.admin.CreateRequestHeaderH\x00R\x06header\x12\x33\n\x06inputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\x06inputsB\x06\n\x04part\"U\n\x1d\x45xecuteTaskSyncResponseHeader\x12\x34\n\x08resource\x18\x01 \x01(\x0b\x32\x18.flyteidl.admin.ResourceR\x08resource\"\xa0\x01\n\x17\x45xecuteTaskSyncResponse\x12G\n\x06header\x18\x01 \x01(\x0b\x32-.flyteidl.admin.ExecuteTaskSyncResponseHeaderH\x00R\x06header\x12\x35\n\x07outputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\x07outputsB\x05\n\x03res\"\xcb\x01\n\x0eGetTaskRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x41\n\rtask_category\x18\x03 \x01(\x0b\x32\x1c.flyteidl.admin.TaskCategoryR\x0ctaskCategory\x12\x30\n\x07secrets\x18\x04 \x03(\x0b\x32\x16.flyteidl.admin.SecretR\x07secrets\"G\n\x0fGetTaskResponse\x12\x34\n\x08resource\x18\x01 \x01(\x0b\x32\x18.flyteidl.admin.ResourceR\x08resource\"\xb3\x02\n\x08Resource\x12/\n\x05state\x18\x01 \x01(\x0e\x32\x15.flyteidl.admin.StateB\x02\x18\x01R\x05state\x12\x33\n\x07outputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x07outputs\x12\x18\n\x07message\x18\x03 \x01(\tR\x07message\x12\x33\n\tlog_links\x18\x04 \x03(\x0b\x32\x16.flyteidl.core.TaskLogR\x08logLinks\x12\x38\n\x05phase\x18\x05 \x01(\x0e\x32\".flyteidl.core.TaskExecution.PhaseR\x05phase\x12\x38\n\x0b\x63ustom_info\x18\x06 \x01(\x0b\x32\x17.google.protobuf.StructR\ncustomInfo\"\xce\x01\n\x11\x44\x65leteTaskRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x41\n\rtask_category\x18\x03 \x01(\x0b\x32\x1c.flyteidl.admin.TaskCategoryR\x0ctaskCategory\x12\x30\n\x07secrets\x18\x04 \x03(\x0b\x32\x16.flyteidl.admin.SecretR\x07secrets\"\x14\n\x12\x44\x65leteTaskResponse\"\xc4\x01\n\x05\x41gent\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x34\n\x14supported_task_types\x18\x02 \x03(\tB\x02\x18\x01R\x12supportedTaskTypes\x12\x17\n\x07is_sync\x18\x03 \x01(\x08R\x06isSync\x12X\n\x19supported_task_categories\x18\x04 \x03(\x0b\x32\x1c.flyteidl.admin.TaskCategoryR\x17supportedTaskCategories\"<\n\x0cTaskCategory\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x18\n\x07version\x18\x02 \x01(\x05R\x07version\"%\n\x0fGetAgentRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\"?\n\x10GetAgentResponse\x12+\n\x05\x61gent\x18\x01 \x01(\x0b\x32\x15.flyteidl.admin.AgentR\x05\x61gent\"\x13\n\x11ListAgentsRequest\"C\n\x12ListAgentsResponse\x12-\n\x06\x61gents\x18\x01 \x03(\x0b\x32\x15.flyteidl.admin.AgentR\x06\x61gents\"\xdb\x02\n\x15GetTaskMetricsRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x18\n\x07queries\x18\x03 \x03(\tR\x07queries\x12\x39\n\nstart_time\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tstartTime\x12\x35\n\x08\x65nd_time\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\x07\x65ndTime\x12-\n\x04step\x18\x06 \x01(\x0b\x32\x19.google.protobuf.DurationR\x04step\x12\x41\n\rtask_category\x18\x07 \x01(\x0b\x32\x1c.flyteidl.admin.TaskCategoryR\x0ctaskCategory\"X\n\x16GetTaskMetricsResponse\x12>\n\x07results\x18\x01 \x03(\x0b\x32$.flyteidl.core.ExecutionMetricResultR\x07results\"\xc9\x01\n\x12GetTaskLogsRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x14\n\x05lines\x18\x03 \x01(\x04R\x05lines\x12\x14\n\x05token\x18\x04 \x01(\tR\x05token\x12\x41\n\rtask_category\x18\x05 \x01(\x0b\x32\x1c.flyteidl.admin.TaskCategoryR\x0ctaskCategory\"1\n\x19GetTaskLogsResponseHeader\x12\x14\n\x05token\x18\x01 \x01(\tR\x05token\"3\n\x17GetTaskLogsResponseBody\x12\x18\n\x07results\x18\x01 \x03(\tR\x07results\"\xa1\x01\n\x13GetTaskLogsResponse\x12\x43\n\x06header\x18\x01 \x01(\x0b\x32).flyteidl.admin.GetTaskLogsResponseHeaderH\x00R\x06header\x12=\n\x04\x62ody\x18\x02 \x01(\x0b\x32\'.flyteidl.admin.GetTaskLogsResponseBodyH\x00R\x04\x62odyB\x06\n\x04part*b\n\x05State\x12\x15\n\x11RETRYABLE_FAILURE\x10\x00\x12\x15\n\x11PERMANENT_FAILURE\x10\x01\x12\x0b\n\x07PENDING\x10\x02\x12\x0b\n\x07RUNNING\x10\x03\x12\r\n\tSUCCEEDED\x10\x04\x1a\x02\x18\x01\x42\xb6\x01\n\x12\x63om.flyteidl.adminB\nAgentProtoP\x01Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\xa2\x02\x03\x46\x41X\xaa\x02\x0e\x46lyteidl.Admin\xca\x02\x0e\x46lyteidl\\Admin\xe2\x02\x1a\x46lyteidl\\Admin\\GPBMetadata\xea\x02\x0f\x46lyteidl::Adminb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.admin.agent_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\022com.flyteidl.adminB\nAgentProtoP\001Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\242\002\003FAX\252\002\016Flyteidl.Admin\312\002\016Flyteidl\\Admin\342\002\032Flyteidl\\Admin\\GPBMetadata\352\002\017Flyteidl::Admin'
  _STATE._options = None
  _STATE._serialized_options = b'\030\001'
  _TASKEXECUTIONMETADATA_LABELSENTRY._options = None
  _TASKEXECUTIONMETADATA_LABELSENTRY._serialized_options = b'8\001'
  _TASKEXECUTIONMETADATA_ANNOTATIONSENTRY._options = None
  _TASKEXECUTIONMETADATA_ANNOTATIONSENTRY._serialized_options = b'8\001'
  _TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY._options = None
  _TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY._serialized_options = b'8\001'
  _GETTASKREQUEST.fields_by_name['task_type']._options = None
  _GETTASKREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _RESOURCE.fields_by_name['state']._options = None
  _RESOURCE.fields_by_name['state']._serialized_options = b'\030\001'
  _DELETETASKREQUEST.fields_by_name['task_type']._options = None
  _DELETETASKREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _AGENT.fields_by_name['supported_task_types']._options = None
  _AGENT.fields_by_name['supported_task_types']._serialized_options = b'\030\001'
  _GETTASKMETRICSREQUEST.fields_by_name['task_type']._options = None
  _GETTASKMETRICSREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _GETTASKLOGSREQUEST.fields_by_name['task_type']._options = None
  _GETTASKLOGSREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _globals['_STATE']._serialized_start=4586
  _globals['_STATE']._serialized_end=4684
  _globals['_TASKEXECUTIONMETADATA']._serialized_start=351
  _globals['_TASKEXECUTIONMETADATA']._serialized_end=1277
  _globals['_TASKEXECUTIONMETADATA_LABELSENTRY']._serialized_start=1083
  _globals['_TASKEXECUTIONMETADATA_LABELSENTRY']._serialized_end=1140
  _globals['_TASKEXECUTIONMETADATA_ANNOTATIONSENTRY']._serialized_start=1142
  _globals['_TASKEXECUTIONMETADATA_ANNOTATIONSENTRY']._serialized_end=1204
  _globals['_TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY']._serialized_start=1206
  _globals['_TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY']._serialized_end=1277
  _globals['_SECRET']._serialized_start=1279
  _globals['_SECRET']._serialized_end=1309
  _globals['_CREATETASKREQUEST']._serialized_start=1312
  _globals['_CREATETASKREQUEST']._serialized_end=1621
  _globals['_CREATETASKRESPONSE']._serialized_start=1623
  _globals['_CREATETASKRESPONSE']._serialized_end=1680
  _globals['_CREATEREQUESTHEADER']._serialized_start=1683
  _globals['_CREATEREQUESTHEADER']._serialized_end=1996
  _globals['_EXECUTETASKSYNCREQUEST']._serialized_start=1999
  _globals['_EXECUTETASKSYNCREQUEST']._serialized_end=2147
  _globals['_EXECUTETASKSYNCRESPONSEHEADER']._serialized_start=2149
  _globals['_EXECUTETASKSYNCRESPONSEHEADER']._serialized_end=2234
  _globals['_EXECUTETASKSYNCRESPONSE']._serialized_start=2237
  _globals['_EXECUTETASKSYNCRESPONSE']._serialized_end=2397
  _globals['_GETTASKREQUEST']._serialized_start=2400
  _globals['_GETTASKREQUEST']._serialized_end=2603
  _globals['_GETTASKRESPONSE']._serialized_start=2605
  _globals['_GETTASKRESPONSE']._serialized_end=2676
  _globals['_RESOURCE']._serialized_start=2679
  _globals['_RESOURCE']._serialized_end=2986
  _globals['_DELETETASKREQUEST']._serialized_start=2989
  _globals['_DELETETASKREQUEST']._serialized_end=3195
  _globals['_DELETETASKRESPONSE']._serialized_start=3197
  _globals['_DELETETASKRESPONSE']._serialized_end=3217
  _globals['_AGENT']._serialized_start=3220
  _globals['_AGENT']._serialized_end=3416
  _globals['_TASKCATEGORY']._serialized_start=3418
  _globals['_TASKCATEGORY']._serialized_end=3478
  _globals['_GETAGENTREQUEST']._serialized_start=3480
  _globals['_GETAGENTREQUEST']._serialized_end=3517
  _globals['_GETAGENTRESPONSE']._serialized_start=3519
  _globals['_GETAGENTRESPONSE']._serialized_end=3582
  _globals['_LISTAGENTSREQUEST']._serialized_start=3584
  _globals['_LISTAGENTSREQUEST']._serialized_end=3603
  _globals['_LISTAGENTSRESPONSE']._serialized_start=3605
  _globals['_LISTAGENTSRESPONSE']._serialized_end=3672
  _globals['_GETTASKMETRICSREQUEST']._serialized_start=3675
  _globals['_GETTASKMETRICSREQUEST']._serialized_end=4022
  _globals['_GETTASKMETRICSRESPONSE']._serialized_start=4024
  _globals['_GETTASKMETRICSRESPONSE']._serialized_end=4112
  _globals['_GETTASKLOGSREQUEST']._serialized_start=4115
  _globals['_GETTASKLOGSREQUEST']._serialized_end=4316
  _globals['_GETTASKLOGSRESPONSEHEADER']._serialized_start=4318
  _globals['_GETTASKLOGSRESPONSEHEADER']._serialized_end=4367
  _globals['_GETTASKLOGSRESPONSEBODY']._serialized_start=4369
  _globals['_GETTASKLOGSRESPONSEBODY']._serialized_end=4420
  _globals['_GETTASKLOGSRESPONSE']._serialized_start=4423
  _globals['_GETTASKLOGSRESPONSE']._serialized_end=4584
# @@protoc_insertion_point(module_scope)
