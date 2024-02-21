import asyncio
import subprocess
from collections import OrderedDict
from shutil import which
from unittest.mock import MagicMock

import grpc
import pytest
from flyteidl.core.execution_pb2 import TaskExecution
from flytekitplugins.mmcloud import MMCloudAgent, MMCloudConfig, MMCloudTask
from flytekitplugins.mmcloud.utils import async_check_output, flyte_to_mmcloud_resources

from flytekit import Resources, task
from flytekit.configuration import DefaultImages, ImageConfig, SerializationSettings
from flytekit.extend import get_serializable
from flytekit.extend.backend.base_agent import AgentRegistry

float_missing = which("float") is None


def test_mmcloud_task():
    task_config = MMCloudConfig(submit_extra="--migratePolicy [enable=true]")
    requests = Resources(cpu="2", mem="4Gi")
    limits = Resources(cpu="4")
    container_image = DefaultImages.default_image()
    environment = {"KEY": "value"}

    @task(
        task_config=task_config,
        requests=requests,
        limits=limits,
        container_image=container_image,
        environment=environment,
    )
    def say_hello(name: str) -> str:
        return f"Hello, {name}."

    assert say_hello.task_config == task_config
    assert say_hello.task_type == "mmcloud_task"
    assert isinstance(say_hello, MMCloudTask)

    serialization_settings = SerializationSettings(image_config=ImageConfig())
    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello)
    template = task_spec.template
    container = template.container

    assert template.custom == {"submit_extra": "--migratePolicy [enable=true]", "resources": ["2", "4", "4", None]}
    assert container.image == container_image
    assert container.env == environment


def test_async_check_output():
    message = "Hello, World!"
    stdout = asyncio.run(async_check_output("echo", message))
    assert stdout.decode() == f"{message}\n"

    with pytest.raises(FileNotFoundError):
        asyncio.run(async_check_output("nonexistent_command"))

    with pytest.raises(subprocess.CalledProcessError):
        asyncio.run(async_check_output("false"))


def test_flyte_to_mmcloud_resources():
    B_IN_GIB = 1073741824
    success_cases = {
        ("0", "0", "0", "0"): (1, 1, 1, 1),
        ("0", "0", None, None): (1, 1, None, None),
        (None, None, "0", "0"): (1, 1, 1, 1),
        ("1", "2Gi", "3", "4Gi"): (1, 2, 3, 4),
        ("1", "2Gi", None, None): (1, 2, None, None),
        (None, None, "3", "4Gi"): (3, 4, 3, 4),
        (None, None, None, None): (1, 1, None, None),
        ("1.1", str(B_IN_GIB + 1), "2.1", str(2 * B_IN_GIB + 1)): (2, 2, 3, 3),
    }

    for (req_cpu, req_mem, lim_cpu, lim_mem), (min_cpu, min_mem, max_cpu, max_mem) in success_cases.items():
        resources = flyte_to_mmcloud_resources(
            requests=Resources(cpu=req_cpu, mem=req_mem),
            limits=Resources(cpu=lim_cpu, mem=lim_mem),
        )
        assert resources == (min_cpu, min_mem, max_cpu, max_mem)

    error_cases = {
        ("1", "2Gi", "2", "1Gi"),
        ("2", "2Gi", "1", "1Gi"),
        ("2", "1Gi", "1", "2Gi"),
    }
    for req_cpu, req_mem, lim_cpu, lim_mem in error_cases:
        with pytest.raises(ValueError):
            flyte_to_mmcloud_resources(
                requests=Resources(cpu=req_cpu, mem=req_mem),
                limits=Resources(cpu=lim_cpu, mem=lim_mem),
            )


@pytest.mark.skipif(float_missing, reason="float binary is required")
def test_async_agent():
    serialization_settings = SerializationSettings(image_config=ImageConfig())
    context = MagicMock(spec=grpc.ServicerContext)
    container_image = DefaultImages.default_image()

    @task(
        task_config=MMCloudConfig(submit_extra="--migratePolicy [enable=true]"),
        requests=Resources(cpu="2", mem="1Gi"),
        limits=Resources(cpu="4", mem="16Gi"),
        container_image=container_image,
    )
    def say_hello0(name: str) -> str:
        return f"Hello, {name}."

    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello0)
    agent = AgentRegistry.get_agent(task_spec.template.type)

    assert isinstance(agent, MMCloudAgent)

    create_task_response = asyncio.run(
        agent.async_create(
            context=context,
            output_prefix="",
            task_template=task_spec.template,
            inputs=None,
        )
    )
    resource_meta = create_task_response.resource_meta

    get_task_response = asyncio.run(agent.async_get(context=context, resource_meta=resource_meta))
    phase = get_task_response.resource.phase
    assert phase in (TaskExecution.RUNNING, TaskExecution.SUCCEEDED)

    asyncio.run(agent.async_delete(context=context, resource_meta=resource_meta))

    get_task_response = asyncio.run(agent.async_get(context=context, resource_meta=resource_meta))
    phase = get_task_response.resource.phase
    assert phase == TaskExecution.FAILED

    @task(
        task_config=MMCloudConfig(submit_extra="--nonexistent"),
        requests=Resources(cpu="2", mem="4Gi"),
        limits=Resources(cpu="4", mem="16Gi"),
        container_image=container_image,
    )
    def say_hello1(name: str) -> str:
        return f"Hello, {name}."

    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello1)
    with pytest.raises(subprocess.CalledProcessError):
        create_task_response = asyncio.run(
            agent.async_create(
                context=context,
                output_prefix="",
                task_template=task_spec.template,
                inputs=None,
            )
        )

    @task(
        task_config=MMCloudConfig(),
        limits=Resources(cpu="3", mem="1Gi"),
        container_image=container_image,
    )
    def say_hello2(name: str) -> str:
        return f"Hello, {name}."

    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello2)
    with pytest.raises(subprocess.CalledProcessError):
        create_task_response = asyncio.run(
            agent.async_create(
                context=context,
                output_prefix="",
                task_template=task_spec.template,
                inputs=None,
            )
        )

    @task(
        task_config=MMCloudConfig(),
        limits=Resources(cpu="2", mem="1Gi"),
        container_image=container_image,
    )
    def say_hello3(name: str) -> str:
        return f"Hello, {name}."

    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello3)
    create_task_response = asyncio.run(
        agent.async_create(
            context=context,
            output_prefix="",
            task_template=task_spec.template,
            inputs=None,
        )
    )
    resource_meta = create_task_response.resource_meta
    asyncio.run(agent.async_delete(context=context, resource_meta=resource_meta))

    @task(
        task_config=MMCloudConfig(),
        requests=Resources(cpu="2", mem="1Gi"),
        container_image=container_image,
    )
    def say_hello4(name: str) -> str:
        return f"Hello, {name}."

    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello4)
    create_task_response = asyncio.run(
        agent.async_create(
            context=context,
            output_prefix="",
            task_template=task_spec.template,
            inputs=None,
        )
    )
    resource_meta = create_task_response.resource_meta
    asyncio.run(agent.async_delete(context=context, resource_meta=resource_meta))

    @task(
        task_config=MMCloudConfig(),
        container_image=container_image,
    )
    def say_hello5(name: str) -> str:
        return f"Hello, {name}."

    task_spec = get_serializable(OrderedDict(), serialization_settings, say_hello5)
    create_task_response = asyncio.run(
        agent.async_create(
            context=context,
            output_prefix="",
            task_template=task_spec.template,
            inputs=None,
        )
    )
    resource_meta = create_task_response.resource_meta
    asyncio.run(agent.async_delete(context=context, resource_meta=resource_meta))
