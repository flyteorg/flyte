import json as _json
import typing

from flyteidl.admin import task_pb2 as _admin_task
from flyteidl.core import compiler_pb2 as _compiler
from flyteidl.core import literals_pb2 as _literals_pb2
from flyteidl.core import tasks_pb2 as _core_task
from google.protobuf import json_format as _json_format
from google.protobuf import struct_pb2 as _struct

from flytekit.models import common as _common
from flytekit.models import interface as _interface
from flytekit.models import literals as _literals
from flytekit.models import security as _sec
from flytekit.models.core import identifier as _identifier
from flytekit.models.documentation import Documentation


class Resources(_common.FlyteIdlEntity):
    class ResourceName(object):
        UNKNOWN = _core_task.Resources.UNKNOWN
        CPU = _core_task.Resources.CPU
        GPU = _core_task.Resources.GPU
        MEMORY = _core_task.Resources.MEMORY
        EPHEMERAL_STORAGE = _core_task.Resources.EPHEMERAL_STORAGE

    class ResourceEntry(_common.FlyteIdlEntity):
        def __init__(self, name, value):
            """
            :param int name: enum value from ResourceName
            :param Text value: a textual value describing the resource need.  Must be a valid k8s quantity.
            """
            self._name = name
            self._value = value

        @property
        def name(self):
            """
            enum value from ResourceName
            :rtype: int
            """
            return self._name

        @property
        def value(self):
            """
            A textual value describing the resource need.  Must be a valid k8s quantity.
            :rtype: Text
            """
            return self._value

        def to_flyte_idl(self):
            """
            :rtype: flyteidl.core.tasks_pb2.ResourceEntry
            """
            return _core_task.Resources.ResourceEntry(name=self.name, value=self.value)

        @classmethod
        def from_flyte_idl(cls, pb2_object):
            """
            :param flyteidl.core.tasks_pb2.Resources.ResourceEntry pb2_object:
            :rtype: Resources.ResourceEntry
            """
            return cls(name=pb2_object.name, value=pb2_object.value)

    def __init__(self, requests, limits):
        """
        :param list[Resources.ResourceEntry] requests: The desired resources for execution.  This is given on a best
            effort basis.
        :param list[Resources.ResourceEntry] limits: These are the limits required.  These are guaranteed to be
            satisfied.
        """
        self._requests = requests
        self._limits = limits

    @property
    def requests(self):
        """
        The desired resources for execution.  This is given on a best effort basis.
        :rtype: list[Resources.ResourceEntry]
        """
        return self._requests

    @property
    def limits(self):
        """
        These are the limits required.  These are guaranteed to be satisfied.
        :rtype: list[Resources.ResourceEntry]
        """
        return self._limits

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.tasks_pb2.Resources
        """
        return _core_task.Resources(
            requests=[r.to_flyte_idl() for r in self.requests],
            limits=[r.to_flyte_idl() for r in self.limits],
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.tasks_pb2.Resources.ResourceEntry pb2_object:
        :rtype: Resources
        """
        return cls(
            requests=[Resources.ResourceEntry.from_flyte_idl(r) for r in pb2_object.requests],
            limits=[Resources.ResourceEntry.from_flyte_idl(l) for l in pb2_object.limits],
        )


class RuntimeMetadata(_common.FlyteIdlEntity):
    class RuntimeType(object):
        OTHER = 0
        FLYTE_SDK = 1

    def __init__(self, type, version, flavor):
        """
        :param int type: Enum type from RuntimeMetadata.RuntimeType
        :param Text version: Version string for SDK version.  Can be used for metrics or managing breaking changes in
            Admin or Propeller
        :param Text flavor: Optional extra information about runtime environment (e.g. Python, GoLang, etc.)
        """
        self._type = type
        self._version = version
        self._flavor = flavor

    @property
    def type(self):
        """
        Enum type from RuntimeMetadata.RuntimeType
        :rtype: int
        """
        return self._type

    @property
    def version(self):
        """
        Version string for SDK version.  Can be used for metrics or managing breaking changes in Admin or Propeller
        :rtype: Text
        """
        return self._version

    @property
    def flavor(self):
        """
        Optional extra information about runtime environment (e.g. Python, GoLang, etc.)
        :rtype: Text
        """
        return self._flavor

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.tasks_pb2.RuntimeMetadata
        """
        return _core_task.RuntimeMetadata(type=self.type, version=self.version, flavor=self.flavor)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.tasks_pb2.RuntimeMetadata pb2_object:
        :rtype: RuntimeMetadata
        """
        return cls(type=pb2_object.type, version=pb2_object.version, flavor=pb2_object.flavor)


