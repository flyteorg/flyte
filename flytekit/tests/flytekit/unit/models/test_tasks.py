from datetime import timedelta
from itertools import product

import pytest
from flyteidl.core.tasks_pb2 import ExtendedResources, TaskMetadata
from google.protobuf import text_format

import flytekit.models.interface as interface_models
import flytekit.models.literals as literal_models
from flytekit import Description, Documentation, SourceCode
from flytekit.extras.accelerators import T4
from flytekit.models import literals, task, types
from flytekit.models.core import identifier
from tests.flytekit.common import parameterizers


def test_resource_entry():
    obj = task.Resources.ResourceEntry(task.Resources.ResourceName.CPU, "blah")
    assert task.Resources.ResourceEntry.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert obj != task.Resources.ResourceEntry(task.Resources.ResourceName.GPU, "blah")
    assert obj != task.Resources.ResourceEntry(task.Resources.ResourceName.CPU, "bloop")
    assert obj.name == task.Resources.ResourceName.CPU
    assert obj.value == "blah"


@pytest.mark.parametrize("resource_list", parameterizers.LIST_OF_RESOURCE_ENTRY_LISTS)
def test_resources(resource_list):
    obj = task.Resources(resource_list, resource_list)
    obj1 = task.Resources([], resource_list)
    obj2 = task.Resources(resource_list, [])

    assert obj.requests == obj2.requests
    assert obj.limits == obj1.limits
    assert obj == task.Resources.from_flyte_idl(obj.to_flyte_idl())


def test_runtime_metadata():
    obj = task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0", "python")
    assert obj.type == task.RuntimeMetadata.RuntimeType.FLYTE_SDK
    assert obj.version == "1.0.0"
    assert obj.flavor == "python"
    assert obj == task.RuntimeMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj != task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.1", "python")
    assert obj != task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.OTHER, "1.0.0", "python")
    assert obj != task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0", "golang")


def test_task_metadata_interruptible_from_flyte_idl():
    # Interruptible not set
    idl = TaskMetadata()
    obj = task.TaskMetadata.from_flyte_idl(idl)
    assert obj.interruptible is None

    idl = TaskMetadata()
    idl.interruptible = True
    obj = task.TaskMetadata.from_flyte_idl(idl)
    assert obj.interruptible is True

    idl = TaskMetadata()
    idl.interruptible = False
    obj = task.TaskMetadata.from_flyte_idl(idl)
    assert obj.interruptible is False


def test_task_metadata():
    obj = task.TaskMetadata(
        True,
        task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0", "python"),
        timedelta(days=1),
        literals.RetryStrategy(3),
        True,
        "0.1.1b0",
        "This is deprecated!",
        True,
        "A",
    )

    assert obj.discoverable is True
    assert obj.retries.retries == 3
    assert obj.interruptible is True
    assert obj.timeout == timedelta(days=1)
    assert obj.runtime.flavor == "python"
    assert obj.runtime.type == task.RuntimeMetadata.RuntimeType.FLYTE_SDK
    assert obj.runtime.version == "1.0.0"
    assert obj.deprecated_error_message == "This is deprecated!"
    assert obj.discovery_version == "0.1.1b0"
    assert obj.pod_template_name == "A"
    assert obj == task.TaskMetadata.from_flyte_idl(obj.to_flyte_idl())


@pytest.mark.parametrize(
    "in_tuple",
    product(parameterizers.LIST_OF_TASK_METADATA, parameterizers.LIST_OF_INTERFACES, parameterizers.LIST_OF_RESOURCES),
)
def test_task_template(in_tuple):
    task_metadata, interfaces, resources = in_tuple
    obj = task.TaskTemplate(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        task_metadata,
        interfaces,
        {"a": 1, "b": {"c": 2, "d": 3}},
        container=task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            resources,
            {"a": "b"},
            {"d": "e"},
        ),
        config={"a": "b"},
        extended_resources=ExtendedResources(gpu_accelerator=T4.to_flyte_idl()),
    )
    assert obj.id.resource_type == identifier.ResourceType.TASK
    assert obj.id.project == "project"
    assert obj.id.domain == "domain"
    assert obj.id.name == "name"
    assert obj.id.version == "version"
    assert obj.type == "python"
    assert obj.metadata == task_metadata
    assert obj.interface == interfaces
    assert obj.custom == {"a": 1, "b": {"c": 2, "d": 3}}
    assert obj.container.image == "my_image"
    assert obj.container.resources == resources
    assert text_format.MessageToString(obj.to_flyte_idl()) == text_format.MessageToString(
        task.TaskTemplate.from_flyte_idl(obj.to_flyte_idl()).to_flyte_idl()
    )
    assert obj.config == {"a": "b"}
    assert obj.extended_resources.gpu_accelerator.device == "nvidia-tesla-t4"
    assert not obj.extended_resources.gpu_accelerator.HasField("unpartitioned")
    assert not obj.extended_resources.gpu_accelerator.HasField("partition_size")


