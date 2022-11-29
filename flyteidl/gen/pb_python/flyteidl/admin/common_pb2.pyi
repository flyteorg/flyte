from flyteidl.core import execution_pb2 as _execution_pb2
from flyteidl.core import identifier_pb2 as _identifier_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor
NAMED_ENTITY_ACTIVE: NamedEntityState
NAMED_ENTITY_ARCHIVED: NamedEntityState
SYSTEM_GENERATED: NamedEntityState

class Annotations(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class AuthRole(_message.Message):
    __slots__ = ["assumable_iam_role", "kubernetes_service_account"]
    ASSUMABLE_IAM_ROLE_FIELD_NUMBER: _ClassVar[int]
    KUBERNETES_SERVICE_ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    assumable_iam_role: str
    kubernetes_service_account: str
    def __init__(self, assumable_iam_role: _Optional[str] = ..., kubernetes_service_account: _Optional[str] = ...) -> None: ...

class EmailNotification(_message.Message):
    __slots__ = ["recipients_email"]
    RECIPIENTS_EMAIL_FIELD_NUMBER: _ClassVar[int]
    recipients_email: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, recipients_email: _Optional[_Iterable[str]] = ...) -> None: ...

class Labels(_message.Message):
    __slots__ = ["values"]
    class ValuesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.ScalarMap[str, str]
    def __init__(self, values: _Optional[_Mapping[str, str]] = ...) -> None: ...

