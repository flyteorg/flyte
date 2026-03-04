from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.common import identifier_pb2 as _identifier_pb2
from flyteidl2.common import list_pb2 as _list_pb2
from flyteidl2.notification import definition_pb2 as _definition_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SaveDeliveryConfigRequest(_message.Message):
    __slots__ = ["delivery_config"]
    DELIVERY_CONFIG_FIELD_NUMBER: _ClassVar[int]
    delivery_config: _definition_pb2.DeliveryConfig
    def __init__(self, delivery_config: _Optional[_Union[_definition_pb2.DeliveryConfig, _Mapping]] = ...) -> None: ...

class SaveDeliveryConfigResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetDeliveryConfigRequest(_message.Message):
    __slots__ = ["delivery_config_id"]
    DELIVERY_CONFIG_ID_FIELD_NUMBER: _ClassVar[int]
    delivery_config_id: _definition_pb2.DeliveryConfigId
    def __init__(self, delivery_config_id: _Optional[_Union[_definition_pb2.DeliveryConfigId, _Mapping]] = ...) -> None: ...

class GetDeliveryConfigResponse(_message.Message):
    __slots__ = ["delivery_config"]
    DELIVERY_CONFIG_FIELD_NUMBER: _ClassVar[int]
    delivery_config: _definition_pb2.DeliveryConfig
    def __init__(self, delivery_config: _Optional[_Union[_definition_pb2.DeliveryConfig, _Mapping]] = ...) -> None: ...

class ListDeliveryConfigsRequest(_message.Message):
    __slots__ = ["request", "org", "project_id"]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ORG_FIELD_NUMBER: _ClassVar[int]
    PROJECT_ID_FIELD_NUMBER: _ClassVar[int]
    request: _list_pb2.ListRequest
    org: str
    project_id: _identifier_pb2.ProjectIdentifier
    def __init__(self, request: _Optional[_Union[_list_pb2.ListRequest, _Mapping]] = ..., org: _Optional[str] = ..., project_id: _Optional[_Union[_identifier_pb2.ProjectIdentifier, _Mapping]] = ...) -> None: ...

class ListDeliveryConfigsResponse(_message.Message):
    __slots__ = ["delivery_configs", "token"]
    DELIVERY_CONFIGS_FIELD_NUMBER: _ClassVar[int]
    TOKEN_FIELD_NUMBER: _ClassVar[int]
    delivery_configs: _containers.RepeatedCompositeFieldContainer[_definition_pb2.DeliveryConfig]
    token: str
    def __init__(self, delivery_configs: _Optional[_Iterable[_Union[_definition_pb2.DeliveryConfig, _Mapping]]] = ..., token: _Optional[str] = ...) -> None: ...

class DeleteDeliveryConfigRequest(_message.Message):
    __slots__ = ["delivery_config_id"]
    DELIVERY_CONFIG_ID_FIELD_NUMBER: _ClassVar[int]
    delivery_config_id: _definition_pb2.DeliveryConfigId
    def __init__(self, delivery_config_id: _Optional[_Union[_definition_pb2.DeliveryConfigId, _Mapping]] = ...) -> None: ...

class DeleteDeliveryConfigResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class SendTestNotificationRequest(_message.Message):
    __slots__ = ["template_data", "delivery_option"]
    TEMPLATE_DATA_FIELD_NUMBER: _ClassVar[int]
    DELIVERY_OPTION_FIELD_NUMBER: _ClassVar[int]
    template_data: _struct_pb2.Struct
    delivery_option: _definition_pb2.DeliveryOption
    def __init__(self, template_data: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., delivery_option: _Optional[_Union[_definition_pb2.DeliveryOption, _Mapping]] = ...) -> None: ...

class SendTestNotificationResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
