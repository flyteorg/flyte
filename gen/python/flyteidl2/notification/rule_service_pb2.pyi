from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.notification import definition_pb2 as _definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SaveRuleRequest(_message.Message):
    __slots__ = ["rule"]
    RULE_FIELD_NUMBER: _ClassVar[int]
    rule: _definition_pb2.Rule
    def __init__(self, rule: _Optional[_Union[_definition_pb2.Rule, _Mapping]] = ...) -> None: ...

class SaveRuleResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetRuleRequest(_message.Message):
    __slots__ = ["rule_id"]
    RULE_ID_FIELD_NUMBER: _ClassVar[int]
    rule_id: _definition_pb2.RuleId
    def __init__(self, rule_id: _Optional[_Union[_definition_pb2.RuleId, _Mapping]] = ...) -> None: ...

class GetRuleResponse(_message.Message):
    __slots__ = ["rule"]
    RULE_FIELD_NUMBER: _ClassVar[int]
    rule: _definition_pb2.Rule
    def __init__(self, rule: _Optional[_Union[_definition_pb2.Rule, _Mapping]] = ...) -> None: ...

class ListRulesRequest(_message.Message):
    __slots__ = ["request", "org", "project_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    project_id: _identifier_pb2.ProjectIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ...) -> None: ...

class ListRulesResponse(_message.Message):
    __slots__ = ["rules", "token"]
    RULES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    rules: _containers.RepeatedCompositeFieldContainer[_definition_pb2.Rule]
    token: str
    def __init__(self, rules: _Optional[_Iterable[_Union[_definition_pb2.Rule, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class DeleteRuleRequest(_message.Message):
    __slots__ = ["rule_id"]
    RULE_ID_FIELD_NUMBER: _ClassVar[int]
    rule_id: _definition_pb2.RuleId
    def __init__(self, rule_id: _Optional[_Union[_definition_pb2.RuleId, _Mapping]] = ...) -> None: ...

class DeleteRuleResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
