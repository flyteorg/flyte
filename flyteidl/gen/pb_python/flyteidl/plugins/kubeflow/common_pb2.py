# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/plugins/kubeflow/common.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n&flyteidl/plugins/kubeflow/common.proto\x12\x19\x66lyteidl.plugins.kubeflow\"\xfa\x01\n\tRunPolicy\x12S\n\x10\x63lean_pod_policy\x18\x01 \x01(\x0e\x32).flyteidl.plugins.kubeflow.CleanPodPolicyR\x0e\x63leanPodPolicy\x12;\n\x1attl_seconds_after_finished\x18\x02 \x01(\x05R\x17ttlSecondsAfterFinished\x12\x36\n\x17\x61\x63tive_deadline_seconds\x18\x03 \x01(\x05R\x15\x61\x63tiveDeadlineSeconds\x12#\n\rbackoff_limit\x18\x04 \x01(\x05R\x0c\x62\x61\x63koffLimit*c\n\rRestartPolicy\x12\x18\n\x14RESTART_POLICY_NEVER\x10\x00\x12\x1d\n\x19RESTART_POLICY_ON_FAILURE\x10\x01\x12\x19\n\x15RESTART_POLICY_ALWAYS\x10\x02*`\n\x0e\x43leanPodPolicy\x12\x18\n\x14\x43LEANPOD_POLICY_NONE\x10\x00\x12\x1b\n\x17\x43LEANPOD_POLICY_RUNNING\x10\x01\x12\x17\n\x13\x43LEANPOD_POLICY_ALL\x10\x02\x42\xf1\x01\n\x1d\x63om.flyteidl.plugins.kubeflowB\x0b\x43ommonProtoP\x01Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins\xa2\x02\x03\x46PK\xaa\x02\x19\x46lyteidl.Plugins.Kubeflow\xca\x02\x19\x46lyteidl\\Plugins\\Kubeflow\xe2\x02%Flyteidl\\Plugins\\Kubeflow\\GPBMetadata\xea\x02\x1b\x46lyteidl::Plugins::Kubeflowb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.plugins.kubeflow.common_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\035com.flyteidl.plugins.kubeflowB\013CommonProtoP\001Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins\242\002\003FPK\252\002\031Flyteidl.Plugins.Kubeflow\312\002\031Flyteidl\\Plugins\\Kubeflow\342\002%Flyteidl\\Plugins\\Kubeflow\\GPBMetadata\352\002\033Flyteidl::Plugins::Kubeflow'
  _globals['_RESTARTPOLICY']._serialized_start=322
  _globals['_RESTARTPOLICY']._serialized_end=421
  _globals['_CLEANPODPOLICY']._serialized_start=423
  _globals['_CLEANPODPOLICY']._serialized_end=519
  _globals['_RUNPOLICY']._serialized_start=70
  _globals['_RUNPOLICY']._serialized_end=320
# @@protoc_insertion_point(module_scope)