class NamedEntity(_message.Message):
    __slots__ = ["id", "metadata", "resource_type"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    id: NamedEntityIdentifier
    metadata: NamedEntityMetadata
    resource_type: _identifier_pb2.ResourceType
    def __init__(self, resource_type: _Optional[_Union[_identifier_pb2.ResourceType, str]] = ..., id: _Optional[_Union[NamedEntityIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[NamedEntityMetadata, _Mapping]] = ...) -> None: ...

class NamedEntityGetRequest(_message.Message):
    __slots__ = ["id", "resource_type"]
    ID_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    id: NamedEntityIdentifier
    resource_type: _identifier_pb2.ResourceType
    def __init__(self, resource_type: _Optional[_Union[_identifier_pb2.ResourceType, str]] = ..., id: _Optional[_Union[NamedEntityIdentifier, _Mapping]] = ...) -> None: ...

class NamedEntityIdentifier(_message.Message):
    __slots__ = ["domain", "name", "project"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    domain: str
    name: str
    project: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class NamedEntityIdentifierList(_message.Message):
    __slots__ = ["entities", "token"]
    ENTITIES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    entities: _containers.RepeatedCompositeFieldContainer[NamedEntityIdentifier]
    token: str
    def __init__(self, entities: _Optional[_Iterable[_Union[NamedEntityIdentifier, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class NamedEntityIdentifierListRequest(_message.Message):
    __slots__ = ["domain", "filters", "limit", "project", "sort_by", "token"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    domain: str
    filters: str
    limit: int
    project: str
    sort_by: Sort
    token: str
    def __init__(self, project: _Optional[str] = ..., domain: _Optional[str] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., sort_by: _Optional[_Union[Sort, _Mapping]] = ..., filters: _Optional[str] = ...) -> None: ...

class NamedEntityList(_message.Message):
    __slots__ = ["entities", "token"]
    ENTITIES_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    entities: _containers.RepeatedCompositeFieldContainer[NamedEntity]
    token: str
    def __init__(self, entities: _Optional[_Iterable[_Union[NamedEntity, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class NamedEntityListRequest(_message.Message):
    __slots__ = ["domain", "filters", "limit", "project", "resource_type", "sort_by", "token"]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    domain: str
    filters: str
    limit: int
    project: str
    resource_type: _identifier_pb2.ResourceType
    sort_by: Sort
    token: str
    def __init__(self, resource_type: _Optional[_Union[_identifier_pb2.ResourceType, str]] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., sort_by: _Optional[_Union[Sort, _Mapping]] = ..., filters: _Optional[str] = ...) -> None: ...

class NamedEntityMetadata(_message.Message):
    __slots__ = ["description", "state"]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    description: str
    state: NamedEntityState
    def __init__(self, description: _Optional[str] = ..., state: _Optional[_Union[NamedEntityState, str]] = ...) -> None: ...

class NamedEntityUpdateRequest(_message.Message):
    __slots__ = ["id", "metadata", "resource_type"]
    ID_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    RESOURCE_TYPE_FIELD_NUMBER: _ClassVar[int]
    id: NamedEntityIdentifier
    metadata: NamedEntityMetadata
    resource_type: _identifier_pb2.ResourceType
    def __init__(self, resource_type: _Optional[_Union[_identifier_pb2.ResourceType, str]] = ..., id: _Optional[_Union[NamedEntityIdentifier, _Mapping]] = ..., metadata: _Optional[_Union[NamedEntityMetadata, _Mapping]] = ...) -> None: ...

class NamedEntityUpdateResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class Notification(_message.Message):
    __slots__ = ["email", "pager_duty", "phases", "slack"]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    PAGER_DUTY_FIELD_NUMBER: _ClassVar[int]
    PHASES_FIELD_NUMBER: _ClassVar[int]
    SLACK_FIELD_NUMBER: _ClassVar[int]
    email: EmailNotification
    pager_duty: PagerDutyNotification
    phases: _containers.RepeatedScalarFieldContainer[_execution_pb2.WorkflowExecution.Phase]
    slack: SlackNotification
    def __init__(self, phases: _Optional[_Iterable[_Union[_execution_pb2.WorkflowExecution.Phase, str]]] = ..., email: _Optional[_Union[EmailNotification, _Mapping]] = ..., pager_duty: _Optional[_Union[PagerDutyNotification, _Mapping]] = ..., slack: _Optional[_Union[SlackNotification, _Mapping]] = ...) -> None: ...

class ObjectGetRequest(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _identifier_pb2.Identifier
    def __init__(self, id: _Optional[_Union[_identifier_pb2.Identifier, _Mapping]] = ...) -> None: ...

class PagerDutyNotification(_message.Message):
    __slots__ = ["recipients_email"]
    RECIPIENTS_EMAIL_FIELD_NUMBER: _ClassVar[int]
    recipients_email: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, recipients_email: _Optional[_Iterable[str]] = ...) -> None: ...

class RawOutputDataConfig(_message.Message):
    __slots__ = ["output_location_prefix"]
    OUTPUT_LOCATION_PREFIX_FIELD_NUMBER: _ClassVar[int]
    output_location_prefix: str
    def __init__(self, output_location_prefix: _Optional[str] = ...) -> None: ...

class ResourceListRequest(_message.Message):
    __slots__ = ["filters", "id", "limit", "sort_by", "token"]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    SORT_BY_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    filters: str
    id: NamedEntityIdentifier
    limit: int
    sort_by: Sort
    token: str
    def __init__(self, id: _Optional[_Union[NamedEntityIdentifier, _Mapping]] = ..., limit: _Optional[int] = ..., token: _Optional[str] = ..., filters: _Optional[str] = ..., sort_by: _Optional[_Union[Sort, _Mapping]] = ...) -> None: ...

class SlackNotification(_message.Message):
    __slots__ = ["recipients_email"]
    RECIPIENTS_EMAIL_FIELD_NUMBER: _ClassVar[int]
    recipients_email: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, recipients_email: _Optional[_Iterable[str]] = ...) -> None: ...

class Sort(_message.Message):
    __slots__ = ["direction", "key"]
    class Direction(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
    ASCENDING: Sort.Direction
    DESCENDING: Sort.Direction
    DIRECTION_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    direction: Sort.Direction
    key: str
    def __init__(self, key: _Optional[str] = ..., direction: _Optional[_Union[Sort.Direction, str]] = ...) -> None: ...

class UrlBlob(_message.Message):
    __slots__ = ["bytes", "url"]
    BYTES_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    bytes: int
    url: str
    def __init__(self, url: _Optional[str] = ..., bytes: _Optional[int] = ...) -> None: ...

class NamedEntityState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
