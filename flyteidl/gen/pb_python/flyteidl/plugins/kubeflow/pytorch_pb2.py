# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/plugins/kubeflow/pytorch.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.core import tasks_pb2 as flyteidl_dot_core_dot_tasks__pb2
from flyteidl.plugins.kubeflow import common_pb2 as flyteidl_dot_plugins_dot_kubeflow_dot_common__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\'flyteidl/plugins/kubeflow/pytorch.proto\x12\x19\x66lyteidl.plugins.kubeflow\x1a\x19\x66lyteidl/core/tasks.proto\x1a&flyteidl/plugins/kubeflow/common.proto\"\xc1\x01\n\rElasticConfig\x12!\n\x0crdzv_backend\x18\x01 \x01(\tR\x0brdzvBackend\x12!\n\x0cmin_replicas\x18\x02 \x01(\x05R\x0bminReplicas\x12!\n\x0cmax_replicas\x18\x03 \x01(\x05R\x0bmaxReplicas\x12$\n\x0enproc_per_node\x18\x04 \x01(\x05R\x0cnprocPerNode\x12!\n\x0cmax_restarts\x18\x05 \x01(\x05R\x0bmaxRestarts\"\x8c\x03\n\x1e\x44istributedPyTorchTrainingTask\x12i\n\x0fworker_replicas\x18\x01 \x01(\x0b\x32@.flyteidl.plugins.kubeflow.DistributedPyTorchTrainingReplicaSpecR\x0eworkerReplicas\x12i\n\x0fmaster_replicas\x18\x02 \x01(\x0b\x32@.flyteidl.plugins.kubeflow.DistributedPyTorchTrainingReplicaSpecR\x0emasterReplicas\x12\x43\n\nrun_policy\x18\x03 \x01(\x0b\x32$.flyteidl.plugins.kubeflow.RunPolicyR\trunPolicy\x12O\n\x0e\x65lastic_config\x18\x04 \x01(\x0b\x32(.flyteidl.plugins.kubeflow.ElasticConfigR\relasticConfig\"\xa0\x03\n%DistributedPyTorchTrainingReplicaSpec\x12\x1a\n\x08replicas\x18\x01 \x01(\x05R\x08replicas\x12\x14\n\x05image\x18\x02 \x01(\tR\x05image\x12\x36\n\tresources\x18\x03 \x01(\x0b\x32\x18.flyteidl.core.ResourcesR\tresources\x12O\n\x0erestart_policy\x18\x04 \x01(\x0e\x32(.flyteidl.plugins.kubeflow.RestartPolicyR\rrestartPolicy\x12z\n\x0enode_selectors\x18\x05 \x03(\x0b\x32S.flyteidl.plugins.kubeflow.DistributedPyTorchTrainingReplicaSpec.NodeSelectorsEntryR\rnodeSelectors\x1a@\n\x12NodeSelectorsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\x42\xf2\x01\n\x1d\x63om.flyteidl.plugins.kubeflowB\x0cPytorchProtoP\x01Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins\xa2\x02\x03\x46PK\xaa\x02\x19\x46lyteidl.Plugins.Kubeflow\xca\x02\x19\x46lyteidl\\Plugins\\Kubeflow\xe2\x02%Flyteidl\\Plugins\\Kubeflow\\GPBMetadata\xea\x02\x1b\x46lyteidl::Plugins::Kubeflowb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.plugins.kubeflow.pytorch_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\035com.flyteidl.plugins.kubeflowB\014PytorchProtoP\001Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins\242\002\003FPK\252\002\031Flyteidl.Plugins.Kubeflow\312\002\031Flyteidl\\Plugins\\Kubeflow\342\002%Flyteidl\\Plugins\\Kubeflow\\GPBMetadata\352\002\033Flyteidl::Plugins::Kubeflow'
  _DISTRIBUTEDPYTORCHTRAININGREPLICASPEC_NODESELECTORSENTRY._options = None
  _DISTRIBUTEDPYTORCHTRAININGREPLICASPEC_NODESELECTORSENTRY._serialized_options = b'8\001'
  _globals['_ELASTICCONFIG']._serialized_start=138
  _globals['_ELASTICCONFIG']._serialized_end=331
  _globals['_DISTRIBUTEDPYTORCHTRAININGTASK']._serialized_start=334
  _globals['_DISTRIBUTEDPYTORCHTRAININGTASK']._serialized_end=730
  _globals['_DISTRIBUTEDPYTORCHTRAININGREPLICASPEC']._serialized_start=733
  _globals['_DISTRIBUTEDPYTORCHTRAININGREPLICASPEC']._serialized_end=1149
  _globals['_DISTRIBUTEDPYTORCHTRAININGREPLICASPEC_NODESELECTORSENTRY']._serialized_start=1085
  _globals['_DISTRIBUTEDPYTORCHTRAININGREPLICASPEC_NODESELECTORSENTRY']._serialized_end=1149
# @@protoc_insertion_point(module_scope)
