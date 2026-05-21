from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.common import phase_pb2 as _phase_pb2
from flyteidl2.common import run_pb2 as _run_pb2
from flyteidl2.core import execution_pb2 as _execution_pb2
from flyteidl2.task import common_pb2 as _common_pb2
from flyteidl2.task import run_pb2 as _run_pb2_1
from flyteidl2.task import task_definition_pb2 as _task_definition_pb2
from flyteidl2.workflow import run_definition_pb2 as _run_definition_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateRunRequest(_message.Message):
    __slots__ = ["run_id", "project_id", "task_id", "task_spec", "trigger_name", "inputs", "offloaded_input_data", "run_spec", "source"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_SPEC_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_NAME_FIELD_NUMBER: _ClassVar[int]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OFFLOADED_INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    RUN_SPEC_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    task_spec: _task_definition_pb2.TaskSpec
    trigger_name: _identifier_pb2.TriggerName
    inputs: _common_pb2.Inputs
    offloaded_input_data: _run_pb2.OffloadedInputData
    run_spec: _run_pb2_1.RunSpec
    source: _run_definition_pb2.RunSource
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ..., task_spec: _Optional[_Union[_task_definition_pb2.TaskSpec, _Mapping]] = ..., trigger_name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ..., inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., offloaded_input_data: _Optional[_Union[_run_pb2.OffloadedInputData, _Mapping]] = ..., run_spec: _Optional[_Union[_run_pb2_1.RunSpec, _Mapping]] = ..., source: _Optional[_Union[_run_definition_pb2.RunSource, str]] = ...) -> None: ...

class CreateRunResponse(_message.Message):
    __slots__ = ["run"]
    RUN_FIELD_NUMBER: _ClassVar[int]
    run: _run_definition_pb2.Run
    def __init__(self, run: _Optional[_Union[_run_definition_pb2.Run, _Mapping]] = ...) -> None: ...

class AbortRunRequest(_message.Message):
    __slots__ = ["run_id", "reason"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    reason: str
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortRunResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRunDetailsRequest(_message.Message):
    __slots__ = ["run_id"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...

class GetRunDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.RunDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.RunDetails, _Mapping]] = ...) -> None: ...

class WatchRunDetailsRequest(_message.Message):
    __slots__ = ["run_id"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...

class WatchRunDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.RunDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.RunDetails, _Mapping]] = ...) -> None: ...

class GetActionDetailsRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetActionDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.ActionDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.ActionDetails, _Mapping]] = ...) -> None: ...

class WatchActionDetailsRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class WatchActionDetailsResponse(_message.Message):
    __slots__ = ["details"]
    DETAILS_FIELD_NUMBER: _ClassVar[int]
    details: _run_definition_pb2.ActionDetails
    def __init__(self, details: _Optional[_Union[_run_definition_pb2.ActionDetails, _Mapping]] = ...) -> None: ...

class GetActionDataRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetActionDataResponse(_message.Message):
    __slots__ = ["inputs", "outputs"]
    INPUTS_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_FIELD_NUMBER: _ClassVar[int]
    inputs: _common_pb2.Inputs
    outputs: _common_pb2.Outputs
    def __init__(self, inputs: _Optional[_Union[_common_pb2.Inputs, _Mapping]] = ..., outputs: _Optional[_Union[_common_pb2.Outputs, _Mapping]] = ...) -> None: ...

class GetActionDataURIsRequest(_message.Message):
    __slots__ = ["action_id"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ...) -> None: ...

class GetActionDataURIsResponse(_message.Message):
    __slots__ = ["inputs_uri", "outputs_uri"]
    INPUTS_URI_FIELD_NUMBER: _ClassVar[int]
    OUTPUTS_URI_FIELD_NUMBER: _ClassVar[int]
    inputs_uri: str
    outputs_uri: str
    def __init__(self, inputs_uri: _Optional[str] = ..., outputs_uri: _Optional[str] = ...) -> None: ...

class GetActionLogContextRequest(_message.Message):
    __slots__ = ["action_id", "attempt"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    attempt: int
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ...) -> None: ...

class GetActionLogContextResponse(_message.Message):
    __slots__ = ["log_context", "cluster"]
    LOG_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_FIELD_NUMBER: _ClassVar[int]
    log_context: _execution_pb2.LogContext
    cluster: str
    def __init__(self, log_context: _Optional[_Union[_execution_pb2.LogContext, _Mapping]] = ..., cluster: _Optional[str] = ...) -> None: ...