class TaskMetadata(_common.FlyteIdlEntity):
    def __init__(
        self,
        discoverable,
        runtime,
        timeout,
        retries,
        interruptible,
        discovery_version,
        deprecated_error_message,
        cache_serializable,
        pod_template_name,
    ):
        """
        Information needed at runtime to determine behavior such as whether or not outputs are discoverable, timeouts,
        and retries.

        :param bool discoverable: Whether or not the outputs of this task should be cached for discovery.
        :param RuntimeMetadata runtime: Metadata describing the runtime environment for this task.
        :param datetime.timedelta timeout: The amount of time to wait before timing out.  This includes queuing and
            scheduler latency.
        :param bool interruptible: Whether or not the task is interruptible.
        :param flytekit.models.literals.RetryStrategy retries: Retry strategy for this task.  0 retries means only
            try once.
        :param Text discovery_version: This is the version used to create a logical version for data in the cache.
            This is only used when `discoverable` is true.  Data is considered discoverable if: the inputs to a given
            task are the same and the discovery_version is also the same.
        :param Text deprecated: This string can be used to mark the task as deprecated.  Consumers of the task will
            receive deprecation warnings.
        :param bool cache_serializable: Whether or not caching operations are executed in serial. This means only a
            single instance over identical inputs is executed, other concurrent executions wait for the cached results.
        :param pod_template_name: The name of the existing PodTemplate resource which will be used in this task.
        """
        self._discoverable = discoverable
        self._runtime = runtime
        self._timeout = timeout
        self._interruptible = interruptible
        self._retries = retries
        self._discovery_version = discovery_version
        self._deprecated_error_message = deprecated_error_message
        self._cache_serializable = cache_serializable
        self._pod_template_name = pod_template_name

    @property
    def discoverable(self):
        """
        Whether or not the outputs of this task should be cached for discovery.
        :rtype: bool
        """
        return self._discoverable

    @property
    def runtime(self):
        """
        Metadata describing the runtime environment for this task.
        :rtype: RuntimeMetadata
        """
        return self._runtime

    @property
    def retries(self):
        """
        Retry strategy for this task.  0 retries means only try once.
        :rtype: flytekit.models.literals.RetryStrategy
        """
        return self._retries

    @property
    def timeout(self):
        """
        The amount of time to wait before timing out.  This includes queuing and scheduler latency.
        :rtype: datetime.timedelta
        """
        return self._timeout

    @property
    def interruptible(self):
        """
        Whether or not the task is interruptible.
        :rtype: bool
        """
        return self._interruptible

    @property
    def discovery_version(self):
        """
        This is the version used to create a logical version for data in the cache.
        This is only used when `discoverable` is true.  Data is considered discoverable if: the inputs to a given
        task are the same and the discovery_version is also the same.
        :rtype: Text
        """
        return self._discovery_version

    @property
    def deprecated_error_message(self):
        """
        This string can be used to mark the task as deprecated.  Consumers of the task will receive deprecation
        warnings.
        :rtype: Text
        """
        return self._deprecated_error_message

    @property
    def cache_serializable(self):
        """
        Whether or not caching operations are executed in serial. This means only a single instance over identical
        inputs is executed, other concurrent executions wait for the cached results.
        :rtype: bool
        """
        return self._cache_serializable

    @property
    def pod_template_name(self):
        """
        The name of the existing PodTemplate resource which will be used in this task.
        :rtype: Text
        """
        return self._pod_template_name

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.task_pb2.TaskMetadata
        """
        tm = _core_task.TaskMetadata(
            discoverable=self.discoverable,
            runtime=self.runtime.to_flyte_idl(),
            retries=self.retries.to_flyte_idl(),
            interruptible=self.interruptible,
            discovery_version=self.discovery_version,
            deprecated_error_message=self.deprecated_error_message,
            cache_serializable=self.cache_serializable,
            pod_template_name=self.pod_template_name,
        )
        if self.timeout:
            tm.timeout.FromTimedelta(self.timeout)
        return tm

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.task_pb2.TaskMetadata pb2_object:
        :rtype: TaskMetadata
        """
        return cls(
            discoverable=pb2_object.discoverable,
            runtime=RuntimeMetadata.from_flyte_idl(pb2_object.runtime),
            timeout=pb2_object.timeout.ToTimedelta(),
            interruptible=pb2_object.interruptible if pb2_object.HasField("interruptible") else None,
            retries=_literals.RetryStrategy.from_flyte_idl(pb2_object.retries),
            discovery_version=pb2_object.discovery_version,
            deprecated_error_message=pb2_object.deprecated_error_message,
            cache_serializable=pb2_object.cache_serializable,
            pod_template_name=pb2_object.pod_template_name,
        )


