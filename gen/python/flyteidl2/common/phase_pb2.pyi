from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class ActionPhase(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    ACTION_PHASE_UNSPECIFIED: _ClassVar[ActionPhase]
    ACTION_PHASE_QUEUED: _ClassVar[ActionPhase]
    ACTION_PHASE_WAITING_FOR_RESOURCES: _ClassVar[ActionPhase]
    ACTION_PHASE_INITIALIZING: _ClassVar[ActionPhase]
    ACTION_PHASE_RUNNING: _ClassVar[ActionPhase]
    ACTION_PHASE_SUCCEEDED: _ClassVar[ActionPhase]
    ACTION_PHASE_FAILED: _ClassVar[ActionPhase]
    ACTION_PHASE_ABORTED: _ClassVar[ActionPhase]
    ACTION_PHASE_TIMED_OUT: _ClassVar[ActionPhase]
ACTION_PHASE_UNSPECIFIED: ActionPhase
ACTION_PHASE_QUEUED: ActionPhase
ACTION_PHASE_WAITING_FOR_RESOURCES: ActionPhase
ACTION_PHASE_INITIALIZING: ActionPhase
ACTION_PHASE_RUNNING: ActionPhase
ACTION_PHASE_SUCCEEDED: ActionPhase
ACTION_PHASE_FAILED: ActionPhase
ACTION_PHASE_ABORTED: ActionPhase
ACTION_PHASE_TIMED_OUT: ActionPhase