def test_task_spec():
    task_metadata = task.TaskMetadata(
        True,
        task.RuntimeMetadata(task.RuntimeMetadata.RuntimeType.FLYTE_SDK, "1.0.0", "python"),
        timedelta(days=1),
        literals.RetryStrategy(3),
        True,
        "0.1.1b0",
        "This is deprecated!",
        True,
        "A",
    )

    int_type = types.LiteralType(types.SimpleType.INTEGER)
    interfaces = interface_models.TypedInterface(
        {"a": interface_models.Variable(int_type, "description1")},
        {
            "b": interface_models.Variable(int_type, "description2"),
            "c": interface_models.Variable(int_type, "description3"),
        },
    )

    resource = [task.Resources.ResourceEntry(task.Resources.ResourceName.CPU, "1")]
    resources = task.Resources(resource, resource)

    template = task.TaskTemplate(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        task_metadata,
        interfaces,
        {"a": 1, "b": {"c": 2, "d": 3}},
        container=task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            resources,
            {"a": "b"},
            {"d": "e"},
        ),
        config={"a": "b"},
        extended_resources=ExtendedResources(gpu_accelerator=T4.to_flyte_idl()),
    )

    short_description = "short"
    long_description = Description(value="long", icon_link="http://icon")
    source_code = SourceCode(link="https://github.com/flyteorg/flytekit")
    docs = Documentation(
        short_description=short_description, long_description=long_description, source_code=source_code
    )

    obj = task.TaskSpec(template, docs)
    assert task.TaskSpec.from_flyte_idl(obj.to_flyte_idl()) == obj
    assert obj.docs == docs
    assert obj.template == template


def test_task_template_k8s_pod_target():
    int_type = types.LiteralType(types.SimpleType.INTEGER)
    obj = task.TaskTemplate(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        task.TaskMetadata(
            False,
            task.RuntimeMetadata(1, "v", "f"),
            timedelta(days=1),
            literal_models.RetryStrategy(5),
            False,
            "1.0",
            "deprecated",
            False,
            "A",
        ),
        interface_models.TypedInterface(
            # inputs
            {"a": interface_models.Variable(int_type, "description1")},
            # outputs
            {
                "b": interface_models.Variable(int_type, "description2"),
                "c": interface_models.Variable(int_type, "description3"),
            },
        ),
        {"a": 1, "b": {"c": 2, "d": 3}},
        config={"a": "b"},
        k8s_pod=task.K8sPod(
            metadata=task.K8sObjectMetadata(labels={"label": "foo"}, annotations={"anno": "bar"}),
            pod_spec={"str": "val", "int": 1},
        ),
        extended_resources=ExtendedResources(gpu_accelerator=T4.to_flyte_idl()),
    )
    assert obj.id.resource_type == identifier.ResourceType.TASK
    assert obj.id.project == "project"
    assert obj.id.domain == "domain"
    assert obj.id.name == "name"
    assert obj.id.version == "version"
    assert obj.type == "python"
    assert obj.custom == {"a": 1, "b": {"c": 2, "d": 3}}
    assert obj.k8s_pod.metadata == task.K8sObjectMetadata(labels={"label": "foo"}, annotations={"anno": "bar"})
    assert obj.k8s_pod.pod_spec == {"str": "val", "int": 1}
    assert text_format.MessageToString(obj.to_flyte_idl()) == text_format.MessageToString(
        task.TaskTemplate.from_flyte_idl(obj.to_flyte_idl()).to_flyte_idl()
    )
    assert obj.config == {"a": "b"}
    assert obj.extended_resources.gpu_accelerator.device == "nvidia-tesla-t4"
    assert not obj.extended_resources.gpu_accelerator.HasField("unpartitioned")
    assert not obj.extended_resources.gpu_accelerator.HasField("partition_size")


