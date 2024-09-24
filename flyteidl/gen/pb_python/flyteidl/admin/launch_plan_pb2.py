# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/admin/launch_plan.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.core import execution_pb2 as flyteidl_dot_core_dot_execution__pb2
from flyteidl.core import execution_envs_pb2 as flyteidl_dot_core_dot_execution__envs__pb2
from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import interface_pb2 as flyteidl_dot_core_dot_interface__pb2
from flyteidl.core import security_pb2 as flyteidl_dot_core_dot_security__pb2
from flyteidl.core import workflow_pb2 as flyteidl_dot_core_dot_workflow__pb2
from flyteidl.admin import schedule_pb2 as flyteidl_dot_admin_dot_schedule__pb2
from flyteidl.admin import common_pb2 as flyteidl_dot_admin_dot_common__pb2
from google.protobuf import any_pb2 as google_dot_protobuf_dot_any__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from google.protobuf import wrappers_pb2 as google_dot_protobuf_dot_wrappers__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n flyteidl/admin/launch_plan.proto\x12\x0e\x66lyteidl.admin\x1a\x1d\x66lyteidl/core/execution.proto\x1a\"flyteidl/core/execution_envs.proto\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1d\x66lyteidl/core/interface.proto\x1a\x1c\x66lyteidl/core/security.proto\x1a\x1c\x66lyteidl/core/workflow.proto\x1a\x1d\x66lyteidl/admin/schedule.proto\x1a\x1b\x66lyteidl/admin/common.proto\x1a\x19google/protobuf/any.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1egoogle/protobuf/wrappers.proto\"x\n\x17LaunchPlanCreateRequest\x12)\n\x02id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x02id\x12\x32\n\x04spec\x18\x02 \x01(\x0b\x32\x1e.flyteidl.admin.LaunchPlanSpecR\x04spec\"\x1a\n\x18LaunchPlanCreateResponse\"\xa8\x01\n\nLaunchPlan\x12)\n\x02id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x02id\x12\x32\n\x04spec\x18\x02 \x01(\x0b\x32\x1e.flyteidl.admin.LaunchPlanSpecR\x04spec\x12;\n\x07\x63losure\x18\x03 \x01(\x0b\x32!.flyteidl.admin.LaunchPlanClosureR\x07\x63losure\"e\n\x0eLaunchPlanList\x12=\n\x0claunch_plans\x18\x01 \x03(\x0b\x32\x1a.flyteidl.admin.LaunchPlanR\x0blaunchPlans\x12\x14\n\x05token\x18\x02 \x01(\tR\x05token\"v\n\x04\x41uth\x12,\n\x12\x61ssumable_iam_role\x18\x01 \x01(\tR\x10\x61ssumableIamRole\x12<\n\x1akubernetes_service_account\x18\x02 \x01(\tR\x18kubernetesServiceAccount:\x02\x18\x01\"\xa0\x08\n\x0eLaunchPlanSpec\x12:\n\x0bworkflow_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\nworkflowId\x12K\n\x0f\x65ntity_metadata\x18\x02 \x01(\x0b\x32\".flyteidl.admin.LaunchPlanMetadataR\x0e\x65ntityMetadata\x12\x42\n\x0e\x64\x65\x66\x61ult_inputs\x18\x03 \x01(\x0b\x32\x1b.flyteidl.core.ParameterMapR\rdefaultInputs\x12<\n\x0c\x66ixed_inputs\x18\x04 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x0b\x66ixedInputs\x12\x16\n\x04role\x18\x05 \x01(\tB\x02\x18\x01R\x04role\x12.\n\x06labels\x18\x06 \x01(\x0b\x32\x16.flyteidl.admin.LabelsR\x06labels\x12=\n\x0b\x61nnotations\x18\x07 \x01(\x0b\x32\x1b.flyteidl.admin.AnnotationsR\x0b\x61nnotations\x12,\n\x04\x61uth\x18\x08 \x01(\x0b\x32\x14.flyteidl.admin.AuthB\x02\x18\x01R\x04\x61uth\x12\x39\n\tauth_role\x18\t \x01(\x0b\x32\x18.flyteidl.admin.AuthRoleB\x02\x18\x01R\x08\x61uthRole\x12I\n\x10security_context\x18\n \x01(\x0b\x32\x1e.flyteidl.core.SecurityContextR\x0fsecurityContext\x12M\n\x12quality_of_service\x18\x10 \x01(\x0b\x32\x1f.flyteidl.core.QualityOfServiceR\x10qualityOfService\x12X\n\x16raw_output_data_config\x18\x11 \x01(\x0b\x32#.flyteidl.admin.RawOutputDataConfigR\x13rawOutputDataConfig\x12\'\n\x0fmax_parallelism\x18\x12 \x01(\x05R\x0emaxParallelism\x12@\n\rinterruptible\x18\x13 \x01(\x0b\x32\x1a.google.protobuf.BoolValueR\rinterruptible\x12\'\n\x0foverwrite_cache\x18\x14 \x01(\x08R\x0eoverwriteCache\x12(\n\x04\x65nvs\x18\x15 \x01(\x0b\x32\x14.flyteidl.admin.EnvsR\x04\x65nvs\x12\x61\n\x19\x65xecution_env_assignments\x18\x16 \x03(\x0b\x32%.flyteidl.core.ExecutionEnvAssignmentR\x17\x65xecutionEnvAssignments\"\xcd\x02\n\x11LaunchPlanClosure\x12\x35\n\x05state\x18\x01 \x01(\x0e\x32\x1f.flyteidl.admin.LaunchPlanStateR\x05state\x12\x44\n\x0f\x65xpected_inputs\x18\x02 \x01(\x0b\x32\x1b.flyteidl.core.ParameterMapR\x0e\x65xpectedInputs\x12\x45\n\x10\x65xpected_outputs\x18\x03 \x01(\x0b\x32\x1a.flyteidl.core.VariableMapR\x0f\x65xpectedOutputs\x12\x39\n\ncreated_at\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x39\n\nupdated_at\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tupdatedAt\"\xd1\x01\n\x12LaunchPlanMetadata\x12\x34\n\x08schedule\x18\x01 \x01(\x0b\x32\x18.flyteidl.admin.ScheduleR\x08schedule\x12\x42\n\rnotifications\x18\x02 \x03(\x0b\x32\x1c.flyteidl.admin.NotificationR\rnotifications\x12\x41\n\x11launch_conditions\x18\x03 \x01(\x0b\x32\x14.google.protobuf.AnyR\x10launchConditions\"{\n\x17LaunchPlanUpdateRequest\x12)\n\x02id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x02id\x12\x35\n\x05state\x18\x02 \x01(\x0e\x32\x1f.flyteidl.admin.LaunchPlanStateR\x05state\"\x1a\n\x18LaunchPlanUpdateResponse\"P\n\x17\x41\x63tiveLaunchPlanRequest\x12\x35\n\x02id\x18\x01 \x01(\x0b\x32%.flyteidl.admin.NamedEntityIdentifierR\x02id\"\xbc\x01\n\x1b\x41\x63tiveLaunchPlanListRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x16\n\x06\x64omain\x18\x02 \x01(\tR\x06\x64omain\x12\x14\n\x05limit\x18\x03 \x01(\rR\x05limit\x12\x14\n\x05token\x18\x04 \x01(\tR\x05token\x12-\n\x07sort_by\x18\x05 \x01(\x0b\x32\x14.flyteidl.admin.SortR\x06sortBy\x12\x10\n\x03org\x18\x06 \x01(\tR\x03org\"1\n\x0fSubNodeIdAsList\x12\x1e\n\x0bsub_node_id\x18\x01 \x03(\tR\tsubNodeId\"P\n\x0bSubNodeList\x12\x41\n\x0csub_node_ids\x18\x01 \x03(\x0b\x32\x1f.flyteidl.admin.SubNodeIdAsListR\nsubNodeIds\"\xb6\x02\n\x1f\x43reateLaunchPlanFromNodeRequest\x12?\n\x0elaunch_plan_id\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\x0claunchPlanId\x12?\n\x0csub_node_ids\x18\x02 \x01(\x0b\x32\x1b.flyteidl.admin.SubNodeListH\x00R\nsubNodeIds\x12\x39\n\rsub_node_spec\x18\x03 \x01(\x0b\x32\x13.flyteidl.core.NodeH\x00R\x0bsubNodeSpec\x12I\n\x10security_context\x18\x04 \x01(\x0b\x32\x1e.flyteidl.core.SecurityContextR\x0fsecurityContextB\x0b\n\tsub_nodes\"_\n CreateLaunchPlanFromNodeResponse\x12;\n\x0blaunch_plan\x18\x01 \x01(\x0b\x32\x1a.flyteidl.admin.LaunchPlanR\nlaunchPlan*+\n\x0fLaunchPlanState\x12\x0c\n\x08INACTIVE\x10\x00\x12\n\n\x06\x41\x43TIVE\x10\x01\x42\xbb\x01\n\x12\x63om.flyteidl.adminB\x0fLaunchPlanProtoP\x01Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\xa2\x02\x03\x46\x41X\xaa\x02\x0e\x46lyteidl.Admin\xca\x02\x0e\x46lyteidl\\Admin\xe2\x02\x1a\x46lyteidl\\Admin\\GPBMetadata\xea\x02\x0f\x46lyteidl::Adminb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.admin.launch_plan_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\022com.flyteidl.adminB\017LaunchPlanProtoP\001Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\242\002\003FAX\252\002\016Flyteidl.Admin\312\002\016Flyteidl\\Admin\342\002\032Flyteidl\\Admin\\GPBMetadata\352\002\017Flyteidl::Admin'
  _AUTH._options = None
  _AUTH._serialized_options = b'\030\001'
  _LAUNCHPLANSPEC.fields_by_name['role']._options = None
  _LAUNCHPLANSPEC.fields_by_name['role']._serialized_options = b'\030\001'
  _LAUNCHPLANSPEC.fields_by_name['auth']._options = None
  _LAUNCHPLANSPEC.fields_by_name['auth']._serialized_options = b'\030\001'
  _LAUNCHPLANSPEC.fields_by_name['auth_role']._options = None
  _LAUNCHPLANSPEC.fields_by_name['auth_role']._serialized_options = b'\030\001'
  _globals['_LAUNCHPLANSTATE']._serialized_start=3544
  _globals['_LAUNCHPLANSTATE']._serialized_end=3587
  _globals['_LAUNCHPLANCREATEREQUEST']._serialized_start=424
  _globals['_LAUNCHPLANCREATEREQUEST']._serialized_end=544
  _globals['_LAUNCHPLANCREATERESPONSE']._serialized_start=546
  _globals['_LAUNCHPLANCREATERESPONSE']._serialized_end=572
  _globals['_LAUNCHPLAN']._serialized_start=575
  _globals['_LAUNCHPLAN']._serialized_end=743
  _globals['_LAUNCHPLANLIST']._serialized_start=745
  _globals['_LAUNCHPLANLIST']._serialized_end=846
  _globals['_AUTH']._serialized_start=848
  _globals['_AUTH']._serialized_end=966
  _globals['_LAUNCHPLANSPEC']._serialized_start=969
  _globals['_LAUNCHPLANSPEC']._serialized_end=2025
  _globals['_LAUNCHPLANCLOSURE']._serialized_start=2028
  _globals['_LAUNCHPLANCLOSURE']._serialized_end=2361
  _globals['_LAUNCHPLANMETADATA']._serialized_start=2364
  _globals['_LAUNCHPLANMETADATA']._serialized_end=2573
  _globals['_LAUNCHPLANUPDATEREQUEST']._serialized_start=2575
  _globals['_LAUNCHPLANUPDATEREQUEST']._serialized_end=2698
  _globals['_LAUNCHPLANUPDATERESPONSE']._serialized_start=2700
  _globals['_LAUNCHPLANUPDATERESPONSE']._serialized_end=2726
  _globals['_ACTIVELAUNCHPLANREQUEST']._serialized_start=2728
  _globals['_ACTIVELAUNCHPLANREQUEST']._serialized_end=2808
  _globals['_ACTIVELAUNCHPLANLISTREQUEST']._serialized_start=2811
  _globals['_ACTIVELAUNCHPLANLISTREQUEST']._serialized_end=2999
  _globals['_SUBNODEIDASLIST']._serialized_start=3001
  _globals['_SUBNODEIDASLIST']._serialized_end=3050
  _globals['_SUBNODELIST']._serialized_start=3052
  _globals['_SUBNODELIST']._serialized_end=3132
  _globals['_CREATELAUNCHPLANFROMNODEREQUEST']._serialized_start=3135
  _globals['_CREATELAUNCHPLANFROMNODEREQUEST']._serialized_end=3445
  _globals['_CREATELAUNCHPLANFROMNODERESPONSE']._serialized_start=3447
  _globals['_CREATELAUNCHPLANFROMNODERESPONSE']._serialized_end=3542
# @@protoc_insertion_point(module_scope)
