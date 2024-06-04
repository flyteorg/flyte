# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/admin/execution.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.admin import cluster_assignment_pb2 as flyteidl_dot_admin_dot_cluster__assignment__pb2
from flyteidl.admin import common_pb2 as flyteidl_dot_admin_dot_common__pb2
from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2
from flyteidl.core import execution_pb2 as flyteidl_dot_core_dot_execution__pb2
from flyteidl.core import execution_envs_pb2 as flyteidl_dot_core_dot_execution__envs__pb2
from flyteidl.core import artifact_id_pb2 as flyteidl_dot_core_dot_artifact__id__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import metrics_pb2 as flyteidl_dot_core_dot_metrics__pb2
from flyteidl.core import security_pb2 as flyteidl_dot_core_dot_security__pb2
from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from google.protobuf import wrappers_pb2 as google_dot_protobuf_dot_wrappers__pb2
from flyteidl.admin import matchable_resource_pb2 as flyteidl_dot_admin_dot_matchable__resource__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1e\x66lyteidl/admin/execution.proto\x12\x0e\x66lyteidl.admin\x1a\'flyteidl/admin/cluster_assignment.proto\x1a\x1b\x66lyteidl/admin/common.proto\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x1d\x66lyteidl/core/execution.proto\x1a\"flyteidl/core/execution_envs.proto\x1a\x1f\x66lyteidl/core/artifact_id.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1b\x66lyteidl/core/metrics.proto\x1a\x1c\x66lyteidl/core/security.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1egoogle/protobuf/wrappers.proto\x1a\'flyteidl/admin/matchable_resource.proto\"\xd6\x01\n\x16\x45xecutionCreateRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x16\n\x06\x64omain\x18\x02 \x01(\tR\x06\x64omain\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x31\n\x04spec\x18\x04 \x01(\x0b\x32\x1d.flyteidl.admin.ExecutionSpecR\x04spec\x12\x31\n\x06inputs\x18\x05 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x06inputs\x12\x10\n\x03org\x18\x06 \x01(\tR\x03org\"\x99\x01\n\x18\x45xecutionRelaunchRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\'\n\x0foverwrite_cache\x18\x04 \x01(\x08R\x0eoverwriteCacheJ\x04\x08\x02\x10\x03\"\xa8\x01\n\x17\x45xecutionRecoverRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12=\n\x08metadata\x18\x03 \x01(\x0b\x32!.flyteidl.admin.ExecutionMetadataR\x08metadata\"U\n\x17\x45xecutionCreateResponse\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\"Y\n\x1bWorkflowExecutionGetRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\"\xb6\x01\n\tExecution\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\x12\x31\n\x04spec\x18\x02 \x01(\x0b\x32\x1d.flyteidl.admin.ExecutionSpecR\x04spec\x12:\n\x07\x63losure\x18\x03 \x01(\x0b\x32 .flyteidl.admin.ExecutionClosureR\x07\x63losure\"`\n\rExecutionList\x12\x39\n\nexecutions\x18\x01 \x03(\x0b\x32\x19.flyteidl.admin.ExecutionR\nexecutions\x12\x14\n\x05token\x18\x02 \x01(\tR\x05token\"e\n\x0eLiteralMapBlob\x12\x37\n\x06values\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapB\x02\x18\x01H\x00R\x06values\x12\x12\n\x03uri\x18\x02 \x01(\tH\x00R\x03uriB\x06\n\x04\x64\x61ta\"C\n\rAbortMetadata\x12\x14\n\x05\x63\x61use\x18\x01 \x01(\tR\x05\x63\x61use\x12\x1c\n\tprincipal\x18\x02 \x01(\tR\tprincipal\"\xdc\x07\n\x10\x45xecutionClosure\x12>\n\x07outputs\x18\x01 \x01(\x0b\x32\x1e.flyteidl.admin.LiteralMapBlobB\x02\x18\x01H\x00R\x07outputs\x12\x35\n\x05\x65rror\x18\x02 \x01(\x0b\x32\x1d.flyteidl.core.ExecutionErrorH\x00R\x05\x65rror\x12%\n\x0b\x61\x62ort_cause\x18\n \x01(\tB\x02\x18\x01H\x00R\nabortCause\x12\x46\n\x0e\x61\x62ort_metadata\x18\x0c \x01(\x0b\x32\x1d.flyteidl.admin.AbortMetadataH\x00R\rabortMetadata\x12@\n\x0boutput_data\x18\r \x01(\x0b\x32\x19.flyteidl.core.LiteralMapB\x02\x18\x01H\x00R\noutputData\x12\x46\n\x0f\x63omputed_inputs\x18\x03 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapB\x02\x18\x01R\x0e\x63omputedInputs\x12<\n\x05phase\x18\x04 \x01(\x0e\x32&.flyteidl.core.WorkflowExecution.PhaseR\x05phase\x12\x39\n\nstarted_at\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tstartedAt\x12\x35\n\x08\x64uration\x18\x06 \x01(\x0b\x32\x19.google.protobuf.DurationR\x08\x64uration\x12\x39\n\ncreated_at\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\x39\n\nupdated_at\x18\x08 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tupdatedAt\x12\x42\n\rnotifications\x18\t \x03(\x0b\x32\x1c.flyteidl.admin.NotificationR\rnotifications\x12:\n\x0bworkflow_id\x18\x0b \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\nworkflowId\x12]\n\x14state_change_details\x18\x0e \x01(\x0b\x32+.flyteidl.admin.ExecutionStateChangeDetailsR\x12stateChangeDetails\x12\x42\n\rresolved_spec\x18\x0f \x01(\x0b\x32\x1d.flyteidl.admin.ExecutionSpecR\x0cresolvedSpecB\x0f\n\routput_result\"[\n\x0eSystemMetadata\x12+\n\x11\x65xecution_cluster\x18\x01 \x01(\tR\x10\x65xecutionCluster\x12\x1c\n\tnamespace\x18\x02 \x01(\tR\tnamespace\"\x85\x05\n\x11\x45xecutionMetadata\x12\x43\n\x04mode\x18\x01 \x01(\x0e\x32/.flyteidl.admin.ExecutionMetadata.ExecutionModeR\x04mode\x12\x1c\n\tprincipal\x18\x02 \x01(\tR\tprincipal\x12\x18\n\x07nesting\x18\x03 \x01(\rR\x07nesting\x12=\n\x0cscheduled_at\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\x0bscheduledAt\x12Z\n\x15parent_node_execution\x18\x05 \x01(\x0b\x32&.flyteidl.core.NodeExecutionIdentifierR\x13parentNodeExecution\x12[\n\x13reference_execution\x18\x10 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x12referenceExecution\x12G\n\x0fsystem_metadata\x18\x11 \x01(\x0b\x32\x1e.flyteidl.admin.SystemMetadataR\x0esystemMetadata\x12<\n\x0c\x61rtifact_ids\x18\x12 \x03(\x0b\x32\x19.flyteidl.core.ArtifactIDR\x0b\x61rtifactIds\"t\n\rExecutionMode\x12\n\n\x06MANUAL\x10\x00\x12\r\n\tSCHEDULED\x10\x01\x12\n\n\x06SYSTEM\x10\x02\x12\x0c\n\x08RELAUNCH\x10\x03\x12\x12\n\x0e\x43HILD_WORKFLOW\x10\x04\x12\r\n\tRECOVERED\x10\x05\x12\x0b\n\x07TRIGGER\x10\x06\"V\n\x10NotificationList\x12\x42\n\rnotifications\x18\x01 \x03(\x0b\x32\x1c.flyteidl.admin.NotificationR\rnotifications\"\xb4\n\n\rExecutionSpec\x12:\n\x0blaunch_plan\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.IdentifierR\nlaunchPlan\x12\x35\n\x06inputs\x18\x02 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapB\x02\x18\x01R\x06inputs\x12=\n\x08metadata\x18\x03 \x01(\x0b\x32!.flyteidl.admin.ExecutionMetadataR\x08metadata\x12H\n\rnotifications\x18\x05 \x01(\x0b\x32 .flyteidl.admin.NotificationListH\x00R\rnotifications\x12!\n\x0b\x64isable_all\x18\x06 \x01(\x08H\x00R\ndisableAll\x12.\n\x06labels\x18\x07 \x01(\x0b\x32\x16.flyteidl.admin.LabelsR\x06labels\x12=\n\x0b\x61nnotations\x18\x08 \x01(\x0b\x32\x1b.flyteidl.admin.AnnotationsR\x0b\x61nnotations\x12I\n\x10security_context\x18\n \x01(\x0b\x32\x1e.flyteidl.core.SecurityContextR\x0fsecurityContext\x12\x39\n\tauth_role\x18\x10 \x01(\x0b\x32\x18.flyteidl.admin.AuthRoleB\x02\x18\x01R\x08\x61uthRole\x12M\n\x12quality_of_service\x18\x11 \x01(\x0b\x32\x1f.flyteidl.core.QualityOfServiceR\x10qualityOfService\x12\'\n\x0fmax_parallelism\x18\x12 \x01(\x05R\x0emaxParallelism\x12X\n\x16raw_output_data_config\x18\x13 \x01(\x0b\x32#.flyteidl.admin.RawOutputDataConfigR\x13rawOutputDataConfig\x12P\n\x12\x63luster_assignment\x18\x14 \x01(\x0b\x32!.flyteidl.admin.ClusterAssignmentR\x11\x63lusterAssignment\x12@\n\rinterruptible\x18\x15 \x01(\x0b\x32\x1a.google.protobuf.BoolValueR\rinterruptible\x12\'\n\x0foverwrite_cache\x18\x16 \x01(\x08R\x0eoverwriteCache\x12(\n\x04\x65nvs\x18\x17 \x01(\x0b\x32\x14.flyteidl.admin.EnvsR\x04\x65nvs\x12\x12\n\x04tags\x18\x18 \x03(\tR\x04tags\x12]\n\x17\x65xecution_cluster_label\x18\x19 \x01(\x0b\x32%.flyteidl.admin.ExecutionClusterLabelR\x15\x65xecutionClusterLabel\x12\x61\n\x19\x65xecution_env_assignments\x18\x1a \x03(\x0b\x32%.flyteidl.core.ExecutionEnvAssignmentR\x17\x65xecutionEnvAssignments\x12`\n\x18task_resource_attributes\x18\x1b \x01(\x0b\x32&.flyteidl.admin.TaskResourceAttributesR\x16taskResourceAttributesB\x18\n\x16notification_overridesJ\x04\x08\x04\x10\x05\"m\n\x19\x45xecutionTerminateRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\x12\x14\n\x05\x63\x61use\x18\x02 \x01(\tR\x05\x63\x61use\"\x1c\n\x1a\x45xecutionTerminateResponse\"]\n\x1fWorkflowExecutionGetDataRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\"\x88\x02\n WorkflowExecutionGetDataResponse\x12\x35\n\x07outputs\x18\x01 \x01(\x0b\x32\x17.flyteidl.admin.UrlBlobB\x02\x18\x01R\x07outputs\x12\x33\n\x06inputs\x18\x02 \x01(\x0b\x32\x17.flyteidl.admin.UrlBlobB\x02\x18\x01R\x06inputs\x12:\n\x0b\x66ull_inputs\x18\x03 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\nfullInputs\x12<\n\x0c\x66ull_outputs\x18\x04 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapR\x0b\x66ullOutputs\"\x8a\x01\n\x16\x45xecutionUpdateRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\x12\x34\n\x05state\x18\x02 \x01(\x0e\x32\x1e.flyteidl.admin.ExecutionStateR\x05state\"\xae\x01\n\x1b\x45xecutionStateChangeDetails\x12\x34\n\x05state\x18\x01 \x01(\x0e\x32\x1e.flyteidl.admin.ExecutionStateR\x05state\x12;\n\x0boccurred_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\noccurredAt\x12\x1c\n\tprincipal\x18\x03 \x01(\tR\tprincipal\"\x19\n\x17\x45xecutionUpdateResponse\"v\n\"WorkflowExecutionGetMetricsRequest\x12:\n\x02id\x18\x01 \x01(\x0b\x32*.flyteidl.core.WorkflowExecutionIdentifierR\x02id\x12\x14\n\x05\x64\x65pth\x18\x02 \x01(\x05R\x05\x64\x65pth\"N\n#WorkflowExecutionGetMetricsResponse\x12\'\n\x04span\x18\x01 \x01(\x0b\x32\x13.flyteidl.core.SpanR\x04span\"y\n\x19\x45xecutionCountsGetRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x16\n\x06\x64omain\x18\x02 \x01(\tR\x06\x64omain\x12\x10\n\x03org\x18\x03 \x01(\tR\x03org\x12\x18\n\x07\x66ilters\x18\x04 \x01(\tR\x07\x66ilters\"l\n\x16\x45xecutionCountsByPhase\x12<\n\x05phase\x18\x01 \x01(\x0e\x32&.flyteidl.core.WorkflowExecution.PhaseR\x05phase\x12\x14\n\x05\x63ount\x18\x02 \x01(\x03R\x05\x63ount\"o\n\x1a\x45xecutionCountsGetResponse\x12Q\n\x10\x65xecution_counts\x18\x01 \x03(\x0b\x32&.flyteidl.admin.ExecutionCountsByPhaseR\x0f\x65xecutionCounts\"f\n RunningExecutionsCountGetRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x16\n\x06\x64omain\x18\x02 \x01(\tR\x06\x64omain\x12\x10\n\x03org\x18\x03 \x01(\tR\x03org\"9\n!RunningExecutionsCountGetResponse\x12\x14\n\x05\x63ount\x18\x01 \x01(\x03R\x05\x63ount*>\n\x0e\x45xecutionState\x12\x14\n\x10\x45XECUTION_ACTIVE\x10\x00\x12\x16\n\x12\x45XECUTION_ARCHIVED\x10\x01\x42\xba\x01\n\x12\x63om.flyteidl.adminB\x0e\x45xecutionProtoP\x01Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\xa2\x02\x03\x46\x41X\xaa\x02\x0e\x46lyteidl.Admin\xca\x02\x0e\x46lyteidl\\Admin\xe2\x02\x1a\x46lyteidl\\Admin\\GPBMetadata\xea\x02\x0f\x46lyteidl::Adminb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.admin.execution_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\022com.flyteidl.adminB\016ExecutionProtoP\001Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\242\002\003FAX\252\002\016Flyteidl.Admin\312\002\016Flyteidl\\Admin\342\002\032Flyteidl\\Admin\\GPBMetadata\352\002\017Flyteidl::Admin'
  _LITERALMAPBLOB.fields_by_name['values']._options = None
  _LITERALMAPBLOB.fields_by_name['values']._serialized_options = b'\030\001'
  _EXECUTIONCLOSURE.fields_by_name['outputs']._options = None
  _EXECUTIONCLOSURE.fields_by_name['outputs']._serialized_options = b'\030\001'
  _EXECUTIONCLOSURE.fields_by_name['abort_cause']._options = None
  _EXECUTIONCLOSURE.fields_by_name['abort_cause']._serialized_options = b'\030\001'
  _EXECUTIONCLOSURE.fields_by_name['output_data']._options = None
  _EXECUTIONCLOSURE.fields_by_name['output_data']._serialized_options = b'\030\001'
  _EXECUTIONCLOSURE.fields_by_name['computed_inputs']._options = None
  _EXECUTIONCLOSURE.fields_by_name['computed_inputs']._serialized_options = b'\030\001'
  _EXECUTIONSPEC.fields_by_name['inputs']._options = None
  _EXECUTIONSPEC.fields_by_name['inputs']._serialized_options = b'\030\001'
  _EXECUTIONSPEC.fields_by_name['auth_role']._options = None
  _EXECUTIONSPEC.fields_by_name['auth_role']._serialized_options = b'\030\001'
  _WORKFLOWEXECUTIONGETDATARESPONSE.fields_by_name['outputs']._options = None
  _WORKFLOWEXECUTIONGETDATARESPONSE.fields_by_name['outputs']._serialized_options = b'\030\001'
  _WORKFLOWEXECUTIONGETDATARESPONSE.fields_by_name['inputs']._options = None
  _WORKFLOWEXECUTIONGETDATARESPONSE.fields_by_name['inputs']._serialized_options = b'\030\001'
  _globals['_EXECUTIONSTATE']._serialized_start=6368
  _globals['_EXECUTIONSTATE']._serialized_end=6430
  _globals['_EXECUTIONCREATEREQUEST']._serialized_start=480
  _globals['_EXECUTIONCREATEREQUEST']._serialized_end=694
  _globals['_EXECUTIONRELAUNCHREQUEST']._serialized_start=697
  _globals['_EXECUTIONRELAUNCHREQUEST']._serialized_end=850
  _globals['_EXECUTIONRECOVERREQUEST']._serialized_start=853
  _globals['_EXECUTIONRECOVERREQUEST']._serialized_end=1021
  _globals['_EXECUTIONCREATERESPONSE']._serialized_start=1023
  _globals['_EXECUTIONCREATERESPONSE']._serialized_end=1108
  _globals['_WORKFLOWEXECUTIONGETREQUEST']._serialized_start=1110
  _globals['_WORKFLOWEXECUTIONGETREQUEST']._serialized_end=1199
  _globals['_EXECUTION']._serialized_start=1202
  _globals['_EXECUTION']._serialized_end=1384
  _globals['_EXECUTIONLIST']._serialized_start=1386
  _globals['_EXECUTIONLIST']._serialized_end=1482
  _globals['_LITERALMAPBLOB']._serialized_start=1484
  _globals['_LITERALMAPBLOB']._serialized_end=1585
  _globals['_ABORTMETADATA']._serialized_start=1587
  _globals['_ABORTMETADATA']._serialized_end=1654
  _globals['_EXECUTIONCLOSURE']._serialized_start=1657
  _globals['_EXECUTIONCLOSURE']._serialized_end=2645
  _globals['_SYSTEMMETADATA']._serialized_start=2647
  _globals['_SYSTEMMETADATA']._serialized_end=2738
  _globals['_EXECUTIONMETADATA']._serialized_start=2741
  _globals['_EXECUTIONMETADATA']._serialized_end=3386
  _globals['_EXECUTIONMETADATA_EXECUTIONMODE']._serialized_start=3270
  _globals['_EXECUTIONMETADATA_EXECUTIONMODE']._serialized_end=3386
  _globals['_NOTIFICATIONLIST']._serialized_start=3388
  _globals['_NOTIFICATIONLIST']._serialized_end=3474
  _globals['_EXECUTIONSPEC']._serialized_start=3477
  _globals['_EXECUTIONSPEC']._serialized_end=4809
  _globals['_EXECUTIONTERMINATEREQUEST']._serialized_start=4811
  _globals['_EXECUTIONTERMINATEREQUEST']._serialized_end=4920
  _globals['_EXECUTIONTERMINATERESPONSE']._serialized_start=4922
  _globals['_EXECUTIONTERMINATERESPONSE']._serialized_end=4950
  _globals['_WORKFLOWEXECUTIONGETDATAREQUEST']._serialized_start=4952
  _globals['_WORKFLOWEXECUTIONGETDATAREQUEST']._serialized_end=5045
  _globals['_WORKFLOWEXECUTIONGETDATARESPONSE']._serialized_start=5048
  _globals['_WORKFLOWEXECUTIONGETDATARESPONSE']._serialized_end=5312
  _globals['_EXECUTIONUPDATEREQUEST']._serialized_start=5315
  _globals['_EXECUTIONUPDATEREQUEST']._serialized_end=5453
  _globals['_EXECUTIONSTATECHANGEDETAILS']._serialized_start=5456
  _globals['_EXECUTIONSTATECHANGEDETAILS']._serialized_end=5630
  _globals['_EXECUTIONUPDATERESPONSE']._serialized_start=5632
  _globals['_EXECUTIONUPDATERESPONSE']._serialized_end=5657
  _globals['_WORKFLOWEXECUTIONGETMETRICSREQUEST']._serialized_start=5659
  _globals['_WORKFLOWEXECUTIONGETMETRICSREQUEST']._serialized_end=5777
  _globals['_WORKFLOWEXECUTIONGETMETRICSRESPONSE']._serialized_start=5779
  _globals['_WORKFLOWEXECUTIONGETMETRICSRESPONSE']._serialized_end=5857
  _globals['_EXECUTIONCOUNTSGETREQUEST']._serialized_start=5859
  _globals['_EXECUTIONCOUNTSGETREQUEST']._serialized_end=5980
  _globals['_EXECUTIONCOUNTSBYPHASE']._serialized_start=5982
  _globals['_EXECUTIONCOUNTSBYPHASE']._serialized_end=6090
  _globals['_EXECUTIONCOUNTSGETRESPONSE']._serialized_start=6092
  _globals['_EXECUTIONCOUNTSGETRESPONSE']._serialized_end=6203
  _globals['_RUNNINGEXECUTIONSCOUNTGETREQUEST']._serialized_start=6205
  _globals['_RUNNINGEXECUTIONSCOUNTGETREQUEST']._serialized_end=6307
  _globals['_RUNNINGEXECUTIONSCOUNTGETRESPONSE']._serialized_start=6309
  _globals['_RUNNINGEXECUTIONSCOUNTGETRESPONSE']._serialized_end=6366
# @@protoc_insertion_point(module_scope)
