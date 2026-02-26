from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class NotificationType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    NOTIFICATION_TYPE_UNSPECIFIED: _ClassVar[NotificationType]
    NOTIFICATION_TYPE_WEBHOOK: _ClassVar[NotificationType]
    NOTIFICATION_TYPE_EMAIL: _ClassVar[NotificationType]
NOTIFICATION_TYPE_UNSPECIFIED: NotificationType
NOTIFICATION_TYPE_WEBHOOK: NotificationType
NOTIFICATION_TYPE_EMAIL: NotificationType

class RuleId(_message.Message):
    __slots__ = ["org", "project", "domain", "name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    name: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class Rule(_message.Message):
    __slots__ = ["rule_id", "delivery_configs", "action_rule"]
    RULE_ID_FIELD_NUMBER: _ClassVar[int]
    DELIVERY_CONFIGS_FIELD_NUMBER: _ClassVar[int]
    ACTION_RULE_FIELD_NUMBER: _ClassVar[int]
    rule_id: RuleId
    delivery_configs: _containers.RepeatedCompositeFieldContainer[DeliveryConfig]
    action_rule: ActionRule
    def __init__(self, rule_id: _Optional[_Union[RuleId, _Mapping]] = ..., delivery_configs: _Optional[_Iterable[_Union[DeliveryConfig, _Mapping]]] = ..., action_rule: _Optional[_Union[ActionRule, _Mapping]] = ...) -> None: ...

class ActionRule(_message.Message):
    __slots__ = ["task_name_regex", "phase_regex"]
    TASK_NAME_REGEX_FIELD_NUMBER: _ClassVar[int]
    PHASE_REGEX_FIELD_NUMBER: _ClassVar[int]
    task_name_regex: str
    phase_regex: str
    def __init__(self, task_name_regex: _Optional[str] = ..., phase_regex: _Optional[str] = ...) -> None: ...

class DeliveryConfigId(_message.Message):
    __slots__ = ["org", "project", "domain", "name"]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    org: str
    project: str
    domain: str
    name: str
    def __init__(self, org: _Optional[str] = ..., project: _Optional[str] = ..., domain: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class DeliveryConfig(_message.Message):
    __slots__ = ["delivery_config_id", "type", "webhook_config", "email_config"]
    DELIVERY_CONFIG_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    WEBHOOK_CONFIG_FIELD_NUMBER: _ClassVar[int]
    EMAIL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    delivery_config_id: DeliveryConfigId
    type: NotificationType
    webhook_config: WebhookDeliveryConfig
    email_config: EmailDeliveryConfig
    def __init__(self, delivery_config_id: _Optional[_Union[DeliveryConfigId, _Mapping]] = ..., type: _Optional[_Union[NotificationType, str]] = ..., webhook_config: _Optional[_Union[WebhookDeliveryConfig, _Mapping]] = ..., email_config: _Optional[_Union[EmailDeliveryConfig, _Mapping]] = ...) -> None: ...

class WebhookDeliveryConfig(_message.Message):
    __slots__ = ["url", "method", "headers", "body"]
    class HeadersEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    URL_FIELD_NUMBER: _ClassVar[int]
    METHOD_FIELD_NUMBER: _ClassVar[int]
    HEADERS_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    url: str
    method: str
    headers: _containers.ScalarMap[str, str]
    body: str
    def __init__(self, url: _Optional[str] = ..., method: _Optional[str] = ..., headers: _Optional[_Mapping[str, str]] = ..., body: _Optional[str] = ...) -> None: ...

class EmailDeliveryConfig(_message.Message):
    __slots__ = ["subject", "to", "cc", "bcc", "html", "text"]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    CC_FIELD_NUMBER: _ClassVar[int]
    BCC_FIELD_NUMBER: _ClassVar[int]
    HTML_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    subject: str
    to: _containers.RepeatedScalarFieldContainer[str]
    cc: _containers.RepeatedScalarFieldContainer[str]
    bcc: _containers.RepeatedScalarFieldContainer[str]
    html: str
    text: str
    def __init__(self, subject: _Optional[str] = ..., to: _Optional[_Iterable[str]] = ..., cc: _Optional[_Iterable[str]] = ..., bcc: _Optional[_Iterable[str]] = ..., html: _Optional[str] = ..., text: _Optional[str] = ...) -> None: ...