class TaskTemplate(_common.FlyteIdlEntity):
    def __init__(
        self,
        id,
        type,
        metadata,
        interface,
        custom,
        container=None,
        task_type_version=0,
        security_context=None,
        config=None,
        k8s_pod=None,
        sql=None,
        extended_resources=None,
    ):
        """
        A task template represents the full set of information necessary to perform a unit of work in the Flyte system.
        It contains the metadata about what inputs and outputs are consumed or produced.  It also contains the metadata
        necessary for Flyte Propeller to do the appropriate work.

        :param flytekit.models.core.identifier.Identifier id: This is generated by the system and uniquely identifies
            the task.
        :param Text type: This is used to define additional extensions for use by Propeller or SDK.
        :param TaskMetadata metadata: This contains information needed at runtime to determine behavior such as
            whether or not outputs are discoverable, timeouts, and retries.
        :param flytekit.models.interface.TypedInterface interface: The interface definition for this task.
        :param dict[Text, T] custom: Dictionary that must be serializable to a protobuf Struct for custom task plugins.
        :param Container container: Provides the necessary entrypoint information for execution.  For instance,
            a Container might be specified with the necessary command line arguments.
        :param int task_type_version: Specific version of this task type used by plugins to potentially modify
            execution behavior or serialization.
        :param dict[str, str] config: For plugin tasks this represents additional configuration information to be used
            in tandem with the custom.
        :param dict[str, str] config: For plugin tasks this represents additional configuration information to be used
            in tandem with the custom.
        :param K8sPod k8s_pod: Alternative to the container used to execute this task.
        :param Sql sql: This is used to execute query in FlytePropeller instead of running container or k8s_pod.
        :param flyteidl.core.tasks_pb2.ExtendedResources extended_resources: The extended resources to allocate to the task.
        """
        if (
            (container is not None and k8s_pod is not None)
            or (container is not None and sql is not None)
            or (k8s_pod is not None and sql is not None)
        ):
            raise ValueError("At most one of container, k8s_pod or sql can be set")
        self._id = id
        self._type = type
        self._metadata = metadata
        self._interface = interface
        self._custom = custom
        self._container = container
        self._task_type_version = task_type_version
        self._config = config
        self._security_context = security_context
        self._k8s_pod = k8s_pod
        self._sql = sql
        self._extended_resources = extended_resources

    @property
    def id(self):
        """
        This is generated by the system and uniquely identifies the task.
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._id

    @property
    def type(self):
        """
        This is used to identify additional extensions for use by Propeller or SDK.
        :rtype: Text
        """
        return self._type

    @property
    def metadata(self):
        """
        This contains information needed at runtime to determine behavior such as whether or not outputs are
        discoverable, timeouts, and retries.
        :rtype: TaskMetadata
        """
        return self._metadata

    @property
    def interface(self):
        """
        The interface definition for this task.
        :rtype: flytekit.models.interface.TypedInterface
        """
        return self._interface

    @property
    def custom(self):
        """
        Arbitrary dictionary containing metadata for custom plugins.
        :rtype: dict[Text, T]
        """
        return self._custom

    @property
    def task_type_version(self):
        return self._task_type_version

    @property
    def container(self):
        """
        If not None, the target of execution should be a container.
        :rtype: Container
        """
        return self._container

    @property
    def config(self):
        """
        Arbitrary dictionary containing metadata for parsing and handling custom plugins.
        :rtype: dict[Text, T]
        """
        return self._config

    @property
    def security_context(self):
        return self._security_context

    @property
    def k8s_pod(self):
        return self._k8s_pod

    @property
    def sql(self):
        return self._sql

    @property
    def extended_resources(self):
        """
        If not None, the extended resources to allocate to the task.
        :rtype: flyteidl.core.tasks_pb2.ExtendedResources
        """
        return self._extended_resources

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.tasks_pb2.TaskTemplate
        """
        task_template = _core_task.TaskTemplate(
            id=self.id.to_flyte_idl(),
            type=self.type,
            metadata=self.metadata.to_flyte_idl(),
            interface=self.interface.to_flyte_idl(),
            custom=_json_format.Parse(_json.dumps(self.custom), _struct.Struct()) if self.custom else None,
            container=self.container.to_flyte_idl() if self.container else None,
            task_type_version=self.task_type_version,
            security_context=self.security_context.to_flyte_idl() if self.security_context else None,
            extended_resources=self.extended_resources,
            config={k: v for k, v in self.config.items()} if self.config is not None else None,
            k8s_pod=self.k8s_pod.to_flyte_idl() if self.k8s_pod else None,
            sql=self.sql.to_flyte_idl() if self.sql else None,
        )
        return task_template

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.tasks_pb2.TaskTemplate pb2_object:
        :rtype: TaskTemplate
        """
        return cls(
            id=_identifier.Identifier.from_flyte_idl(pb2_object.id),
            type=pb2_object.type,
            metadata=TaskMetadata.from_flyte_idl(pb2_object.metadata),
            interface=_interface.TypedInterface.from_flyte_idl(pb2_object.interface),
            custom=_json_format.MessageToDict(pb2_object.custom) if pb2_object else None,
            container=Container.from_flyte_idl(pb2_object.container) if pb2_object.HasField("container") else None,
            task_type_version=pb2_object.task_type_version,
            security_context=_sec.SecurityContext.from_flyte_idl(pb2_object.security_context)
            if pb2_object.security_context and pb2_object.security_context.ByteSize() > 0
            else None,
            extended_resources=pb2_object.extended_resources if pb2_object.HasField("extended_resources") else None,
            config={k: v for k, v in pb2_object.config.items()} if pb2_object.config is not None else None,
            k8s_pod=K8sPod.from_flyte_idl(pb2_object.k8s_pod) if pb2_object.HasField("k8s_pod") else None,
            sql=Sql.from_flyte_idl(pb2_object.sql) if pb2_object.HasField("sql") else None,
        )


class TaskSpec(_common.FlyteIdlEntity):
    def __init__(self, template: TaskTemplate, docs: typing.Optional[Documentation] = None):
        """
        :param TaskTemplate template:
        :param Documentation docs:
        """
        self._template = template
        self._docs = docs

    @property
    def template(self):
        """
        :rtype: TaskTemplate
        """
        return self._template

    @property
    def docs(self):
        """
        :rtype: Description entity for the task
        """
        return self._docs

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.tasks_pb2.TaskSpec
        """
        return _admin_task.TaskSpec(
            template=self.template.to_flyte_idl(), description=self.docs.to_flyte_idl() if self.docs else None
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.tasks_pb2.TaskSpec pb2_object:
        :rtype: TaskSpec
        """
        return cls(
            TaskTemplate.from_flyte_idl(pb2_object.template),
            Documentation.from_flyte_idl(pb2_object.description) if pb2_object.description else None,
        )


class Task(_common.FlyteIdlEntity):
    def __init__(self, id, closure):
        """
        :param flytekit.models.core.identifier.Identifier id: The (project, domain, name) identifier for this task.
        :param TaskClosure closure: The closure for the underlying workload.
        """
        self._id = id
        self._closure = closure

    @property
    def id(self):
        """
        The (project, domain, name, version) identifier for this task.
        :rtype: flytekit.models.core.identifier.Identifier
        """
        return self._id

    @property
    def closure(self):
        """
        The closure for the underlying workload.
        :rtype: TaskClosure
        """
        return self._closure

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.task_pb2.Task
        """
        return _admin_task.Task(
            closure=self.closure.to_flyte_idl(),
            id=self.id.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.task_pb2.Task pb2_object:
        :rtype: TaskDefinition
        """
        return cls(
            closure=TaskClosure.from_flyte_idl(pb2_object.closure),
            id=_identifier.Identifier.from_flyte_idl(pb2_object.id),
        )


class TaskClosure(_common.FlyteIdlEntity):
    def __init__(self, compiled_task):
        """
        :param CompiledTask compiled_task:
        """
        self._compiled_task = compiled_task

    @property
    def compiled_task(self):
        """
        :rtype: CompiledTask
        """
        return self._compiled_task

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.task_pb2.TaskClosure
        """
        return _admin_task.TaskClosure(compiled_task=self.compiled_task.to_flyte_idl())

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.task_pb2.TaskClosure pb2_object:
        :rtype: TaskClosure
        """
        return cls(compiled_task=CompiledTask.from_flyte_idl(pb2_object.compiled_task))


class CompiledTask(_common.FlyteIdlEntity):
    def __init__(self, template):
        """
        :param TaskTemplate template:
        """
        self._template = template

    @property
    def template(self):
        """
        :rtype: TaskTemplate
        """
        return self._template

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.compiler_pb2.CompiledTask
        """
        return _compiler.CompiledTask(template=self.template.to_flyte_idl())

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.core.compiler_pb2.CompiledTask pb2_object:
        :rtype: CompiledTask
        """
        return cls(template=TaskTemplate.from_flyte_idl(pb2_object.template))


class IOStrategy(_common.FlyteIdlEntity):
    """
    Provides methods to manage data in and out of the Raw container using Download Modes. This can only be used if DataLoadingConfig is enabled.
    """

    DOWNLOAD_MODE_EAGER = _core_task.IOStrategy.DOWNLOAD_EAGER
    DOWNLOAD_MODE_STREAM = _core_task.IOStrategy.DOWNLOAD_STREAM
    DOWNLOAD_MODE_NO_DOWNLOAD = _core_task.IOStrategy.DO_NOT_DOWNLOAD

    UPLOAD_MODE_EAGER = _core_task.IOStrategy.UPLOAD_EAGER
    UPLOAD_MODE_ON_EXIT = _core_task.IOStrategy.UPLOAD_ON_EXIT
    UPLOAD_MODE_NO_UPLOAD = _core_task.IOStrategy.DO_NOT_UPLOAD

    def __init__(
        self,
        download_mode: _core_task.IOStrategy.DownloadMode = DOWNLOAD_MODE_EAGER,
        upload_mode: _core_task.IOStrategy.UploadMode = UPLOAD_MODE_ON_EXIT,
    ):
        self._download_mode = download_mode
        self._upload_mode = upload_mode

    def to_flyte_idl(self) -> _core_task.IOStrategy:
        return _core_task.IOStrategy(download_mode=self._download_mode, upload_mode=self._upload_mode)

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_task.IOStrategy):
        if pb2_object is None:
            return None
        return cls(
            download_mode=pb2_object.download_mode,
            upload_mode=pb2_object.upload_mode,
        )


class DataLoadingConfig(_common.FlyteIdlEntity):
    LITERALMAP_FORMAT_PROTO = _core_task.DataLoadingConfig.PROTO
    LITERALMAP_FORMAT_JSON = _core_task.DataLoadingConfig.JSON
    LITERALMAP_FORMAT_YAML = _core_task.DataLoadingConfig.YAML
    _LITERALMAP_FORMATS = frozenset([LITERALMAP_FORMAT_JSON, LITERALMAP_FORMAT_PROTO, LITERALMAP_FORMAT_YAML])

    def __init__(
        self,
        input_path: str,
        output_path: str,
        enabled: bool = True,
        format: _core_task.DataLoadingConfig.LiteralMapFormat = LITERALMAP_FORMAT_PROTO,
        io_strategy: IOStrategy = None,
    ):
        if format not in self._LITERALMAP_FORMATS:
            raise ValueError(
                "Metadata format {} not supported. Should be one of {}".format(format, self._LITERALMAP_FORMATS)
            )
        self._input_path = input_path
        self._output_path = output_path
        self._enabled = enabled
        self._format = format
        self._io_strategy = io_strategy

    def to_flyte_idl(self) -> _core_task.DataLoadingConfig:
        return _core_task.DataLoadingConfig(
            input_path=self._input_path,
            output_path=self._output_path,
            format=self._format,
            enabled=self._enabled,
            io_strategy=self._io_strategy.to_flyte_idl() if self._io_strategy is not None else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2: _core_task.DataLoadingConfig) -> "DataLoadingConfig":
        if pb2 is None:
            return None
        return cls(
            input_path=pb2.input_path,
            output_path=pb2.output_path,
            enabled=pb2.enabled,
            format=pb2.format,
            io_strategy=IOStrategy.from_flyte_idl(pb2.io_strategy) if pb2.HasField("io_strategy") else None,
        )


class Container(_common.FlyteIdlEntity):
    def __init__(self, image, command, args, resources, env, config, data_loading_config=None):
        """
        This defines a container target.  It will execute the appropriate command line on the appropriate image with
        the given configurations.

        :param Text image: The fully-qualified identifier for the image.
        :param list[Text] command: A list of 'words' for the command.  i.e. ['aws', 's3', 'ls']
        :param list[Text] args: A list of arguments for the command.  i.e. ['s3://some/path', '/tmp/local/path']
        :param Resources resources: A definition of requisite compute resources.
        :param dict[Text, Text] env: A definition of key-value pairs for environment variables.
        :param dict[Text, Text] config: A definition of configuration key-value pairs.
        :type DataLoadingConfig data_loading_config: object
        """
        self._data_loading_config = data_loading_config
        self._image = image
        self._command = command
        self._args = args
        self._resources = resources
        self._env = env
        self._config = config

    @property
    def image(self):
        """
        The fully-qualified identifier for the image.
        :rtype: Text
        """
        return self._image

    @property
    def command(self):
        """
        A list of 'words' for the command.  i.e. ['aws', 's3', 'ls']
        :rtype: list[Text]
        """
        return self._command

    @property
    def args(self):
        """
         A list of arguments for the command.  i.e. ['s3://some/path', '/tmp/local/path']
        :rtype: list[Text]
        """
        return self._args

    @property
    def resources(self):
        """
        A definition of requisite compute resources.
        :rtype: Resources
        """
        return self._resources

    @property
    def env(self):
        """
        A definition of key-value pairs for environment variables.  Currently, only str->str is
            supported.
        :rtype: dict[Text, Text]
        """
        return self._env

    def add_env(self, key: str, val: str):
        self._env[key] = val

    @property
    def config(self):
        """
        A definition of key-value pairs for configuration.  Currently, only str->str is
            supported.
        :rtype: dict[Text, Text]
        """
        return self._config

    @property
    def data_loading_config(self):
        """
        :rtype: DataLoadingConfig
        """
        return self._data_loading_config

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.core.tasks_pb2.Container
        """
        return _core_task.Container(
            image=self.image,
            command=self.command,
            args=self.args,
            resources=self.resources.to_flyte_idl(),
            env=[_literals_pb2.KeyValuePair(key=k, value=v) for k, v in self.env.items()],
            config=[_literals_pb2.KeyValuePair(key=k, value=v) for k, v in self.config.items()],
            data_config=self._data_loading_config.to_flyte_idl() if self._data_loading_config else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.task_pb2.Task pb2_object:
        :rtype: Container
        """
        return cls(
            image=pb2_object.image,
            command=pb2_object.command,
            args=pb2_object.args,
            resources=Resources.from_flyte_idl(pb2_object.resources),
            env={kv.key: kv.value for kv in pb2_object.env},
            config={kv.key: kv.value for kv in pb2_object.config},
            data_loading_config=DataLoadingConfig.from_flyte_idl(pb2_object.data_config)
            if pb2_object.HasField("data_config")
            else None,
        )


class K8sObjectMetadata(_common.FlyteIdlEntity):
    def __init__(self, labels: typing.Dict[str, str] = None, annotations: typing.Dict[str, str] = None):
        """
        This defines additional metadata for building a kubernetes pod.
        """
        self._labels = labels
        self._annotations = annotations

    @property
    def labels(self) -> typing.Dict[str, str]:
        return self._labels

    @property
    def annotations(self) -> typing.Dict[str, str]:
        return self._annotations

    def to_flyte_idl(self) -> _core_task.K8sObjectMetadata:
        return _core_task.K8sObjectMetadata(
            labels={k: v for k, v in self.labels.items()} if self.labels is not None else None,
            annotations={k: v for k, v in self.annotations.items()} if self.annotations is not None else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_task.K8sObjectMetadata):
        return cls(
            labels={k: v for k, v in pb2_object.labels.items()} if pb2_object.labels is not None else None,
            annotations={k: v for k, v in pb2_object.annotations.items()}
            if pb2_object.annotations is not None
            else None,
        )


class K8sPod(_common.FlyteIdlEntity):
    def __init__(
        self,
        metadata: K8sObjectMetadata = None,
        pod_spec: typing.Dict[str, typing.Any] = None,
        data_config: typing.Optional[DataLoadingConfig] = None,
    ):
        """
        This defines a kubernetes pod target.  It will build the pod target during task execution
        """
        self._metadata = metadata
        self._pod_spec = pod_spec
        self._data_config = data_config

    @property
    def metadata(self) -> K8sObjectMetadata:
        return self._metadata

    @property
    def pod_spec(self) -> typing.Dict[str, typing.Any]:
        return self._pod_spec

    @property
    def data_config(self) -> typing.Optional[DataLoadingConfig]:
        return self._data_config

    def to_flyte_idl(self) -> _core_task.K8sPod:
        return _core_task.K8sPod(
            metadata=self._metadata.to_flyte_idl(),
            pod_spec=_json_format.Parse(_json.dumps(self.pod_spec), _struct.Struct()) if self.pod_spec else None,
            data_config=self.data_config.to_flyte_idl() if self.data_config else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_task.K8sPod):
        return cls(
            metadata=K8sObjectMetadata.from_flyte_idl(pb2_object.metadata),
            pod_spec=_json_format.MessageToDict(pb2_object.pod_spec) if pb2_object.HasField("pod_spec") else None,
            data_config=DataLoadingConfig.from_flyte_idl(pb2_object.data_config)
            if pb2_object.HasField("data_config")
            else None,
        )


class Sql(_common.FlyteIdlEntity):
    class Dialect(object):
        ANSI = 0
        HIVE = 1

    def __init__(self, statement: str = None, dialect: int = 0):
        """
        This defines a kubernetes pod target. It will build the pod target during task execution
        """
        self._statement = statement
        self._dialect = dialect

    @property
    def statement(self) -> str:
        return self._statement

    @property
    def dialect(self) -> int:
        return self._dialect

    def to_flyte_idl(self) -> _core_task.Sql:
        return _core_task.Sql(statement=self.statement, dialect=self.dialect)

    @classmethod
    def from_flyte_idl(cls, pb2_object: _core_task.Sql):
        return cls(
            statement=pb2_object.statement,
            dialect=pb2_object.dialect,
        )
