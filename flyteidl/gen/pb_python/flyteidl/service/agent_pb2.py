# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/service/agent.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.admin import agent_pb2 as flyteidl_dot_admin_dot_agent__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1c\x66lyteidl/service/agent.proto\x12\x10\x66lyteidl.service\x1a\x1a\x66lyteidl/admin/agent.proto2\xda\x02\n\x11\x41syncAgentService\x12U\n\nCreateTask\x12!.flyteidl.admin.CreateTaskRequest\x1a\".flyteidl.admin.CreateTaskResponse\"\x00\x12L\n\x07GetTask\x12\x1e.flyteidl.admin.GetTaskRequest\x1a\x1f.flyteidl.admin.GetTaskResponse\"\x00\x12U\n\nDeleteTask\x12!.flyteidl.admin.DeleteTaskRequest\x1a\".flyteidl.admin.DeleteTaskResponse\"\x00\x12I\n\x06\x44oTask\x12\x1d.flyteidl.admin.DoTaskRequest\x1a\x1e.flyteidl.admin.DoTaskResponse\"\x00\x42\xc2\x01\n\x14\x63om.flyteidl.serviceB\nAgentProtoP\x01Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service\xa2\x02\x03\x46SX\xaa\x02\x10\x46lyteidl.Service\xca\x02\x10\x46lyteidl\\Service\xe2\x02\x1c\x46lyteidl\\Service\\GPBMetadata\xea\x02\x11\x46lyteidl::Serviceb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.service.agent_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\024com.flyteidl.serviceB\nAgentProtoP\001Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service\242\002\003FSX\252\002\020Flyteidl.Service\312\002\020Flyteidl\\Service\342\002\034Flyteidl\\Service\\GPBMetadata\352\002\021Flyteidl::Service'
  _globals['_ASYNCAGENTSERVICE']._serialized_start=79
  _globals['_ASYNCAGENTSERVICE']._serialized_end=425
# @@protoc_insertion_point(module_scope)
