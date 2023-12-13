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
from flyteidl.core import interface_pb2 as flyteidl_dot_core_dot_interface__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import execution_pb2 as flyteidl_dot_core_dot_execution__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1a\x66lyteidl/admin/agent.proto\x12\x0e\x66lyteidl.admin\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x19\x66lyteidl/core/tasks.proto\x1a\x1d\x66lyteidl/core/interface.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1d\x66lyteidl/core/execution.proto\"\x98\x05\n\x15TaskExecutionMetadata\x12R\n\x11task_execution_id\x18\x01 \x01(\x0b\x32&.flyteidl.core.TaskExecutionIdentifierR\x0ftaskExecutionId\x12\x1c\n\tnamespace\x18\x02 \x01(\tR\tnamespace\x12I\n\x06labels\x18\x03 \x03(\x0b\x32\x31.flyteidl.admin.TaskExecutionMetadata.LabelsEntryR\x06labels\x12X\n\x0b\x61nnotations\x18\x04 \x03(\x0b\x32\x36.flyteidl.admin.TaskExecutionMetadata.AnnotationsEntryR\x0b\x61nnotations\x12.\n\x13k8s_service_account\x18\x05 \x01(\tR\x11k8sServiceAccount\x12t\n\x15\x65nvironment_variables\x18\x06 \x03(\x0b\x32?.flyteidl.admin.TaskExecutionMetadata.EnvironmentVariablesEntryR\x14\x65nvironmentVariables\x1a\x39\n\x0bLabelsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x1a>\n\x10\x41nnotationsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x1aG\n\x19\x45nvironmentVariablesEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"\x83\x02\n\x11\x43reateTaskRequest\x12\x31\n\x06inputs\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x06inputs\x12\x37\n\x08template\x18\x02 \x01(\x0b\x32\x1b.flyteidl.core.TaskTemplateR\x08template\x12#\n\routput_prefix\x18\x03 \x01(\tR\x0coutputPrefix\x12]\n\x17task_execution_metadata\x18\x04 \x01(\x0b\x32%.flyteidl.admin.TaskExecutionMetadataR\x15taskExecutionMetadata\"\xaf\x01\n\x12\x43reateTaskResponse\x12%\n\rresource_meta\x18\x01 \x01(\x0cH\x00R\x0cresourceMeta\x12\x36\n\x08resource\x18\x02 \x01(\x0b\x32\x18.flyteidl.admin.ResourceH\x00R\x08resource\x12\x33\n\tlog_links\x18\x03 \x03(\x0b\x32\x16.flyteidl.core.TaskLogR\x08logLinksB\x05\n\x03res\"R\n\x0eGetTaskRequest\x12\x1b\n\ttask_type\x18\x01 \x01(\tR\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\"|\n\x0fGetTaskResponse\x12\x34\n\x08resource\x18\x01 \x01(\x0b\x32\x18.flyteidl.admin.ResourceR\x08resource\x12\x33\n\tlog_links\x18\x02 \x03(\x0b\x32\x16.flyteidl.core.TaskLogR\x08logLinks\"\x86\x01\n\x08Resource\x12+\n\x05state\x18\x01 \x01(\x0e\x32\x15.flyteidl.admin.StateR\x05state\x12\x33\n\x07outputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x07outputs\x12\x18\n\x07message\x18\x03 \x01(\tR\x07message\"U\n\x11\x44\x65leteTaskRequest\x12\x1b\n\ttask_type\x18\x01 \x01(\tR\x08taskType\x12#\n\rresource_meta\x18\x02 \x01(\x0cR\x0cresourceMeta\"\x14\n\x12\x44\x65leteTaskResponse*^\n\x05State\x12\x15\n\x11RETRYABLE_FAILURE\x10\x00\x12\x15\n\x11PERMANENT_FAILURE\x10\x01\x12\x0b\n\x07PENDING\x10\x02\x12\x0b\n\x07RUNNING\x10\x03\x12\r\n\tSUCCEEDED\x10\x04\x42\xb6\x01\n\x12\x63om.flyteidl.adminB\nAgentProtoP\x01Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\xa2\x02\x03\x46\x41X\xaa\x02\x0e\x46lyteidl.Admin\xca\x02\x0e\x46lyteidl\\Admin\xe2\x02\x1a\x46lyteidl\\Admin\\GPBMetadata\xea\x02\x0f\x46lyteidl::Adminb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.admin.agent_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\022com.flyteidl.adminB\nAgentProtoP\001Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\242\002\003FAX\252\002\016Flyteidl.Admin\312\002\016Flyteidl\\Admin\342\002\032Flyteidl\\Admin\\GPBMetadata\352\002\017Flyteidl::Admin'
  _TASKEXECUTIONMETADATA_LABELSENTRY._options = None
  _TASKEXECUTIONMETADATA_LABELSENTRY._serialized_options = b'8\001'
  _TASKEXECUTIONMETADATA_ANNOTATIONSENTRY._options = None
  _TASKEXECUTIONMETADATA_ANNOTATIONSENTRY._serialized_options = b'8\001'
  _TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY._options = None
  _TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY._serialized_options = b'8\001'
  _globals['_STATE']._serialized_start=1760
  _globals['_STATE']._serialized_end=1854
  _globals['_TASKEXECUTIONMETADATA']._serialized_start=198
  _globals['_TASKEXECUTIONMETADATA']._serialized_end=862
  _globals['_TASKEXECUTIONMETADATA_LABELSENTRY']._serialized_start=668
  _globals['_TASKEXECUTIONMETADATA_LABELSENTRY']._serialized_end=725
  _globals['_TASKEXECUTIONMETADATA_ANNOTATIONSENTRY']._serialized_start=727
  _globals['_TASKEXECUTIONMETADATA_ANNOTATIONSENTRY']._serialized_end=789
  _globals['_TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY']._serialized_start=791
  _globals['_TASKEXECUTIONMETADATA_ENVIRONMENTVARIABLESENTRY']._serialized_end=862
  _globals['_CREATETASKREQUEST']._serialized_start=865
  _globals['_CREATETASKREQUEST']._serialized_end=1124
  _globals['_CREATETASKRESPONSE']._serialized_start=1127
  _globals['_CREATETASKRESPONSE']._serialized_end=1302
  _globals['_GETTASKREQUEST']._serialized_start=1304
  _globals['_GETTASKREQUEST']._serialized_end=1386
  _globals['_GETTASKRESPONSE']._serialized_start=1388
  _globals['_GETTASKRESPONSE']._serialized_end=1512
  _globals['_RESOURCE']._serialized_start=1515
  _globals['_RESOURCE']._serialized_end=1649
  _globals['_DELETETASKREQUEST']._serialized_start=1651
  _globals['_DELETETASKREQUEST']._serialized_end=1736
  _globals['_DELETETASKRESPONSE']._serialized_start=1738
  _globals['_DELETETASKRESPONSE']._serialized_end=1758
# @@protoc_insertion_point(module_scope)