class ListRunsRequest(_message.Message):
    __slots__ = ["request", "org", "project_id", "trigger_name", "task_name", "task_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TRIGGER_NAME_FIELD_NUMBER: _ClassVar[int]
    TASK_NAME_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    project_id: _identifier_pb2.ProjectIdentifier
    trigger_name: _identifier_pb2.TriggerName
    task_name: _task_definition_pb2.TaskName
    task_id: _task_definition_pb2.TaskIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., trigger_name: _Optional[_Union[_identifier_pb2.TriggerName, _Mapping]] = ..., task_name: _Optional[_Union[_task_definition_pb2.TaskName, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ...) -> None: ...

class ListRunsResponse(_message.Message):
    __slots__ = ["runs", "token"]
    RUNS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    runs: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.Run]
    token: str
    def __init__(self, runs: _Optional[_Iterable[_Union[_run_definition_pb2.Run, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class WatchRunsRequest(_message.Message):
    __slots__ = ["org", "cluster_id", "project_id", "task_id"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    TASK_ID_FIELD_NUMBER: _ClassVar[int]
    org: str
    cluster_id: _identifier_pb2.ClusterIdentifier
    project_id: _identifier_pb2.ProjectIdentifier
    task_id: _task_definition_pb2.TaskIdentifier
    def __init__(self, org: _Optional[str] = ..., cluster_id: _Optional[_Union[_identifier_pb2.ClusterIdentifier, _Mapping]] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., task_id: _Optional[_Union[_task_definition_pb2.TaskIdentifier, _Mapping]] = ...) -> None: ...

class WatchRunsResponse(_message.Message):
    __slots__ = ["runs"]
    RUNS_FIELD_NUMBER: _ClassVar[int]
    runs: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.Run]
    def __init__(self, runs: _Optional[_Iterable[_Union[_run_definition_pb2.Run, _Mapping]]] = ...) -> None: ...

class ListActionsRequest(_message.Message):
    __slots__ = ["request", "run_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    run_id: _identifier_pb2.RunIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ...) -> None: ...

class ListActionsResponse(_message.Message):
    __slots__ = ["actions", "token"]
    ACTIONS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    actions: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.Action]
    token: str
    def __init__(self, actions: _Optional[_Iterable[_Union[_run_definition_pb2.Action, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class WatchActionsRequest(_message.Message):
    __slots__ = ["run_id", "filter", "enable_run_store"]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    FILTER_FIELD_NUMBER: _ClassVar[int]
    ENABLE_RUN_STORE_FIELD_NUMBER: _ClassVar[int]
    run_id: _identifier_pb2.RunIdentifier
    filter: _containers.RepeatedCompositeFieldContainer[_list_pb2.Filter]
    enable_run_store: bool
    def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., filter: _Optional[_Iterable[_Union[_list_pb2.Filter, _Mapping]]] = ..., enable_run_store: bool = ...) -> None: ...

class WatchActionsResponse(_message.Message):
    __slots__ = ["enriched_actions"]
    ENRICHED_ACTIONS_FIELD_NUMBER: _ClassVar[int]
    enriched_actions: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.EnrichedAction]
    def __init__(self, enriched_actions: _Optional[_Iterable[_Union[_run_definition_pb2.EnrichedAction, _Mapping]]] = ...) -> None: ...

class WatchWindowedActionsRequest(_message.Message):
    __slots__ = ["subscribe", "update_window"]
    class Subscribe(_message.Message):
        __slots__ = ["run_id", "selected_item_id", "overscan_before", "overscan_after", "expanded_nodes", "phase_filter", "name_filter"]
        class ExpandedNodesEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: NodeExpansionParams
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[NodeExpansionParams, _Mapping]] = ...) -> None: ...
        RUN_ID_FIELD_NUMBER: _ClassVar[int]
        SELECTED_ITEM_ID_FIELD_NUMBER: _ClassVar[int]
        OVERSCAN_BEFORE_FIELD_NUMBER: _ClassVar[int]
        OVERSCAN_AFTER_FIELD_NUMBER: _ClassVar[int]
        EXPANDED_NODES_FIELD_NUMBER: _ClassVar[int]
        PHASE_FILTER_FIELD_NUMBER: _ClassVar[int]
        NAME_FILTER_FIELD_NUMBER: _ClassVar[int]
        run_id: _identifier_pb2.RunIdentifier
        selected_item_id: str
        overscan_before: int
        overscan_after: int
        expanded_nodes: _containers.MessageMap[str, NodeExpansionParams]
        phase_filter: _containers.RepeatedScalarFieldContainer[_phase_pb2.ActionPhase]
        name_filter: str
        def __init__(self, run_id: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., selected_item_id: _Optional[str] = ..., overscan_before: _Optional[int] = ..., overscan_after: _Optional[int] = ..., expanded_nodes: _Optional[_Mapping[str, NodeExpansionParams]] = ..., phase_filter: _Optional[_Iterable[_Union[_phase_pb2.ActionPhase, str]]] = ..., name_filter: _Optional[str] = ...) -> None: ...
    class UpdateWindow(_message.Message):
        __slots__ = ["selected_item_id", "overscan_before", "overscan_after", "expanded_nodes", "phase_filter", "name_filter"]
        class ExpandedNodesEntry(_message.Message):
            __slots__ = ["key", "value"]
            KEY_FIELD_NUMBER: _ClassVar[int]
            VALUE_FIELD_NUMBER: _ClassVar[int]
            key: str
            value: NodeExpansionParams
            def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[NodeExpansionParams, _Mapping]] = ...) -> None: ...
        SELECTED_ITEM_ID_FIELD_NUMBER: _ClassVar[int]
        OVERSCAN_BEFORE_FIELD_NUMBER: _ClassVar[int]
        OVERSCAN_AFTER_FIELD_NUMBER: _ClassVar[int]
        EXPANDED_NODES_FIELD_NUMBER: _ClassVar[int]
        PHASE_FILTER_FIELD_NUMBER: _ClassVar[int]
        NAME_FILTER_FIELD_NUMBER: _ClassVar[int]
        selected_item_id: str
        overscan_before: int
        overscan_after: int
        expanded_nodes: _containers.MessageMap[str, NodeExpansionParams]
        phase_filter: _containers.RepeatedScalarFieldContainer[_phase_pb2.ActionPhase]
        name_filter: str
        def __init__(self, selected_item_id: _Optional[str] = ..., overscan_before: _Optional[int] = ..., overscan_after: _Optional[int] = ..., expanded_nodes: _Optional[_Mapping[str, NodeExpansionParams]] = ..., phase_filter: _Optional[_Iterable[_Union[_phase_pb2.ActionPhase, str]]] = ..., name_filter: _Optional[str] = ...) -> None: ...
    SUBSCRIBE_FIELD_NUMBER: _ClassVar[int]
    UPDATE_WINDOW_FIELD_NUMBER: _ClassVar[int]
    subscribe: WatchWindowedActionsRequest.Subscribe
    update_window: WatchWindowedActionsRequest.UpdateWindow
    def __init__(self, subscribe: _Optional[_Union[WatchWindowedActionsRequest.Subscribe, _Mapping]] = ..., update_window: _Optional[_Union[WatchWindowedActionsRequest.UpdateWindow, _Mapping]] = ...) -> None: ...

class NodeExpansionParams(_message.Message):
    __slots__ = ["offset", "limit"]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    offset: int
    limit: int
    def __init__(self, offset: _Optional[int] = ..., limit: _Optional[int] = ...) -> None: ...

class WatchWindowedActionsResponse(_message.Message):
    __slots__ = ["window_items", "ancestors", "total_flat_count", "selected_flat_index", "initial_snapshot_complete", "truncations", "resync_hint"]
    WINDOW_ITEMS_FIELD_NUMBER: _ClassVar[int]
    ANCESTORS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FLAT_COUNT_FIELD_NUMBER: _ClassVar[int]
    SELECTED_FLAT_INDEX_FIELD_NUMBER: _ClassVar[int]
    INITIAL_SNAPSHOT_COMPLETE_FIELD_NUMBER: _ClassVar[int]
    TRUNCATIONS_FIELD_NUMBER: _ClassVar[int]
    RESYNC_HINT_FIELD_NUMBER: _ClassVar[int]
    window_items: _containers.RepeatedCompositeFieldContainer[WindowedItem]
    ancestors: _containers.RepeatedCompositeFieldContainer[WindowedItem]
    total_flat_count: int
    selected_flat_index: int
    initial_snapshot_complete: bool
    truncations: _containers.RepeatedCompositeFieldContainer[TruncationNotice]
    resync_hint: bool
    def __init__(self, window_items: _Optional[_Iterable[_Union[WindowedItem, _Mapping]]] = ..., ancestors: _Optional[_Iterable[_Union[WindowedItem, _Mapping]]] = ..., total_flat_count: _Optional[int] = ..., selected_flat_index: _Optional[int] = ..., initial_snapshot_complete: bool = ..., truncations: _Optional[_Iterable[_Union[TruncationNotice, _Mapping]]] = ..., resync_hint: bool = ...) -> None: ...

class WindowedItem(_message.Message):
    __slots__ = ["action", "group", "depth", "is_expanded"]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    GROUP_FIELD_NUMBER: _ClassVar[int]
    DEPTH_FIELD_NUMBER: _ClassVar[int]
    IS_EXPANDED_FIELD_NUMBER: _ClassVar[int]
    action: _run_definition_pb2.EnrichedAction
    group: GroupNode
    depth: int
    is_expanded: bool
    def __init__(self, action: _Optional[_Union[_run_definition_pb2.EnrichedAction, _Mapping]] = ..., group: _Optional[_Union[GroupNode, _Mapping]] = ..., depth: _Optional[int] = ..., is_expanded: bool = ...) -> None: ...

class ActionLeaf(_message.Message):
    __slots__ = ["action_id", "short_name", "duration"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    SHORT_NAME_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    action_id: str
    short_name: str
    duration: _duration_pb2.Duration
    def __init__(self, action_id: _Optional[str] = ..., short_name: _Optional[str] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class GroupAggregations(_message.Message):
    __slots__ = ["failed", "longest_duration", "longest_running", "longest_setup"]
    FAILED_FIELD_NUMBER: _ClassVar[int]
    LONGEST_DURATION_FIELD_NUMBER: _ClassVar[int]
    LONGEST_RUNNING_FIELD_NUMBER: _ClassVar[int]
    LONGEST_SETUP_FIELD_NUMBER: _ClassVar[int]
    failed: _containers.RepeatedCompositeFieldContainer[ActionLeaf]
    longest_duration: _containers.RepeatedCompositeFieldContainer[ActionLeaf]
    longest_running: _containers.RepeatedCompositeFieldContainer[ActionLeaf]
    longest_setup: _containers.RepeatedCompositeFieldContainer[ActionLeaf]
    def __init__(self, failed: _Optional[_Iterable[_Union[ActionLeaf, _Mapping]]] = ..., longest_duration: _Optional[_Iterable[_Union[ActionLeaf, _Mapping]]] = ..., longest_running: _Optional[_Iterable[_Union[ActionLeaf, _Mapping]]] = ..., longest_setup: _Optional[_Iterable[_Union[ActionLeaf, _Mapping]]] = ...) -> None: ...

class GroupNode(_message.Message):
    __slots__ = ["id", "group_name", "parent_id", "child_phase_counts", "earliest_start_time", "latest_end_time", "total_children", "meets_filter", "aggregations"]
    class ChildPhaseCountsEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: int
        value: int
        def __init__(self, key: _Optional[int] = ..., value: _Optional[int] = ...) -> None: ...
    ID_FIELD_NUMBER: _ClassVar[int]
    GROUP_NAME_FIELD_NUMBER: _ClassVar[int]
    PARENT_ID_FIELD_NUMBER: _ClassVar[int]
    CHILD_PHASE_COUNTS_FIELD_NUMBER: _ClassVar[int]
    EARLIEST_START_TIME_FIELD_NUMBER: _ClassVar[int]
    LATEST_END_TIME_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CHILDREN_FIELD_NUMBER: _ClassVar[int]
    MEETS_FILTER_FIELD_NUMBER: _ClassVar[int]
    AGGREGATIONS_FIELD_NUMBER: _ClassVar[int]
    id: str
    group_name: str
    parent_id: str
    child_phase_counts: _containers.ScalarMap[int, int]
    earliest_start_time: _timestamp_pb2.Timestamp
    latest_end_time: _timestamp_pb2.Timestamp
    total_children: int
    meets_filter: bool
    aggregations: GroupAggregations
    def __init__(self, id: _Optional[str] = ..., group_name: _Optional[str] = ..., parent_id: _Optional[str] = ..., child_phase_counts: _Optional[_Mapping[int, int]] = ..., earliest_start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., latest_end_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., total_children: _Optional[int] = ..., meets_filter: bool = ..., aggregations: _Optional[_Union[GroupAggregations, _Mapping]] = ...) -> None: ...

class TruncationNotice(_message.Message):
    __slots__ = ["reason", "tracked_action_count", "known_total_action_count", "message"]
    class Reason(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        REASON_UNSPECIFIED: _ClassVar[TruncationNotice.Reason]
        REASON_RUN_NODE_LIMIT: _ClassVar[TruncationNotice.Reason]
        REASON_PARENT_CHILD_LIMIT: _ClassVar[TruncationNotice.Reason]
        REASON_HYDRATING: _ClassVar[TruncationNotice.Reason]
    REASON_UNSPECIFIED: TruncationNotice.Reason
    REASON_RUN_NODE_LIMIT: TruncationNotice.Reason
    REASON_PARENT_CHILD_LIMIT: TruncationNotice.Reason
    REASON_HYDRATING: TruncationNotice.Reason
    REASON_FIELD_NUMBER: _ClassVar[int]
    TRACKED_ACTION_COUNT_FIELD_NUMBER: _ClassVar[int]
    KNOWN_TOTAL_ACTION_COUNT_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    reason: TruncationNotice.Reason
    tracked_action_count: int
    known_total_action_count: int
    message: str
    def __init__(self, reason: _Optional[_Union[TruncationNotice.Reason, str]] = ..., tracked_action_count: _Optional[int] = ..., known_total_action_count: _Optional[int] = ..., message: _Optional[str] = ...) -> None: ...

class WatchClusterEventsRequest(_message.Message):
    __slots__ = ["id", "attempt"]
    ID_FIELD_NUMBER: _ClassVar[int]
    ATTEMPT_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.ActionIdentifier
    attempt: int
    def __init__(self, id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., attempt: _Optional[int] = ...) -> None: ...

class WatchClusterEventsResponse(_message.Message):
    __slots__ = ["cluster_events"]
    CLUSTER_EVENTS_FIELD_NUMBER: _ClassVar[int]
    cluster_events: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.ClusterEvent]
    def __init__(self, cluster_events: _Optional[_Iterable[_Union[_run_definition_pb2.ClusterEvent, _Mapping]]] = ...) -> None: ...

class AbortActionRequest(_message.Message):
    __slots__ = ["action_id", "reason"]
    ACTION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    action_id: _identifier_pb2.ActionIdentifier
    reason: str
    def __init__(self, action_id: _Optional[_Union[_identifier_pb2.ActionIdentifier, _Mapping]] = ..., reason: _Optional[str] = ...) -> None: ...

class AbortActionResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class WatchGroupsRequest(_message.Message):
    __slots__ = ["project_id", "start_date", "end_date", "request", "known_sort_fields"]
    class KnownSortField(_message.Message):
        __slots__ = ["created_at"]
        CREATED_AT_FIELD_NUMBER: _ClassVar[int]
        created_at: _list_pb2.Sort.Direction
        def __init__(self, created_at: _Optional[_Union[_list_pb2.Sort.Direction, str]] = ...) -> None: ...
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    START_DATE_FIELD_NUMBER: _ClassVar[int]
    END_DATE_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    KNOWN_SORT_FIELDS_FIELD_NUMBER: _ClassVar[int]
    project_id: _identifier_pb2.ProjectIdentifier
    start_date: _timestamp_pb2.Timestamp
    end_date: _timestamp_pb2.Timestamp
    request: _list_pb2.ListRequest
    known_sort_fields: _containers.RepeatedCompositeFieldContainer[WatchGroupsRequest.KnownSortField]
    def __init__(self, project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ..., start_date: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., end_date: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., known_sort_fields: _Optional[_Iterable[_Union[WatchGroupsRequest.KnownSortField, _Mapping]]] = ...) -> None: ...

class WatchGroupsResponse(_message.Message):
    __slots__ = ["task_groups", "sentinel"]
    TASK_GROUPS_FIELD_NUMBER: _ClassVar[int]
    SENTINEL_FIELD_NUMBER: _ClassVar[int]
    task_groups: _containers.RepeatedCompositeFieldContainer[_run_definition_pb2.TaskGroup]
    sentinel: bool
    def __init__(self, task_groups: _Optional[_Iterable[_Union[_run_definition_pb2.TaskGroup, _Mapping]]] = ..., sentinel: bool = ...) -> None: ...
