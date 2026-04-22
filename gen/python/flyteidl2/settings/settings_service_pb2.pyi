from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.settings import settings_definition_pb2 as _settings_definition_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SettingsRecord(_message.Message):
    __slots__ = ["key", "settings", "version"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    SETTINGS_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    key: _settings_definition_pb2.SettingsKey
    settings: _settings_definition_pb2.Settings
    version: int
    def __init__(self, key: _Optional[_Union[_settings_definition_pb2.SettingsKey, _Mapping]] = ..., settings: _Optional[_Union[_settings_definition_pb2.Settings, _Mapping]] = ..., version: _Optional[int] = ...) -> None: ...

class GetSettingsRequest(_message.Message):
    __slots__ = ["key"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: _settings_definition_pb2.SettingsKey
    def __init__(self, key: _Optional[_Union[_settings_definition_pb2.SettingsKey, _Mapping]] = ...) -> None: ...

class GetSettingsResponse(_message.Message):
    __slots__ = ["settingsRecord"]
    SETTINGSRECORD_FIELD_NUMBER: _ClassVar[int]
    settingsRecord: SettingsRecord
    def __init__(self, settingsRecord: _Optional[_Union[SettingsRecord, _Mapping]] = ...) -> None: ...

class GetSettingsForEditRequest(_message.Message):
    __slots__ = ["key"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    key: _settings_definition_pb2.SettingsKey
    def __init__(self, key: _Optional[_Union[_settings_definition_pb2.SettingsKey, _Mapping]] = ...) -> None: ...

class GetSettingsForEditResponse(_message.Message):
    __slots__ = ["requestedKey", "levels"]
    REQUESTEDKEY_FIELD_NUMBER: _ClassVar[int]
    LEVELS_FIELD_NUMBER: _ClassVar[int]
    requestedKey: _settings_definition_pb2.SettingsKey
    levels: _containers.RepeatedCompositeFieldContainer[SettingsRecord]
    def __init__(self, requestedKey: _Optional[_Union[_settings_definition_pb2.SettingsKey, _Mapping]] = ..., levels: _Optional[_Iterable[_Union[SettingsRecord, _Mapping]]] = ...) -> None: ...

class CreateSettingsRequest(_message.Message):
    __slots__ = ["key", "settings"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    SETTINGS_FIELD_NUMBER: _ClassVar[int]
    key: _settings_definition_pb2.SettingsKey
    settings: _settings_definition_pb2.Settings
    def __init__(self, key: _Optional[_Union[_settings_definition_pb2.SettingsKey, _Mapping]] = ..., settings: _Optional[_Union[_settings_definition_pb2.Settings, _Mapping]] = ...) -> None: ...

class CreateSettingsResponse(_message.Message):
    __slots__ = ["settingsRecord"]
    SETTINGSRECORD_FIELD_NUMBER: _ClassVar[int]
    settingsRecord: SettingsRecord
    def __init__(self, settingsRecord: _Optional[_Union[SettingsRecord, _Mapping]] = ...) -> None: ...

class UpdateSettingsRequest(_message.Message):
    __slots__ = ["key", "settings", "version"]
    KEY_FIELD_NUMBER: _ClassVar[int]
    SETTINGS_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    key: _settings_definition_pb2.SettingsKey
    settings: _settings_definition_pb2.Settings
    version: int
    def __init__(self, key: _Optional[_Union[_settings_definition_pb2.SettingsKey, _Mapping]] = ..., settings: _Optional[_Union[_settings_definition_pb2.Settings, _Mapping]] = ..., version: _Optional[int] = ...) -> None: ...

class UpdateSettingsResponse(_message.Message):
    __slots__ = ["settingsRecord"]
    SETTINGSRECORD_FIELD_NUMBER: _ClassVar[int]
    settingsRecord: SettingsRecord
    def __init__(self, settingsRecord: _Optional[_Union[SettingsRecord, _Mapping]] = ...) -> None: ...
