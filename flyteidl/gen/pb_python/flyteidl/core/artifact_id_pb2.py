# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/core/artifact_id.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1f\x66lyteidl/core/artifact_id.proto\x12\rflyteidl.core\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1e\x66lyteidl/core/identifier.proto\"e\n\x0b\x41rtifactKey\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x16\n\x06\x64omain\x18\x02 \x01(\tR\x06\x64omain\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x10\n\x03org\x18\x04 \x01(\tR\x03org\"\xb9\x01\n\x13\x41rtifactBindingData\x12\x14\n\x05index\x18\x01 \x01(\rR\x05index\x12%\n\rpartition_key\x18\x02 \x01(\tH\x00R\x0cpartitionKey\x12\x35\n\x16\x62ind_to_time_partition\x18\x03 \x01(\x08H\x00R\x13\x62indToTimePartition\x12\x1c\n\ttransform\x18\x04 \x01(\tR\ttransformB\x10\n\x0epartition_data\"$\n\x10InputBindingData\x12\x10\n\x03var\x18\x01 \x01(\tR\x03var\"\x92\x02\n\nLabelValue\x12#\n\x0cstatic_value\x18\x01 \x01(\tH\x00R\x0bstaticValue\x12;\n\ntime_value\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampH\x00R\ttimeValue\x12Q\n\x11triggered_binding\x18\x03 \x01(\x0b\x32\".flyteidl.core.ArtifactBindingDataH\x00R\x10triggeredBinding\x12\x46\n\rinput_binding\x18\x04 \x01(\x0b\x32\x1f.flyteidl.core.InputBindingDataH\x00R\x0cinputBindingB\x07\n\x05value\"\x9d\x01\n\nPartitions\x12:\n\x05value\x18\x01 \x03(\x0b\x32$.flyteidl.core.Partitions.ValueEntryR\x05value\x1aS\n\nValueEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12/\n\x05value\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LabelValueR\x05value:\x02\x38\x01\"@\n\rTimePartition\x12/\n\x05value\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.LabelValueR\x05value\"\xe5\x01\n\nArtifactID\x12=\n\x0c\x61rtifact_key\x18\x01 \x01(\x0b\x32\x1a.flyteidl.core.ArtifactKeyR\x0b\x61rtifactKey\x12\x18\n\x07version\x18\x02 \x01(\tR\x07version\x12\x39\n\npartitions\x18\x03 \x01(\x0b\x32\x19.flyteidl.core.PartitionsR\npartitions\x12\x43\n\x0etime_partition\x18\x04 \x01(\x0b\x32\x1c.flyteidl.core.TimePartitionR\rtimePartition\"}\n\x0b\x41rtifactTag\x12=\n\x0c\x61rtifact_key\x18\x01 \x01(\x0b\x32\x1a.flyteidl.core.ArtifactKeyR\x0b\x61rtifactKey\x12/\n\x05value\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LabelValueR\x05value\"\xf0\x01\n\rArtifactQuery\x12<\n\x0b\x61rtifact_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.ArtifactIDH\x00R\nartifactId\x12?\n\x0c\x61rtifact_tag\x18\x02 \x01(\x0b\x32\x1a.flyteidl.core.ArtifactTagH\x00R\x0b\x61rtifactTag\x12\x12\n\x03uri\x18\x03 \x01(\tH\x00R\x03uri\x12>\n\x07\x62inding\x18\x04 \x01(\x0b\x32\".flyteidl.core.ArtifactBindingDataH\x00R\x07\x62indingB\x0c\n\nidentifierB\xb5\x01\n\x11\x63om.flyteidl.coreB\x0f\x41rtifactIdProtoP\x01Z:github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core\xa2\x02\x03\x46\x43X\xaa\x02\rFlyteidl.Core\xca\x02\rFlyteidl\\Core\xe2\x02\x19\x46lyteidl\\Core\\GPBMetadata\xea\x02\x0e\x46lyteidl::Coreb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.core.artifact_id_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\021com.flyteidl.coreB\017ArtifactIdProtoP\001Z:github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core\242\002\003FCX\252\002\rFlyteidl.Core\312\002\rFlyteidl\\Core\342\002\031Flyteidl\\Core\\GPBMetadata\352\002\016Flyteidl::Core'
  _PARTITIONS_VALUEENTRY._options = None
  _PARTITIONS_VALUEENTRY._serialized_options = b'8\001'
  _globals['_ARTIFACTKEY']._serialized_start=115
  _globals['_ARTIFACTKEY']._serialized_end=216
  _globals['_ARTIFACTBINDINGDATA']._serialized_start=219
  _globals['_ARTIFACTBINDINGDATA']._serialized_end=404
  _globals['_INPUTBINDINGDATA']._serialized_start=406
  _globals['_INPUTBINDINGDATA']._serialized_end=442
  _globals['_LABELVALUE']._serialized_start=445
  _globals['_LABELVALUE']._serialized_end=719
  _globals['_PARTITIONS']._serialized_start=722
  _globals['_PARTITIONS']._serialized_end=879
  _globals['_PARTITIONS_VALUEENTRY']._serialized_start=796
  _globals['_PARTITIONS_VALUEENTRY']._serialized_end=879
  _globals['_TIMEPARTITION']._serialized_start=881
  _globals['_TIMEPARTITION']._serialized_end=945
  _globals['_ARTIFACTID']._serialized_start=948
  _globals['_ARTIFACTID']._serialized_end=1177
  _globals['_ARTIFACTTAG']._serialized_start=1179
  _globals['_ARTIFACTTAG']._serialized_end=1304
  _globals['_ARTIFACTQUERY']._serialized_start=1307
  _globals['_ARTIFACTQUERY']._serialized_end=1547
# @@protoc_insertion_point(module_scope)
