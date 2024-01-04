# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/artifact/artifacts.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import any_pb2 as google_dot_protobuf_dot_any__pb2
from google.api import annotations_pb2 as google_dot_api_dot_annotations__pb2
from flyteidl.admin import launch_plan_pb2 as flyteidl_dot_admin_dot_launch__plan__pb2
from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2
from flyteidl.core import types_pb2 as flyteidl_dot_core_dot_types__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import artifact_id_pb2 as flyteidl_dot_core_dot_artifact__id__pb2
from flyteidl.core import interface_pb2 as flyteidl_dot_core_dot_interface__pb2
from flyteidl.event import cloudevents_pb2 as flyteidl_dot_event_dot_cloudevents__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n!flyteidl/artifact/artifacts.proto\x12\x11\x66lyteidl.artifact\x1a\x19google/protobuf/any.proto\x1a\x1cgoogle/api/annotations.proto\x1a flyteidl/admin/launch_plan.proto\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x19\x66lyteidl/core/types.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1f\x66lyteidl/core/artifact_id.proto\x1a\x1d\x66lyteidl/core/interface.proto\x1a flyteidl/event/cloudevents.proto\"\xca\x01\n\x08\x41rtifact\x12:\n\x0b\x61rtifact_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.ArtifactIDR\nartifactId\x12\x33\n\x04spec\x18\x02 \x01(\x0b\x32\x1f.flyteidl.artifact.ArtifactSpecR\x04spec\x12\x12\n\x04tags\x18\x03 \x03(\tR\x04tags\x12\x39\n\x06source\x18\x04 \x01(\x0b\x32!.flyteidl.artifact.ArtifactSourceR\x06source\"\x8b\x03\n\x15\x43reateArtifactRequest\x12=\n\x0c\x61rtifact_key\x18\x01 \x01(\x0b\x32\x1a.flyteidl.core.ArtifactKeyR\x0b\x61rtifactKey\x12\x18\n\x07version\x18\x03 \x01(\tR\x07version\x12\x33\n\x04spec\x18\x02 \x01(\x0b\x32\x1f.flyteidl.artifact.ArtifactSpecR\x04spec\x12X\n\npartitions\x18\x04 \x03(\x0b\x32\x38.flyteidl.artifact.CreateArtifactRequest.PartitionsEntryR\npartitions\x12\x10\n\x03tag\x18\x05 \x01(\tR\x03tag\x12\x39\n\x06source\x18\x06 \x01(\x0b\x32!.flyteidl.artifact.ArtifactSourceR\x06source\x1a=\n\x0fPartitionsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"\xfb\x01\n\x0e\x41rtifactSource\x12Y\n\x12workflow_execution\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x11workflowExecution\x12\x17\n\x07node_id\x18\x02 \x01(\tR\x06nodeId\x12\x32\n\x07task_id\x18\x03 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x06taskId\x12#\n\rretry_attempt\x18\x04 \x01(\rR\x0cretryAttempt\x12\x1c\n\tprincipal\x18\x05 \x01(\tR\tprincipal\"\xf9\x01\n\x0c\x41rtifactSpec\x12,\n\x05value\x18\x01 \x01(\x0b\x32\x16.flyteidl.core.LiteralR\x05value\x12.\n\x04type\x18\x02 \x01(\x0b\x32\x1a.flyteidl.core.LiteralTypeR\x04type\x12+\n\x11short_description\x18\x03 \x01(\tR\x10shortDescription\x12\x39\n\ruser_metadata\x18\x04 \x01(\x0b\x32\x14.google.protobuf.AnyR\x0cuserMetadata\x12#\n\rmetadata_type\x18\x05 \x01(\tR\x0cmetadataType\"Q\n\x16\x43reateArtifactResponse\x12\x37\n\x08\x61rtifact\x18\x01 \x01(\x0b\x32\x1b.flyteidl.artifact.ArtifactR\x08\x61rtifact\"b\n\x12GetArtifactRequest\x12\x32\n\x05query\x18\x01 \x01(\x0b\x32\x1c.flyteidl.core.ArtifactQueryR\x05query\x12\x18\n\x07\x64\x65tails\x18\x02 \x01(\x08R\x07\x64\x65tails\"N\n\x13GetArtifactResponse\x12\x37\n\x08\x61rtifact\x18\x01 \x01(\x0b\x32\x1b.flyteidl.artifact.ArtifactR\x08\x61rtifact\"`\n\rSearchOptions\x12+\n\x11strict_partitions\x18\x01 \x01(\x08R\x10strictPartitions\x12\"\n\rlatest_by_key\x18\x02 \x01(\x08R\x0blatestByKey\"\xb2\x02\n\x16SearchArtifactsRequest\x12=\n\x0c\x61rtifact_key\x18\x01 \x01(\x0b\x32\x1a.flyteidl.core.ArtifactKeyR\x0b\x61rtifactKey\x12\x39\n\npartitions\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.PartitionsR\npartitions\x12\x1c\n\tprincipal\x18\x03 \x01(\tR\tprincipal\x12\x18\n\x07version\x18\x04 \x01(\tR\x07version\x12:\n\x07options\x18\x05 \x01(\x0b\x32 .flyteidl.artifact.SearchOptionsR\x07options\x12\x14\n\x05token\x18\x06 \x01(\tR\x05token\x12\x14\n\x05limit\x18\x07 \x01(\x05R\x05limit\"j\n\x17SearchArtifactsResponse\x12\x39\n\tartifacts\x18\x01 \x03(\x0b\x32\x1b.flyteidl.artifact.ArtifactR\tartifacts\x12\x14\n\x05token\x18\x02 \x01(\tR\x05token\"\xdc\x01\n\x19\x46indByWorkflowExecRequest\x12\x43\n\x07\x65xec_id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x06\x65xecId\x12T\n\tdirection\x18\x02 \x01(\x0e\x32\x36.flyteidl.artifact.FindByWorkflowExecRequest.DirectionR\tdirection\"$\n\tDirection\x12\n\n\x06INPUTS\x10\x00\x12\x0b\n\x07OUTPUTS\x10\x01\"\x7f\n\rAddTagRequest\x12:\n\x0b\x61rtifact_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.ArtifactIDR\nartifactId\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value\x12\x1c\n\toverwrite\x18\x03 \x01(\x08R\toverwrite\"\x10\n\x0e\x41\x64\x64TagResponse\"b\n\x14\x43reateTriggerRequest\x12J\n\x13trigger_launch_plan\x18\x01 \x01(\x0b\x32\x1a.flyteidl.admin.LaunchPlanR\x11triggerLaunchPlan\"\x17\n\x15\x43reateTriggerResponse\"T\n\x18\x44\x65\x61\x63tivateTriggerRequest\x12\x38\n\ntrigger_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\ttriggerId\"\x1b\n\x19\x44\x65\x61\x63tivateTriggerResponse\"\x80\x01\n\x10\x41rtifactProducer\x12\x36\n\tentity_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x08\x65ntityId\x12\x34\n\x07outputs\x18\x02 \x01(\x0b\x32\x1a.flyteidl.core.VariableMapR\x07outputs\"\\\n\x17RegisterProducerRequest\x12\x41\n\tproducers\x18\x01 \x03(\x0b\x32#.flyteidl.artifact.ArtifactProducerR\tproducers\"\x7f\n\x10\x41rtifactConsumer\x12\x36\n\tentity_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x08\x65ntityId\x12\x33\n\x06inputs\x18\x02 \x01(\x0b\x32\x1b.flyteidl.core.ParameterMapR\x06inputs\"\\\n\x17RegisterConsumerRequest\x12\x41\n\tconsumers\x18\x01 \x03(\x0b\x32#.flyteidl.artifact.ArtifactConsumerR\tconsumers\"\x12\n\x10RegisterResponse\"\x9a\x01\n\x16\x45xecutionInputsRequest\x12M\n\x0c\x65xecution_id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x0b\x65xecutionId\x12\x31\n\x06inputs\x18\x02 \x03(\x0b\x32\x19.flyteidl.core.ArtifactIDR\x06inputs\"\x19\n\x17\x45xecutionInputsResponse2\xf9\x0e\n\x10\x41rtifactRegistry\x12g\n\x0e\x43reateArtifact\x12(.flyteidl.artifact.CreateArtifactRequest\x1a).flyteidl.artifact.CreateArtifactResponse\"\x00\x12\xf1\x04\n\x0bGetArtifact\x12%.flyteidl.artifact.GetArtifactRequest\x1a&.flyteidl.artifact.GetArtifactResponse\"\x92\x04\x82\xd3\xe4\x93\x02\x8b\x04Z\xb3\x01\x12\xb0\x01/artifacts/api/v1/artifact/id/{query.artifact_id.artifact_key.project}/{query.artifact_id.artifact_key.domain}/{query.artifact_id.artifact_key.name}/{query.artifact_id.version}Z\x97\x01\x12\x94\x01/artifacts/api/v1/artifact/id/{query.artifact_id.artifact_key.project}/{query.artifact_id.artifact_key.domain}/{query.artifact_id.artifact_key.name}Z\x9b\x01\x12\x98\x01/artifacts/api/v1/artifact/tag/{query.artifact_tag.artifact_key.project}/{query.artifact_tag.artifact_key.domain}/{query.artifact_tag.artifact_key.name}\x12\x1b/artifacts/api/v1/artifacts\x12\x96\x02\n\x0fSearchArtifacts\x12).flyteidl.artifact.SearchArtifactsRequest\x1a*.flyteidl.artifact.SearchArtifactsResponse\"\xab\x01\x82\xd3\xe4\x93\x02\xa4\x01ZG\x12\x45/artifacts/api/v1/search/{artifact_key.project}/{artifact_key.domain}\x12Y/artifacts/api/v1/search/{artifact_key.project}/{artifact_key.domain}/{artifact_key.name}\x12\x64\n\rCreateTrigger\x12\'.flyteidl.artifact.CreateTriggerRequest\x1a(.flyteidl.artifact.CreateTriggerResponse\"\x00\x12\x9f\x01\n\x11\x44\x65\x61\x63tivateTrigger\x12+.flyteidl.artifact.DeactivateTriggerRequest\x1a,.flyteidl.artifact.DeactivateTriggerResponse\"/\x82\xd3\xe4\x93\x02):\x01*\"$/artifacts/api/v1/trigger/deactivate\x12O\n\x06\x41\x64\x64Tag\x12 .flyteidl.artifact.AddTagRequest\x1a!.flyteidl.artifact.AddTagResponse\"\x00\x12\x65\n\x10RegisterProducer\x12*.flyteidl.artifact.RegisterProducerRequest\x1a#.flyteidl.artifact.RegisterResponse\"\x00\x12\x65\n\x10RegisterConsumer\x12*.flyteidl.artifact.RegisterConsumerRequest\x1a#.flyteidl.artifact.RegisterResponse\"\x00\x12m\n\x12SetExecutionInputs\x12).flyteidl.artifact.ExecutionInputsRequest\x1a*.flyteidl.artifact.ExecutionInputsResponse\"\x00\x12\xd8\x01\n\x12\x46indByWorkflowExec\x12,.flyteidl.artifact.FindByWorkflowExecRequest\x1a*.flyteidl.artifact.SearchArtifactsResponse\"h\x82\xd3\xe4\x93\x02\x62\x12`/artifacts/api/v1/search/execution/{exec_id.project}/{exec_id.domain}/{exec_id.name}/{direction}B\xcc\x01\n\x15\x63om.flyteidl.artifactB\x0e\x41rtifactsProtoP\x01Z>github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact\xa2\x02\x03\x46\x41X\xaa\x02\x11\x46lyteidl.Artifact\xca\x02\x11\x46lyteidl\\Artifact\xe2\x02\x1d\x46lyteidl\\Artifact\\GPBMetadata\xea\x02\x12\x46lyteidl::Artifactb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.artifact.artifacts_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\025com.flyteidl.artifactB\016ArtifactsProtoP\001Z>github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact\242\002\003FAX\252\002\021Flyteidl.Artifact\312\002\021Flyteidl\\Artifact\342\002\035Flyteidl\\Artifact\\GPBMetadata\352\002\022Flyteidl::Artifact'
  _CREATEARTIFACTREQUEST_PARTITIONSENTRY._options = None
  _CREATEARTIFACTREQUEST_PARTITIONSENTRY._serialized_options = b'8\001'
  _ARTIFACTREGISTRY.methods_by_name['GetArtifact']._options = None
  _ARTIFACTREGISTRY.methods_by_name['GetArtifact']._serialized_options = b'\202\323\344\223\002\213\004Z\263\001\022\260\001/artifacts/api/v1/artifact/id/{query.artifact_id.artifact_key.project}/{query.artifact_id.artifact_key.domain}/{query.artifact_id.artifact_key.name}/{query.artifact_id.version}Z\227\001\022\224\001/artifacts/api/v1/artifact/id/{query.artifact_id.artifact_key.project}/{query.artifact_id.artifact_key.domain}/{query.artifact_id.artifact_key.name}Z\233\001\022\230\001/artifacts/api/v1/artifact/tag/{query.artifact_tag.artifact_key.project}/{query.artifact_tag.artifact_key.domain}/{query.artifact_tag.artifact_key.name}\022\033/artifacts/api/v1/artifacts'
  _ARTIFACTREGISTRY.methods_by_name['SearchArtifacts']._options = None
  _ARTIFACTREGISTRY.methods_by_name['SearchArtifacts']._serialized_options = b'\202\323\344\223\002\244\001ZG\022E/artifacts/api/v1/search/{artifact_key.project}/{artifact_key.domain}\022Y/artifacts/api/v1/search/{artifact_key.project}/{artifact_key.domain}/{artifact_key.name}'
  _ARTIFACTREGISTRY.methods_by_name['DeactivateTrigger']._options = None
  _ARTIFACTREGISTRY.methods_by_name['DeactivateTrigger']._serialized_options = b'\202\323\344\223\002):\001*\"$/artifacts/api/v1/trigger/deactivate'
  _ARTIFACTREGISTRY.methods_by_name['FindByWorkflowExec']._options = None
  _ARTIFACTREGISTRY.methods_by_name['FindByWorkflowExec']._serialized_options = b'\202\323\344\223\002b\022`/artifacts/api/v1/search/execution/{exec_id.project}/{exec_id.domain}/{exec_id.name}/{direction}'
  _globals['_ARTIFACT']._serialized_start=335
  _globals['_ARTIFACT']._serialized_end=537
  _globals['_CREATEARTIFACTREQUEST']._serialized_start=540
  _globals['_CREATEARTIFACTREQUEST']._serialized_end=935
  _globals['_CREATEARTIFACTREQUEST_PARTITIONSENTRY']._serialized_start=874
  _globals['_CREATEARTIFACTREQUEST_PARTITIONSENTRY']._serialized_end=935
  _globals['_ARTIFACTSOURCE']._serialized_start=938
  _globals['_ARTIFACTSOURCE']._serialized_end=1189
  _globals['_ARTIFACTSPEC']._serialized_start=1192
  _globals['_ARTIFACTSPEC']._serialized_end=1441
  _globals['_CREATEARTIFACTRESPONSE']._serialized_start=1443
  _globals['_CREATEARTIFACTRESPONSE']._serialized_end=1524
  _globals['_GETARTIFACTREQUEST']._serialized_start=1526
  _globals['_GETARTIFACTREQUEST']._serialized_end=1624
  _globals['_GETARTIFACTRESPONSE']._serialized_start=1626
  _globals['_GETARTIFACTRESPONSE']._serialized_end=1704
  _globals['_SEARCHOPTIONS']._serialized_start=1706
  _globals['_SEARCHOPTIONS']._serialized_end=1802
  _globals['_SEARCHARTIFACTSREQUEST']._serialized_start=1805
  _globals['_SEARCHARTIFACTSREQUEST']._serialized_end=2111
  _globals['_SEARCHARTIFACTSRESPONSE']._serialized_start=2113
  _globals['_SEARCHARTIFACTSRESPONSE']._serialized_end=2219
  _globals['_FINDBYWORKFLOWEXECREQUEST']._serialized_start=2222
  _globals['_FINDBYWORKFLOWEXECREQUEST']._serialized_end=2442
  _globals['_FINDBYWORKFLOWEXECREQUEST_DIRECTION']._serialized_start=2406
  _globals['_FINDBYWORKFLOWEXECREQUEST_DIRECTION']._serialized_end=2442
  _globals['_ADDTAGREQUEST']._serialized_start=2444
  _globals['_ADDTAGREQUEST']._serialized_end=2571
  _globals['_ADDTAGRESPONSE']._serialized_start=2573
  _globals['_ADDTAGRESPONSE']._serialized_end=2589
  _globals['_CREATETRIGGERREQUEST']._serialized_start=2591
  _globals['_CREATETRIGGERREQUEST']._serialized_end=2689
  _globals['_CREATETRIGGERRESPONSE']._serialized_start=2691
  _globals['_CREATETRIGGERRESPONSE']._serialized_end=2714
  _globals['_DEACTIVATETRIGGERREQUEST']._serialized_start=2716
  _globals['_DEACTIVATETRIGGERREQUEST']._serialized_end=2800
  _globals['_DEACTIVATETRIGGERRESPONSE']._serialized_start=2802
  _globals['_DEACTIVATETRIGGERRESPONSE']._serialized_end=2829
  _globals['_ARTIFACTPRODUCER']._serialized_start=2832
  _globals['_ARTIFACTPRODUCER']._serialized_end=2960
  _globals['_REGISTERPRODUCERREQUEST']._serialized_start=2962
  _globals['_REGISTERPRODUCERREQUEST']._serialized_end=3054
  _globals['_ARTIFACTCONSUMER']._serialized_start=3056
  _globals['_ARTIFACTCONSUMER']._serialized_end=3183
  _globals['_REGISTERCONSUMERREQUEST']._serialized_start=3185
  _globals['_REGISTERCONSUMERREQUEST']._serialized_end=3277
  _globals['_REGISTERRESPONSE']._serialized_start=3279
  _globals['_REGISTERRESPONSE']._serialized_end=3297
  _globals['_EXECUTIONINPUTSREQUEST']._serialized_start=3300
  _globals['_EXECUTIONINPUTSREQUEST']._serialized_end=3454
  _globals['_EXECUTIONINPUTSRESPONSE']._serialized_start=3456
  _globals['_EXECUTIONINPUTSRESPONSE']._serialized_end=3481
  _globals['_ARTIFACTREGISTRY']._serialized_start=3484
  _globals['_ARTIFACTREGISTRY']._serialized_end=5397
# @@protoc_insertion_point(module_scope)
