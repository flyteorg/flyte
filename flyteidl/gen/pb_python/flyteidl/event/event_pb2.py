# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/event/event.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2
from flyteidl.core import compiler_pb2 as flyteidl_dot_core_dot_compiler__pb2
from flyteidl.core import execution_pb2 as flyteidl_dot_core_dot_execution__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import catalog_pb2 as flyteidl_dot_core_dot_catalog__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1a\x66lyteidl/event/event.proto\x12\x0e\x66lyteidl.event\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x1c\x66lyteidl/core/compiler.proto\x1a\x1d\x66lyteidl/core/execution.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1b\x66lyteidl/core/catalog.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1cgoogle/protobuf/struct.proto\"\xaa\x03\n\x16WorkflowExecutionEvent\x12M\n\x0c\x65xecution_id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x0b\x65xecutionId\x12\x1f\n\x0bproducer_id\x18\x02 \x01(\tR\nproducerId\x12<\n\x05phase\x18\x03 \x01(\x0e\x32&.flyteidl.core.WorkflowExecution.PhaseR\x05phase\x12;\n\x0boccurred_at\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\noccurredAt\x12\x1f\n\noutput_uri\x18\x05 \x01(\tH\x00R\toutputUri\x12\x35\n\x05\x65rror\x18\x06 \x01(\x0b\x32\x1d.flyteidl.core.ExecutionErrorH\x00R\x05\x65rror\x12<\n\x0boutput_data\x18\x07 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\noutputDataB\x0f\n\routput_result\"\xb4\n\n\x12NodeExecutionEvent\x12\x36\n\x02id\x18\x01 \x01(\x0b\x32&.flyteidl.core.NodeExecutionIdentifierR\x02id\x12\x1f\n\x0bproducer_id\x18\x02 \x01(\tR\nproducerId\x12\x38\n\x05phase\x18\x03 \x01(\x0e\x32\".flyteidl.core.NodeExecution.PhaseR\x05phase\x12;\n\x0boccurred_at\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\noccurredAt\x12\x1d\n\tinput_uri\x18\x05 \x01(\tH\x00R\x08inputUri\x12:\n\ninput_data\x18\x14 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\tinputData\x12\x1f\n\noutput_uri\x18\x06 \x01(\tH\x01R\toutputUri\x12\x35\n\x05\x65rror\x18\x07 \x01(\x0b\x32\x1d.flyteidl.core.ExecutionErrorH\x01R\x05\x65rror\x12<\n\x0boutput_data\x18\x0f \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x01R\noutputData\x12\\\n\x16workflow_node_metadata\x18\x08 \x01(\x0b\x32$.flyteidl.event.WorkflowNodeMetadataH\x02R\x14workflowNodeMetadata\x12P\n\x12task_node_metadata\x18\x0e \x01(\x0b\x32 .flyteidl.event.TaskNodeMetadataH\x02R\x10taskNodeMetadata\x12]\n\x14parent_task_metadata\x18\t \x01(\x0b\x32+.flyteidl.event.ParentTaskExecutionMetadataR\x12parentTaskMetadata\x12]\n\x14parent_node_metadata\x18\n \x01(\x0b\x32+.flyteidl.event.ParentNodeExecutionMetadataR\x12parentNodeMetadata\x12\x1f\n\x0bretry_group\x18\x0b \x01(\tR\nretryGroup\x12 \n\x0cspec_node_id\x18\x0c \x01(\tR\nspecNodeId\x12\x1b\n\tnode_name\x18\r \x01(\tR\x08nodeName\x12#\n\revent_version\x18\x10 \x01(\x05R\x0c\x65ventVersion\x12\x1b\n\tis_parent\x18\x11 \x01(\x08R\x08isParent\x12\x1d\n\nis_dynamic\x18\x12 \x01(\x08R\tisDynamic\x12\x19\n\x08\x64\x65\x63k_uri\x18\x13 \x01(\tR\x07\x64\x65\x63kUri\x12;\n\x0breported_at\x18\x15 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\nreportedAt\x12\x19\n\x08is_array\x18\x16 \x01(\x08R\x07isArray\x12>\n\rtarget_entity\x18\x17 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x0ctargetEntity\x12-\n\x13is_in_dynamic_chain\x18\x18 \x01(\x08R\x10isInDynamicChain\x12\x19\n\x08is_eager\x18\x19 \x01(\x08R\x07isEagerB\r\n\x0binput_valueB\x0f\n\routput_resultB\x11\n\x0ftarget_metadata\"e\n\x14WorkflowNodeMetadata\x12M\n\x0c\x65xecution_id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x0b\x65xecutionId\"\xf1\x02\n\x10TaskNodeMetadata\x12\x44\n\x0c\x63\x61\x63he_status\x18\x01 \x01(\x0e\x32!.flyteidl.core.CatalogCacheStatusR\x0b\x63\x61\x63heStatus\x12?\n\x0b\x63\x61talog_key\x18\x02 \x01(\x0b\x32\x1e.flyteidl.core.CatalogMetadataR\ncatalogKey\x12W\n\x12reservation_status\x18\x03 \x01(\x0e\x32(.flyteidl.core.CatalogReservation.StatusR\x11reservationStatus\x12%\n\x0e\x63heckpoint_uri\x18\x04 \x01(\tR\rcheckpointUri\x12V\n\x10\x64ynamic_workflow\x18\x10 \x01(\x0b\x32+.flyteidl.event.DynamicWorkflowNodeMetadataR\x0f\x64ynamicWorkflow\"\xce\x01\n\x1b\x44ynamicWorkflowNodeMetadata\x12)\n\x02id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x02id\x12S\n\x11\x63ompiled_workflow\x18\x02 \x01(\x0b\x32&.flyteidl.core.CompiledWorkflowClosureR\x10\x63ompiledWorkflow\x12/\n\x14\x64ynamic_job_spec_uri\x18\x03 \x01(\tR\x11\x64ynamicJobSpecUri\"U\n\x1bParentTaskExecutionMetadata\x12\x36\n\x02id\x18\x01 \x01(\x0b\x32&.flyteidl.core.TaskExecutionIdentifierR\x02id\"6\n\x1bParentNodeExecutionMetadata\x12\x17\n\x07node_id\x18\x01 \x01(\tR\x06nodeId\"b\n\x0b\x45ventReason\x12\x16\n\x06reason\x18\x01 \x01(\tR\x06reason\x12;\n\x0boccurred_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\noccurredAt\"\xd3\x08\n\x12TaskExecutionEvent\x12\x32\n\x07task_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x06taskId\x12_\n\x18parent_node_execution_id\x18\x02 \x01(\x0b\x32&.flyteidl.core.NodeExecutionIdentifierR\x15parentNodeExecutionId\x12#\n\rretry_attempt\x18\x03 \x01(\rR\x0cretryAttempt\x12\x38\n\x05phase\x18\x04 \x01(\x0e\x32\".flyteidl.core.TaskExecution.PhaseR\x05phase\x12\x1f\n\x0bproducer_id\x18\x05 \x01(\tR\nproducerId\x12*\n\x04logs\x18\x06 \x03(\x0b\x32\x16.flyteidl.core.TaskLogR\x04logs\x12;\n\x0boccurred_at\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\noccurredAt\x12\x1d\n\tinput_uri\x18\x08 \x01(\tH\x00R\x08inputUri\x12:\n\ninput_data\x18\x13 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\tinputData\x12\x1f\n\noutput_uri\x18\t \x01(\tH\x01R\toutputUri\x12\x35\n\x05\x65rror\x18\n \x01(\x0b\x32\x1d.flyteidl.core.ExecutionErrorH\x01R\x05\x65rror\x12<\n\x0boutput_data\x18\x11 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x01R\noutputData\x12\x38\n\x0b\x63ustom_info\x18\x0b \x01(\x0b\x32\x17.google.protobuf.StructR\ncustomInfo\x12#\n\rphase_version\x18\x0c \x01(\rR\x0cphaseVersion\x12\x1a\n\x06reason\x18\r \x01(\tB\x02\x18\x01R\x06reason\x12\x35\n\x07reasons\x18\x15 \x03(\x0b\x32\x1b.flyteidl.event.EventReasonR\x07reasons\x12\x1b\n\ttask_type\x18\x0e \x01(\tR\x08taskType\x12\x41\n\x08metadata\x18\x10 \x01(\x0b\x32%.flyteidl.event.TaskExecutionMetadataR\x08metadata\x12#\n\revent_version\x18\x12 \x01(\x05R\x0c\x65ventVersion\x12;\n\x0breported_at\x18\x14 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\nreportedAt\x12:\n\x0blog_context\x18\x16 \x01(\x0b\x32\x19.flyteidl.core.LogContextR\nlogContextB\r\n\x0binput_valueB\x0f\n\routput_result\"\x85\x04\n\x14\x45xternalResourceInfo\x12\x1f\n\x0b\x65xternal_id\x18\x01 \x01(\tR\nexternalId\x12\x14\n\x05index\x18\x02 \x01(\rR\x05index\x12#\n\rretry_attempt\x18\x03 \x01(\rR\x0cretryAttempt\x12\x38\n\x05phase\x18\x04 \x01(\x0e\x32\".flyteidl.core.TaskExecution.PhaseR\x05phase\x12\x44\n\x0c\x63\x61\x63he_status\x18\x05 \x01(\x0e\x32!.flyteidl.core.CatalogCacheStatusR\x0b\x63\x61\x63heStatus\x12*\n\x04logs\x18\x06 \x03(\x0b\x32\x16.flyteidl.core.TaskLogR\x04logs\x12\\\n\x16workflow_node_metadata\x18\x07 \x01(\x0b\x32$.flyteidl.event.WorkflowNodeMetadataH\x00R\x14workflowNodeMetadata\x12\x38\n\x0b\x63ustom_info\x18\x08 \x01(\x0b\x32\x17.google.protobuf.StructR\ncustomInfo\x12:\n\x0blog_context\x18\t \x01(\x0b\x32\x19.flyteidl.core.LogContextR\nlogContextB\x11\n\x0ftarget_metadata\"[\n\x10ResourcePoolInfo\x12)\n\x10\x61llocation_token\x18\x01 \x01(\tR\x0f\x61llocationToken\x12\x1c\n\tnamespace\x18\x02 \x01(\tR\tnamespace\"\x9d\x03\n\x15TaskExecutionMetadata\x12%\n\x0egenerated_name\x18\x01 \x01(\tR\rgeneratedName\x12S\n\x12\x65xternal_resources\x18\x02 \x03(\x0b\x32$.flyteidl.event.ExternalResourceInfoR\x11\x65xternalResources\x12N\n\x12resource_pool_info\x18\x03 \x03(\x0b\x32 .flyteidl.event.ResourcePoolInfoR\x10resourcePoolInfo\x12+\n\x11plugin_identifier\x18\x04 \x01(\tR\x10pluginIdentifier\x12Z\n\x0einstance_class\x18\x10 \x01(\x0e\x32\x33.flyteidl.event.TaskExecutionMetadata.InstanceClassR\rinstanceClass\"/\n\rInstanceClass\x12\x0b\n\x07\x44\x45\x46\x41ULT\x10\x00\x12\x11\n\rINTERRUPTIBLE\x10\x01\x42\xb6\x01\n\x12\x63om.flyteidl.eventB\nEventProtoP\x01Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event\xa2\x02\x03\x46\x45X\xaa\x02\x0e\x46lyteidl.Event\xca\x02\x0e\x46lyteidl\\Event\xe2\x02\x1a\x46lyteidl\\Event\\GPBMetadata\xea\x02\x0f\x46lyteidl::Eventb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.event.event_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\022com.flyteidl.eventB\nEventProtoP\001Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event\242\002\003FEX\252\002\016Flyteidl.Event\312\002\016Flyteidl\\Event\342\002\032Flyteidl\\Event\\GPBMetadata\352\002\017Flyteidl::Event'
  _TASKEXECUTIONEVENT.fields_by_name['reason']._options = None
  _TASKEXECUTIONEVENT.fields_by_name['reason']._serialized_options = b'\030\001'
  _globals['_WORKFLOWEXECUTIONEVENT']._serialized_start=262
  _globals['_WORKFLOWEXECUTIONEVENT']._serialized_end=688
  _globals['_NODEEXECUTIONEVENT']._serialized_start=691
  _globals['_NODEEXECUTIONEVENT']._serialized_end=2023
  _globals['_WORKFLOWNODEMETADATA']._serialized_start=2025
  _globals['_WORKFLOWNODEMETADATA']._serialized_end=2126
  _globals['_TASKNODEMETADATA']._serialized_start=2129
  _globals['_TASKNODEMETADATA']._serialized_end=2498
  _globals['_DYNAMICWORKFLOWNODEMETADATA']._serialized_start=2501
  _globals['_DYNAMICWORKFLOWNODEMETADATA']._serialized_end=2707
  _globals['_PARENTTASKEXECUTIONMETADATA']._serialized_start=2709
  _globals['_PARENTTASKEXECUTIONMETADATA']._serialized_end=2794
  _globals['_PARENTNODEEXECUTIONMETADATA']._serialized_start=2796
  _globals['_PARENTNODEEXECUTIONMETADATA']._serialized_end=2850
  _globals['_EVENTREASON']._serialized_start=2852
  _globals['_EVENTREASON']._serialized_end=2950
  _globals['_TASKEXECUTIONEVENT']._serialized_start=2953
  _globals['_TASKEXECUTIONEVENT']._serialized_end=4060
  _globals['_EXTERNALRESOURCEINFO']._serialized_start=4063
  _globals['_EXTERNALRESOURCEINFO']._serialized_end=4580
  _globals['_RESOURCEPOOLINFO']._serialized_start=4582
  _globals['_RESOURCEPOOLINFO']._serialized_end=4673
  _globals['_TASKEXECUTIONMETADATA']._serialized_start=4676
  _globals['_TASKEXECUTIONMETADATA']._serialized_end=5089
  _globals['_TASKEXECUTIONMETADATA_INSTANCECLASS']._serialized_start=5042
  _globals['_TASKEXECUTIONMETADATA_INSTANCECLASS']._serialized_end=5089
# @@protoc_insertion_point(module_scope)