@pytest.mark.parametrize("sec_ctx", parameterizers.LIST_OF_SECURITY_CONTEXT)
def test_task_template_security_context(sec_ctx):
    obj = task.TaskTemplate(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        parameterizers.LIST_OF_TASK_METADATA[0],
        parameterizers.LIST_OF_INTERFACES[0],
        {"a": 1, "b": {"c": 2, "d": 3}},
        container=task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            parameterizers.LIST_OF_RESOURCES[0],
            {"a": "b"},
            {"d": "e"},
        ),
        security_context=sec_ctx,
    )
    assert obj.security_context == sec_ctx
    expected = obj.security_context
    if sec_ctx:
        if sec_ctx.run_as is None and sec_ctx.secrets is None and sec_ctx.tokens is None:
            expected = None
    assert task.TaskTemplate.from_flyte_idl(obj.to_flyte_idl()).security_context == expected


@pytest.mark.parametrize("extended_resources", parameterizers.LIST_OF_EXTENDED_RESOURCES)
def test_task_template_extended_resources(extended_resources):
    obj = task.TaskTemplate(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        "python",
        parameterizers.LIST_OF_TASK_METADATA[0],
        parameterizers.LIST_OF_INTERFACES[0],
        {"a": 1, "b": {"c": 2, "d": 3}},
        container=task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            parameterizers.LIST_OF_RESOURCES[0],
            {"a": "b"},
            {"d": "e"},
        ),
        extended_resources=extended_resources,
    )
    assert obj.extended_resources == extended_resources
    assert task.TaskTemplate.from_flyte_idl(obj.to_flyte_idl()).extended_resources == extended_resources


@pytest.mark.parametrize("task_closure", parameterizers.LIST_OF_TASK_CLOSURES)
def test_task(task_closure):
    obj = task.Task(
        identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version"),
        task_closure,
    )
    assert obj.id.project == "project"
    assert obj.id.domain == "domain"
    assert obj.id.name == "name"
    assert obj.id.version == "version"
    assert obj.closure == task_closure
    assert obj == task.Task.from_flyte_idl(obj.to_flyte_idl())


@pytest.mark.parametrize("resources", parameterizers.LIST_OF_RESOURCES)
def test_container(resources):
    obj = task.Container(
        "my_image",
        ["this", "is", "a", "cmd"],
        ["this", "is", "an", "arg"],
        resources,
        {"a": "b"},
        {"d": "e"},
    )
    obj.image == "my_image"
    obj.command == ["this", "is", "a", "cmd"]
    obj.args == ["this", "is", "an", "arg"]
    obj.resources == resources
    obj.env == {"a": "b"}
    obj.config == {"d": "e"}
    assert obj == task.Container.from_flyte_idl(obj.to_flyte_idl())


def test_dataloadingconfig():
    dlc = task.DataLoadingConfig(
        "s3://input/path",
        "s3://output/path",
        True,
        task.DataLoadingConfig.LITERALMAP_FORMAT_YAML,
    )
    dlc2 = task.DataLoadingConfig.from_flyte_idl(dlc.to_flyte_idl())
    assert dlc2 == dlc

    dlc = task.DataLoadingConfig(
        "s3://input/path",
        "s3://output/path",
        True,
        task.DataLoadingConfig.LITERALMAP_FORMAT_YAML,
        io_strategy=task.IOStrategy(),
    )
    dlc2 = task.DataLoadingConfig.from_flyte_idl(dlc.to_flyte_idl())
    assert dlc2 == dlc


def test_ioconfig():
    io = task.IOStrategy(task.IOStrategy.DOWNLOAD_MODE_NO_DOWNLOAD, task.IOStrategy.UPLOAD_MODE_NO_UPLOAD)
    assert io == task.IOStrategy.from_flyte_idl(io.to_flyte_idl())


def test_k8s_metadata():
    obj = task.K8sObjectMetadata(labels={"label": "foo"}, annotations={"anno": "bar"})
    assert obj.labels == {"label": "foo"}
    assert obj.annotations == {"anno": "bar"}
    assert obj == task.K8sObjectMetadata.from_flyte_idl(obj.to_flyte_idl())


def test_k8s_pod():
    obj = task.K8sPod(metadata=task.K8sObjectMetadata(labels={"label": "foo"}), pod_spec={"pod_spec": "bar"})
    assert obj.metadata.labels == {"label": "foo"}
    assert obj.pod_spec == {"pod_spec": "bar"}
    assert obj == task.K8sPod.from_flyte_idl(obj.to_flyte_idl())
