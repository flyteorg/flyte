from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import phase_pb2 as _phase_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    EVENT_TYPE_UNSPECIFIED: _ClassVar[EventType]
    EVENT_TYPE_RUN_COMPLETED: _ClassVar[EventType]

class HttpMethod(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    HTTP_METHOD_UNSPECIFIED: _ClassVar[HttpMethod]
    HTTP_METHOD_GET: _ClassVar[HttpMethod]
    HTTP_METHOD_HEAD: _ClassVar[HttpMethod]
    HTTP_METHOD_POST: _ClassVar[HttpMethod]
    HTTP_METHOD_PUT: _ClassVar[HttpMethod]
    HTTP_METHOD_DELETE: _ClassVar[HttpMethod]
    HTTP_METHOD_CONNECT: _ClassVar[HttpMethod]
    HTTP_METHOD_OPTIONS: _ClassVar[HttpMethod]
    HTTP_METHOD_TRACE: _ClassVar[HttpMethod]
    HTTP_METHOD_PATCH: _ClassVar[HttpMethod]
EVENT_TYPE_UNSPECIFIED: EventType
EVENT_TYPE_RUN_COMPLETED: EventType
HTTP_METHOD_UNSPECIFIED: HttpMethod
HTTP_METHOD_GET: HttpMethod
HTTP_METHOD_HEAD: HttpMethod
HTTP_METHOD_POST: HttpMethod
HTTP_METHOD_PUT: HttpMethod
HTTP_METHOD_DELETE: HttpMethod
HTTP_METHOD_CONNECT: HttpMethod
HTTP_METHOD_OPTIONS: HttpMethod
HTTP_METHOD_TRACE: HttpMethod
HTTP_METHOD_PATCH: HttpMethod

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
    __slots__ = ["event_type", "run_rule"]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    RUN_RULE_FIELD_NUMBER: _ClassVar[int]
    event_type: EventType
    run_rule: RunCompletedRule
    def __init__(self, event_type: _Optional[_Union[EventType, str]] = ..., run_rule: _Optional[_Union[RunCompletedRule, _Mapping]] = ...) -> None: ...

class RunCompletedRule(_message.Message):
    __slots__ = ["rule_id", "delivery_config_ids", "checks"]
    RULE_ID_FIELD_NUMBER: _ClassVar[int]
    DELIVERY_CONFIG_IDS_FIELD_NUMBER: _ClassVar[int]
    CHECKS_FIELD_NUMBER: _ClassVar[int]
    rule_id: RuleId
    delivery_config_ids: _containers.RepeatedCompositeFieldContainer[DeliveryConfigId]
    checks: RunCompletedRuleChecks
    def __init__(self, rule_id: _Optional[_Union[RuleId, _Mapping]] = ..., delivery_config_ids: _Optional[_Iterable[_Union[DeliveryConfigId, _Mapping]]] = ..., checks: _Optional[_Union[RunCompletedRuleChecks, _Mapping]] = ...) -> None: ...

class RunCompletedRuleChecks(_message.Message):
    __slots__ = ["project_regex", "domain_regex", "task_name_regex", "phase_regex"]
    PROJECT_REGEX_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_REGEX_FIELD_NUMBER: _ClassVar[int]
    TASK_NAME_REGEX_FIELD_NUMBER: _ClassVar[int]
    PHASE_REGEX_FIELD_NUMBER: _ClassVar[int]
    project_regex: str
    domain_regex: str
    task_name_regex: str
    phase_regex: str
    def __init__(self, project_regex: _Optional[str] = ..., domain_regex: _Optional[str] = ..., task_name_regex: _Optional[str] = ..., phase_regex: _Optional[str] = ...) -> None: ...

class DeliveryOption(_message.Message):
    __slots__ = ["config_id", "config"]
    CONFIG_ID_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    config_id: DeliveryConfigId
    config: DeliveryConfig
    def __init__(self, config_id: _Optional[_Union[DeliveryConfigId, _Mapping]] = ..., config: _Optional[_Union[DeliveryConfig, _Mapping]] = ...) -> None: ...

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

class DeliveryConfigTemplate(_message.Message):
    __slots__ = ["webhook", "email"]
    WEBHOOK_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    webhook: WebhookDeliveryTemplate
    email: EmailDeliveryTemplate
    def __init__(self, webhook: _Optional[_Union[WebhookDeliveryTemplate, _Mapping]] = ..., email: _Optional[_Union[EmailDeliveryTemplate, _Mapping]] = ...) -> None: ...

class RunCompletedNotificationTemplateData(_message.Message):
    __slots__ = ["run", "phase", "error"]
    RUN_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    run: _identifier_pb2.RunIdentifier
    phase: _phase_pb2.ActionPhase
    error: str
    def __init__(self, run: _Optional[_Union[_identifier_pb2.RunIdentifier, _Mapping]] = ..., phase: _Optional[_Union[_phase_pb2.ActionPhase, str]] = ..., error: _Optional[str] = ...) -> None: ...

class DeliveryConfig(_message.Message):
    __slots__ = ["delivery_config_id", "event_type", "template"]
    DELIVERY_CONFIG_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    delivery_config_id: DeliveryConfigId
    event_type: EventType
    template: DeliveryConfigTemplate
    def __init__(self, delivery_config_id: _Optional[_Union[DeliveryConfigId, _Mapping]] = ..., event_type: _Optional[_Union[EventType, str]] = ..., template: _Optional[_Union[DeliveryConfigTemplate, _Mapping]] = ...) -> None: ...

class WebhookDeliveryTemplate(_message.Message):
    __slots__ = ["url", "method", "headers", "body_template"]
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
    BODY_TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    url: str
    method: HttpMethod
    headers: _containers.ScalarMap[str, str]
    body_template: str
    def __init__(self, url: _Optional[str] = ..., method: _Optional[_Union[HttpMethod, str]] = ..., headers: _Optional[_Mapping[str, str]] = ..., body_template: _Optional[str] = ...) -> None: ...

class EmailDeliveryTemplate(_message.Message):
    __slots__ = ["subject", "to", "cc", "bcc", "html_template", "text_template"]
    SUBJECT_FIELD_NUMBER: _ClassVar[int]
    TO_FIELD_NUMBER: _ClassVar[int]
    CC_FIELD_NUMBER: _ClassVar[int]
    BCC_FIELD_NUMBER: _ClassVar[int]
    HTML_TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    TEXT_TEMPLATE_FIELD_NUMBER: _ClassVar[int]
    subject: str
    to: _containers.RepeatedScalarFieldContainer[str]
    cc: _containers.RepeatedScalarFieldContainer[str]
    bcc: _containers.RepeatedScalarFieldContainer[str]
    html_template: str
    text_template: str
    def __init__(self, subject: _Optional[str] = ..., to: _Optional[_Iterable[str]] = ..., cc: _Optional[_Iterable[str]] = ..., bcc: _Optional[_Iterable[str]] = ..., html_template: _Optional[str] = ..., text_template: _Optional[str] = ...) -> None: ...
