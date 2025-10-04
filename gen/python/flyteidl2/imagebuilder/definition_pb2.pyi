from buf.validate import validate_pb2 as _validate_pb2
from flyteidl2.core import security_pb2 as _security_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ImageIdentifier(_message.Message):
    __slots__ = ["name"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    name: str
    def __init__(self, name: _Optional[str] = ...) -> None: ...

class Image(_message.Message):
    __slots__ = ["id", "fqin"]
    ID_FIELD_NUMBER: _ClassVar[int]
    FQIN_FIELD_NUMBER: _ClassVar[int]
    id: ImageIdentifier
    fqin: str
    def __init__(self, id: _Optional[_Union[ImageIdentifier, _Mapping]] = ..., fqin: _Optional[str] = ...) -> None: ...

class AptPackages(_message.Message):
    __slots__ = ["packages", "secret_mounts"]
    PACKAGES_FIELD_NUMBER: _ClassVar[int]
    SECRET_MOUNTS_FIELD_NUMBER: _ClassVar[int]
    packages: _containers.RepeatedScalarFieldContainer[str]
    secret_mounts: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    def __init__(self, packages: _Optional[_Iterable[str]] = ..., secret_mounts: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ...) -> None: ...

class PipOptions(_message.Message):
    __slots__ = ["index_url", "extra_index_urls", "pre", "extra_args"]
    INDEX_URL_FIELD_NUMBER: _ClassVar[int]
    EXTRA_INDEX_URLS_FIELD_NUMBER: _ClassVar[int]
    PRE_FIELD_NUMBER: _ClassVar[int]
    EXTRA_ARGS_FIELD_NUMBER: _ClassVar[int]
    index_url: str
    extra_index_urls: _containers.RepeatedScalarFieldContainer[str]
    pre: bool
    extra_args: str
    def __init__(self, index_url: _Optional[str] = ..., extra_index_urls: _Optional[_Iterable[str]] = ..., pre: bool = ..., extra_args: _Optional[str] = ...) -> None: ...

class PipPackages(_message.Message):
    __slots__ = ["packages", "options", "secret_mounts"]
    PACKAGES_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    SECRET_MOUNTS_FIELD_NUMBER: _ClassVar[int]
    packages: _containers.RepeatedScalarFieldContainer[str]
    options: PipOptions
    secret_mounts: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    def __init__(self, packages: _Optional[_Iterable[str]] = ..., options: _Optional[_Union[PipOptions, _Mapping]] = ..., secret_mounts: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ...) -> None: ...

class Requirements(_message.Message):
    __slots__ = ["file", "options", "secret_mounts"]
    FILE_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    SECRET_MOUNTS_FIELD_NUMBER: _ClassVar[int]
    file: str
    options: PipOptions
    secret_mounts: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    def __init__(self, file: _Optional[str] = ..., options: _Optional[_Union[PipOptions, _Mapping]] = ..., secret_mounts: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ...) -> None: ...

class PythonWheels(_message.Message):
    __slots__ = ["dir", "options", "secret_mounts"]
    DIR_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    SECRET_MOUNTS_FIELD_NUMBER: _ClassVar[int]
    dir: str
    options: PipOptions
    secret_mounts: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    def __init__(self, dir: _Optional[str] = ..., options: _Optional[_Union[PipOptions, _Mapping]] = ..., secret_mounts: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ...) -> None: ...

class UVProject(_message.Message):
    __slots__ = ["pyproject", "uvlock", "options", "secret_mounts"]
    PYPROJECT_FIELD_NUMBER: _ClassVar[int]
    UVLOCK_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    SECRET_MOUNTS_FIELD_NUMBER: _ClassVar[int]
    pyproject: str
    uvlock: str
    options: PipOptions
    secret_mounts: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    def __init__(self, pyproject: _Optional[str] = ..., uvlock: _Optional[str] = ..., options: _Optional[_Union[PipOptions, _Mapping]] = ..., secret_mounts: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ...) -> None: ...

class Commands(_message.Message):
    __slots__ = ["cmd", "secret_mounts"]
    CMD_FIELD_NUMBER: _ClassVar[int]
    SECRET_MOUNTS_FIELD_NUMBER: _ClassVar[int]
    cmd: _containers.RepeatedScalarFieldContainer[str]
    secret_mounts: _containers.RepeatedCompositeFieldContainer[_security_pb2.Secret]
    def __init__(self, cmd: _Optional[_Iterable[str]] = ..., secret_mounts: _Optional[_Iterable[_Union[_security_pb2.Secret, _Mapping]]] = ...) -> None: ...

class WorkDir(_message.Message):
    __slots__ = ["workdir"]
    WORKDIR_FIELD_NUMBER: _ClassVar[int]
    workdir: str
    def __init__(self, workdir: _Optional[str] = ...) -> None: ...

class CopyConfig(_message.Message):
    __slots__ = ["src", "dst"]
    SRC_FIELD_NUMBER: _ClassVar[int]
    DST_FIELD_NUMBER: _ClassVar[int]
    src: str
    dst: str
    def __init__(self, src: _Optional[str] = ..., dst: _Optional[str] = ...) -> None: ...

class Env(_message.Message):
    __slots__ = ["env_variables"]
    class EnvVariablesEntry(_message.Message):
        __slots__ = ["key", "value"]
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    ENV_VARIABLES_FIELD_NUMBER: _ClassVar[int]
    env_variables: _containers.ScalarMap[str, str]
    def __init__(self, env_variables: _Optional[_Mapping[str, str]] = ...) -> None: ...

class Layer(_message.Message):
    __slots__ = ["apt_packages", "pip_packages", "commands", "requirements", "python_wheels", "workdir", "copy_config", "uv_project", "env"]
    APT_PACKAGES_FIELD_NUMBER: _ClassVar[int]
    PIP_PACKAGES_FIELD_NUMBER: _ClassVar[int]
    COMMANDS_FIELD_NUMBER: _ClassVar[int]
    REQUIREMENTS_FIELD_NUMBER: _ClassVar[int]
    PYTHON_WHEELS_FIELD_NUMBER: _ClassVar[int]
    WORKDIR_FIELD_NUMBER: _ClassVar[int]
    COPY_CONFIG_FIELD_NUMBER: _ClassVar[int]
    UV_PROJECT_FIELD_NUMBER: _ClassVar[int]
    ENV_FIELD_NUMBER: _ClassVar[int]
    apt_packages: AptPackages
    pip_packages: PipPackages
    commands: Commands
    requirements: Requirements
    python_wheels: PythonWheels
    workdir: WorkDir
    copy_config: CopyConfig
    uv_project: UVProject
    env: Env
    def __init__(self, apt_packages: _Optional[_Union[AptPackages, _Mapping]] = ..., pip_packages: _Optional[_Union[PipPackages, _Mapping]] = ..., commands: _Optional[_Union[Commands, _Mapping]] = ..., requirements: _Optional[_Union[Requirements, _Mapping]] = ..., python_wheels: _Optional[_Union[PythonWheels, _Mapping]] = ..., workdir: _Optional[_Union[WorkDir, _Mapping]] = ..., copy_config: _Optional[_Union[CopyConfig, _Mapping]] = ..., uv_project: _Optional[_Union[UVProject, _Mapping]] = ..., env: _Optional[_Union[Env, _Mapping]] = ...) -> None: ...

class ImageSpec(_message.Message):
    __slots__ = ["base_image", "python_version", "layers", "platform"]
    BASE_IMAGE_FIELD_NUMBER: _ClassVar[int]
    PYTHON_VERSION_FIELD_NUMBER: _ClassVar[int]
    LAYERS_FIELD_NUMBER: _ClassVar[int]
    PLATFORM_FIELD_NUMBER: _ClassVar[int]
    base_image: str
    python_version: str
    layers: _containers.RepeatedCompositeFieldContainer[Layer]
    platform: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, base_image: _Optional[str] = ..., python_version: _Optional[str] = ..., layers: _Optional[_Iterable[_Union[Layer, _Mapping]]] = ..., platform: _Optional[_Iterable[str]] = ...) -> None: ...
