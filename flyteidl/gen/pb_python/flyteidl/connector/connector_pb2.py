# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/connector/connector.proto
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


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\"flyteidl/connector/connector.proto\x12\x12\x66lyteidl.connector\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x19\x66lyteidl/core/tasks.proto\x1a\x1c\x66lyteidl/core/workflow.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1d\x66lyteidl/core/execution.proto\x1a\x1b\x66lyteidl/core/metrics.proto\x1a\x1c\x66lyteidl/core/security.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1cgoogle/protobuf/struct.proto\"\xaa\x07\n\x15TaskExecutionMetadata\x12R\n\x11task_execution_id\x18\x01 \x01(\x0b\x32&.flyteidl.core.TaskExecutionIdentifierR\x0ftaskExecutionId\x12\x1c\n\tnamespace\x18\x02 \x01(\tR\tnamespace\x12M\n\x06labels\x18\x03 \x03(\x0b\x32\x35.flyteidl.connector.TaskExecutionMetadata.LabelsEntryR\x06labels\x12\\\n\x0b\x61nnotations\x18\x04 \x03(\x0b\x32:.flyteidl.connector.TaskExecutionMetadata.AnnotationsEntryR\x0b\x61nnotations\x12.\n\x13k8s_service_account\x18\x05 \x01(\tR\x11k8sServiceAccount\x12x\n\x15\x65nvironment_variables\x18\x06 \x03(\x0b\x32\x43.flyteidl.connector.TaskExecutionMetadata.EnvironmentVariablesEntryR\x14\x65nvironmentVariables\x12!\n\x0cmax_attempts\x18\x07 \x01(\x05R\x0bmaxAttempts\x12$\n\rinterruptible\x18\x08 \x01(\x08R\rinterruptible\x12\x46\n\x1finterruptible_failure_threshold\x18\t \x01(\x05R\x1dinterruptibleFailureThreshold\x12>\n\toverrides\x18\n \x01(\x0b\x32 .flyteidl.core.TaskNodeOverridesR\toverrides\x12\x33\n\x08identity\x18\x0b \x01(\x0b\x32\x17.flyteidl.core.IdentityR\x08identity\x1a\x39\n\x0bLabelsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x1a>\n\x10\x41nnotationsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x1aG\n\x19\x45nvironmentVariablesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"\x87\x02\n\x11\x43reateTaskRequest\x12\x31\n\x06inputs\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x06inputs\x12\x37\n\x08template\x18\x02 \x01(\x0b\x32\x1b.flyteidl.core.TaskTemplateR\x08template\x12#\n\routput_prefix\x18\x03 \x01(\tR\x0coutputPrefix\x12\x61\n\x17task_execution_metadata\x18\x04 \x01(\x0b\x32).flyteidl.connector.TaskExecutionMetadataR\x15taskExecutionMetadata\"9\n\x12\x43reateTaskResponse\x12#\n\rresource_meta\x18\x01 \x01(\x0cR\x0cresourceMeta\"\x8b\x02\n\x13\x43reateRequestHeader\x12\x37\n\x08template\x18\x01 \x01(\x0b\x32\x1b.flyteidl.core.TaskTemplateR\x08template\x12#\n\routput_prefix\x18\x02 \x01(\tR\x0coutputPrefix\x12\x61\n\x17task_execution_metadata\x18\x03 \x01(\x0b\x32).flyteidl.connector.TaskExecutionMetadataR\x15taskExecutionMetadata\x12\x33\n\x16max_dataset_size_bytes\x18\x04 \x01(\x03R\x13maxDatasetSizeBytes\"\x98\x01\n\x16\x45xecuteTaskSyncRequest\x12\x41\n\x06header\x18\x01 \x01(\x0b\x32\'.flyteidl.connector.CreateRequestHeaderH\x00R\x06header\x12\x33\n\x06inputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\x06inputsB\x06\n\x04part\"Y\n\x1d\x45xecuteTaskSyncResponseHeader\x12\x38\n\x08resource\x18\x01 \x01(\x0b\x32\x1c.flyteidl.connector.ResourceR\x08resource\"\xa4\x01\n\x17\x45xecuteTaskSyncResponse\x12K\n\x06header\x18\x01 \x01(\x0b\x32\x31.flyteidl.connector.ExecuteTaskSyncResponseHeaderH\x00R\x06header\x12\x35\n\x07outputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\x07outputsB\x05\n\x03res\"\xc8\x01\n\x0eGetTaskRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x45\n\rtask_category\x18\x03 \x01(\x0b\x32 .flyteidl.connector.TaskCategoryR\x0ctaskCategory\x12#\n\routput_prefix\x18\x04 \x01(\tR\x0coutputPrefixJ\x04\x08\x05\x10\x06\"K\n\x0fGetTaskResponse\x12\x38\n\x08resource\x18\x01 \x01(\x0b\x32\x1c.flyteidl.connector.ResourceR\x08resource\"\xd5\x02\n\x08Resource\x12\x33\n\x07outputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x07outputs\x12\x18\n\x07message\x18\x03 \x01(\tR\x07message\x12\x33\n\tlog_links\x18\x04 \x03(\x0b\x32\x16.flyteidl.core.TaskLogR\x08logLinks\x12\x38\n\x05phase\x18\x05 \x01(\x0e\x32\".flyteidl.core.TaskExecution.PhaseR\x05phase\x12\x38\n\x0b\x63ustom_info\x18\x06 \x01(\x0b\x32\x17.google.protobuf.StructR\ncustomInfo\x12K\n\x0f\x63onnector_error\x18\x07 \x01(\x0b\x32\".flyteidl.connector.ConnectorErrorR\x0e\x63onnectorErrorJ\x04\x08\x01\x10\x02\"\xa0\x01\n\x11\x44\x65leteTaskRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x45\n\rtask_category\x18\x03 \x01(\x0b\x32 .flyteidl.connector.TaskCategoryR\x0ctaskCategory\"\x14\n\x12\x44\x65leteTaskResponse\"\x9c\x01\n\tConnector\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x17\n\x07is_sync\x18\x03 \x01(\x08R\x06isSync\x12\\\n\x19supported_task_categories\x18\x04 \x03(\x0b\x32 .flyteidl.connector.TaskCategoryR\x17supportedTaskCategoriesJ\x04\x08\x02\x10\x03\"<\n\x0cTaskCategory\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x18\n\x07version\x18\x02 \x01(\x05R\x07version\")\n\x13GetConnectorRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\"S\n\x14GetConnectorResponse\x12;\n\tconnector\x18\x01 \x01(\x0b\x32\x1d.flyteidl.connector.ConnectorR\tconnector\"\x17\n\x15ListConnectorsRequest\"W\n\x16ListConnectorsResponse\x12=\n\nconnectors\x18\x01 \x03(\x0b\x32\x1d.flyteidl.connector.ConnectorR\nconnectors\"\xdf\x02\n\x15GetTaskMetricsRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x18\n\x07queries\x18\x03 \x03(\tR\x07queries\x12\x39\n\nstart_time\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tstartTime\x12\x35\n\x08\x65nd_time\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\x07\x65ndTime\x12-\n\x04step\x18\x06 \x01(\x0b\x32\x19.google.protobuf.DurationR\x04step\x12\x45\n\rtask_category\x18\x07 \x01(\x0b\x32 .flyteidl.connector.TaskCategoryR\x0ctaskCategory\"X\n\x16GetTaskMetricsResponse\x12>\n\x07results\x18\x01 \x03(\x0b\x32$.flyteidl.core.ExecutionMetricResultR\x07results\"\xcd\x01\n\x12GetTaskLogsRequest\x12\x1f\n\ttask_type\x18\x01 \x01(\tB\x02\x18\x01R\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\x12\x14\n\x05lines\x18\x03 \x01(\x04R\x05lines\x12\x14\n\x05token\x18\x04 \x01(\tR\x05token\x12\x45\n\rtask_category\x18\x05 \x01(\x0b\x32 .flyteidl.connector.TaskCategoryR\x0ctaskCategory\"1\n\x19GetTaskLogsResponseHeader\x12\x14\n\x05token\x18\x01 \x01(\tR\x05token\"3\n\x17GetTaskLogsResponseBody\x12\x18\n\x07results\x18\x01 \x03(\tR\x07results\"\xa9\x01\n\x13GetTaskLogsResponse\x12G\n\x06header\x18\x01 \x01(\x0b\x32-.flyteidl.connector.GetTaskLogsResponseHeaderH\x00R\x06header\x12\x41\n\x04\x62ody\x18\x02 \x01(\x0b\x32+.flyteidl.connector.GetTaskLogsResponseBodyH\x00R\x04\x62odyB\x06\n\x04part\"\xd0\x01\n\x0e\x43onnectorError\x12\x12\n\x04\x63ode\x18\x01 \x01(\tR\x04\x63ode\x12;\n\x04kind\x18\x03 \x01(\x0e\x32\'.flyteidl.connector.ConnectorError.KindR\x04kind\x12?\n\x06origin\x18\x04 \x01(\x0e\x32\'.flyteidl.core.ExecutionError.ErrorKindR\x06origin\",\n\x04Kind\x12\x13\n\x0fNON_RECOVERABLE\x10\x00\x12\x0f\n\x0bRECOVERABLE\x10\x01\x42\xd2\x01\n\x16\x63om.flyteidl.connectorB\x0e\x43onnectorProtoP\x01Z?github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/connector\xa2\x02\x03\x46\x43X\xaa\x02\x12\x46lyteidl.Connector\xca\x02\x12\x46lyteidl\\Connector\xe2\x02\x1e\x46lyteidl\\Connector\\GPBMetadata\xea\x02\x13\x46lyteidl::Connectorb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.connector.connector_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\026com.flyteidl.connectorB\016ConnectorProtoP\001Z?github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/connector\242\002\003FCX\252\002\022Flyteidl.Connector\312\002\022Flyteidl\\Connector\342\002\036Flyteidl\\Connector\\GPBMetadata\352\002\023Flyteidl::Connector'
  _TASKEXECUTIONMETADATA_LABELSENTRY._options = None
  _TASKEXECUTIONMETADATA_LABELSENTRY._serialized_options = b'8\001'
  _TASKEXECUTIONMETADATA_ANNOTATIONSENTRY._options = None
  _TASKEXECUTIONMETADATA_ANNOTATIONSENTRY._serialized_options = b'8\001'
  _TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY._options = None
  _TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY._serialized_options = b'8\001'
  _GETTASKREQUEST.fields_by_name['task_type']._options = None
  _GETTASKREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _DELETETASKREQUEST.fields_by_name['task_type']._options = None
  _DELETETASKREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _GETTASKMETRICSREQUEST.fields_by_name['task_type']._options = None
  _GETTASKMETRICSREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _GETTASKLOGSREQUEST.fields_by_name['task_type']._options = None
  _GETTASKLOGSREQUEST.fields_by_name['task_type']._serialized_options = b'\030\001'
  _globals['_TASKEXECUTIONMETADATA']._serialized_start=363
  _globals['_TASKEXECUTIONMETADATA']._serialized_end=1301
  _globals['_TASKEXECUTIONMETADATA_LABELSENTRY']._serialized_start=1107
  _globals['_TASKEXECUTIONMETADATA_LABELSENTRY']._serialized_end=1164
  _globals['_TASKEXECUTIONMETADATA_ANNOTATIONSENTRY']._serialized_start=1166
  _globals['_TASKEXECUTIONMETADATA_ANNOTATIONSENTRY']._serialized_end=1228
  _globals['_TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY']._serialized_start=1230
  _globals['_TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY']._serialized_end=1301
  _globals['_CREATETASKREQUEST']._serialized_start=1304
  _globals['_CREATETASKREQUEST']._serialized_end=1567
  _globals['_CREATETASKRESPONSE']._serialized_start=1569
  _globals['_CREATETASKRESPONSE']._serialized_end=1626
  _globals['_CREATEREQUESTHEADER']._serialized_start=1629
  _globals['_CREATEREQUESTHEADER']._serialized_end=1896
  _globals['_EXECUTETASKSYNCREQUEST']._serialized_start=1899
  _globals['_EXECUTETASKSYNCREQUEST']._serialized_end=2051
  _globals['_EXECUTETASKSYNCRESPONSEHEADER']._serialized_start=2053
  _globals['_EXECUTETASKSYNCRESPONSEHEADER']._serialized_end=2142
  _globals['_EXECUTETASKSYNCRESPONSE']._serialized_start=2145
  _globals['_EXECUTETASKSYNCRESPONSE']._serialized_end=2309
  _globals['_GETTASKREQUEST']._serialized_start=2312
  _globals['_GETTASKREQUEST']._serialized_end=2512
  _globals['_GETTASKRESPONSE']._serialized_start=2514
  _globals['_GETTASKRESPONSE']._serialized_end=2589
  _globals['_RESOURCE']._serialized_start=2592
  _globals['_RESOURCE']._serialized_end=2933
  _globals['_DELETETASKREQUEST']._serialized_start=2936
  _globals['_DELETETASKREQUEST']._serialized_end=3096
  _globals['_DELETETASKRESPONSE']._serialized_start=3098
  _globals['_DELETETASKRESPONSE']._serialized_end=3118
  _globals['_CONNECTOR']._serialized_start=3121
  _globals['_CONNECTOR']._serialized_end=3277
  _globals['_TASKCATEGORY']._serialized_start=3279
  _globals['_TASKCATEGORY']._serialized_end=3339
  _globals['_GETCONNECTORREQUEST']._serialized_start=3341
  _globals['_GETCONNECTORREQUEST']._serialized_end=3382
  _globals['_GETCONNECTORRESPONSE']._serialized_start=3384
  _globals['_GETCONNECTORRESPONSE']._serialized_end=3467
  _globals['_LISTCONNECTORSREQUEST']._serialized_start=3469
  _globals['_LISTCONNECTORSREQUEST']._serialized_end=3492
  _globals['_LISTCONNECTORSRESPONSE']._serialized_start=3494
  _globals['_LISTCONNECTORSRESPONSE']._serialized_end=3581
  _globals['_GETTASKMETRICSREQUEST']._serialized_start=3584
  _globals['_GETTASKMETRICSREQUEST']._serialized_end=3935
  _globals['_GETTASKMETRICSRESPONSE']._serialized_start=3937
  _globals['_GETTASKMETRICSRESPONSE']._serialized_end=4025
  _globals['_GETTASKLOGSREQUEST']._serialized_start=4028
  _globals['_GETTASKLOGSREQUEST']._serialized_end=4233
  _globals['_GETTASKLOGSRESPONSEHEADER']._serialized_start=4235
  _globals['_GETTASKLOGSRESPONSEHEADER']._serialized_end=4284
  _globals['_GETTASKLOGSRESPONSEBODY']._serialized_start=4286
  _globals['_GETTASKLOGSRESPONSEBODY']._serialized_end=4337
  _globals['_GETTASKLOGSRESPONSE']._serialized_start=4340
  _globals['_GETTASKLOGSRESPONSE']._serialized_end=4509
  _globals['_CONNECTORERROR']._serialized_start=4512
  _globals['_CONNECTORERROR']._serialized_end=4720
  _globals['_CONNECTORERROR_KIND']._serialized_start=4676
  _globals['_CONNECTORERROR_KIND']._serialized_end=4720
# @@protoc_insertion_point(module_scope)
