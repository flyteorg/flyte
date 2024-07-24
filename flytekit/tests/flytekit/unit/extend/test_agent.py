import asyncio
import json
import typing
from collections import OrderedDict
from dataclasses import asdict, dataclass
from unittest.mock import MagicMock, patch

import grpc
import pytest
from flyteidl.admin.agent_pb2 import (
    CreateTaskRequest,
    CreateTaskResponse,
    DeleteTaskRequest,
    DeleteTaskResponse,
    GetTaskRequest,
    GetTaskResponse,
    ListAgentsRequest,
    ListAgentsResponse,
    Resource,
)
from flyteidl.core.execution_pb2 import TaskExecution

from flytekit import PythonFunctionTask, task
from flytekit.configuration import FastSerializationSettings, Image, ImageConfig, SerializationSettings
from flytekit.extend.backend.agent_service import AgentMetadataService, AsyncAgentService
from flytekit.extend.backend.base_agent import (
    AgentBase,
    AgentRegistry,
    AsyncAgentExecutorMixin,
    convert_to_flyte_phase,
    get_agent_secret,
    is_terminal_phase,
    render_task_template,
)
from flytekit.models import literals
from flytekit.models.core.execution import TaskLog
from flytekit.models.literals import LiteralMap
from flytekit.models.task import TaskTemplate
from flytekit.tools.translator import get_serializable

dummy_id = "dummy_id"
loop = asyncio.get_event_loop()


@dataclass
class Metadata:
    job_id: str


class DummyAgent(AgentBase):
    name = "Dummy Agent"

    def __init__(self):
        super().__init__(task_type="dummy", asynchronous=False)

    def create(
        self, output_prefix: str, task_template: TaskTemplate, inputs: typing.Optional[LiteralMap] = None, **kwargs
    ) -> CreateTaskResponse:
        return CreateTaskResponse(resource_meta=json.dumps(asdict(Metadata(job_id=dummy_id))).encode("utf-8"))

    def get(self, resource_meta: bytes, **kwargs) -> GetTaskResponse:
        return GetTaskResponse(
            resource=Resource(phase=TaskExecution.SUCCEEDED),
            log_links=[TaskLog(name="console", uri="localhost:3000").to_flyte_idl()],
        )

    def delete(self, resource_meta: bytes, **kwargs) -> DeleteTaskResponse:
        return DeleteTaskResponse()


class AsyncDummyAgent(AgentBase):
    name = "Async Dummy Agent"

    def __init__(self):
        super().__init__(task_type="async_dummy")

    async def create(
        self,
        output_prefix: str,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
        **kwargs,
    ) -> CreateTaskResponse:
        return CreateTaskResponse(resource_meta=json.dumps(asdict(Metadata(job_id=dummy_id))).encode("utf-8"))

    async def get(self, resource_meta: bytes, **kwargs) -> GetTaskResponse:
        return GetTaskResponse(resource=Resource(phase=TaskExecution.SUCCEEDED))

    async def delete(self, resource_meta: bytes, **kwargs) -> DeleteTaskResponse:
        return DeleteTaskResponse()


class SyncDummyAgent(AgentBase):
    name = "Sync Dummy Agent"

    def __init__(self):
        super().__init__(task_type="sync_dummy")

    def create(
        self,
        output_prefix: str,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
        **kwargs,
    ) -> CreateTaskResponse:
        return CreateTaskResponse(
            resource=Resource(phase=TaskExecution.SUCCEEDED, outputs=LiteralMap({}).to_flyte_idl())
        )


def get_task_template(task_type: str) -> TaskTemplate:
    @task
    def simple_task(i: int):
        print(i)

    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="project",
        domain="domain",
        version="version",
        env={"FOO": "baz"},
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
        fast_serialization_settings=FastSerializationSettings(enabled=True),
    )
    serialized = get_serializable(OrderedDict(), serialization_settings, simple_task)
    serialized.template._type = task_type
    return serialized.template


task_inputs = literals.LiteralMap(
    {
        "a": literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=1))),
    },
)


dummy_template = get_task_template("dummy")
async_dummy_template = get_task_template("async_dummy")
sync_dummy_template = get_task_template("sync_dummy")


def test_dummy_agent():
    AgentRegistry.register(DummyAgent())
    agent = AgentRegistry.get_agent("dummy")
    metadata_bytes = json.dumps(asdict(Metadata(job_id=dummy_id))).encode("utf-8")
    assert agent.create("/tmp", dummy_template, task_inputs).resource_meta == metadata_bytes
    res = agent.get(metadata_bytes)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    assert res.log_links[0].name == "console"
    assert res.log_links[0].uri == "localhost:3000"
    assert agent.delete(metadata_bytes) == DeleteTaskResponse()

    class DummyTask(AsyncAgentExecutorMixin, PythonFunctionTask):
        def __init__(self, **kwargs):
            super().__init__(
                task_type="dummy",
                **kwargs,
            )

    t = DummyTask(task_config={}, task_function=lambda: None, container_image="dummy")
    t.execute()

    t._task_type = "non-exist-type"
    with pytest.raises(Exception, match="Cannot find agent for task type: non-exist-type."):
        t.execute()

    agent_metadata = AgentRegistry.get_agent_metadata("Dummy Agent")
    assert agent_metadata.name == "Dummy Agent"
    assert agent_metadata.supported_task_types == ["dummy"]


@pytest.mark.asyncio
async def test_async_dummy_agent():
    AgentRegistry.register(AsyncDummyAgent())
    agent = AgentRegistry.get_agent("async_dummy")
    metadata_bytes = json.dumps(asdict(Metadata(job_id=dummy_id))).encode("utf-8")
    res = await agent.create("/tmp", async_dummy_template, task_inputs)
    assert res.resource_meta == metadata_bytes
    res = await agent.get(metadata_bytes)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    res = await agent.delete(metadata_bytes)
    assert res == DeleteTaskResponse()

    agent_metadata = AgentRegistry.get_agent_metadata("Async Dummy Agent")
    assert agent_metadata.name == "Async Dummy Agent"
    assert agent_metadata.supported_task_types == ["async_dummy"]


@pytest.mark.asyncio
async def test_sync_dummy_agent():
    AgentRegistry.register(SyncDummyAgent())
    agent = AgentRegistry.get_agent("sync_dummy")
    res = agent.create("/tmp", sync_dummy_template, task_inputs)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    assert res.resource.outputs == LiteralMap({}).to_flyte_idl()

    agent_metadata = AgentRegistry.get_agent_metadata("Sync Dummy Agent")
    assert agent_metadata.name == "Sync Dummy Agent"
    assert agent_metadata.supported_task_types == ["sync_dummy"]


@pytest.mark.asyncio
async def run_agent_server():
    service = AsyncAgentService()
    ctx = MagicMock(spec=grpc.ServicerContext)
    request = CreateTaskRequest(
        inputs=task_inputs.to_flyte_idl(), output_prefix="/tmp", template=dummy_template.to_flyte_idl()
    )
    async_request = CreateTaskRequest(
        inputs=task_inputs.to_flyte_idl(), output_prefix="/tmp", template=async_dummy_template.to_flyte_idl()
    )
    sync_request = CreateTaskRequest(
        inputs=task_inputs.to_flyte_idl(), output_prefix="/tmp", template=sync_dummy_template.to_flyte_idl()
    )
    fake_agent = "fake"
    metadata_bytes = json.dumps(asdict(Metadata(job_id=dummy_id))).encode("utf-8")

    res = await service.CreateTask(request, ctx)
    assert res.resource_meta == metadata_bytes
    res = await service.GetTask(GetTaskRequest(task_type="dummy", resource_meta=metadata_bytes), ctx)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    res = await service.DeleteTask(DeleteTaskRequest(task_type="dummy", resource_meta=metadata_bytes), ctx)
    assert isinstance(res, DeleteTaskResponse)

    res = await service.CreateTask(async_request, ctx)
    assert res.resource_meta == metadata_bytes
    res = await service.GetTask(GetTaskRequest(task_type="async_dummy", resource_meta=metadata_bytes), ctx)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    res = await service.DeleteTask(DeleteTaskRequest(task_type="async_dummy", resource_meta=metadata_bytes), ctx)
    assert isinstance(res, DeleteTaskResponse)

    res = await service.CreateTask(sync_request, ctx)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    assert res.resource.outputs == LiteralMap({}).to_flyte_idl()

    res = await service.GetTask(GetTaskRequest(task_type=fake_agent, resource_meta=metadata_bytes), ctx)
    assert res is None

    metadata_service = AgentMetadataService()
    res = await metadata_service.ListAgent(ListAgentsRequest(), ctx)
    assert isinstance(res, ListAgentsResponse)


def test_agent_server():
    loop.run_in_executor(None, run_agent_server)


def test_is_terminal_phase():
    assert is_terminal_phase(TaskExecution.SUCCEEDED)
    assert is_terminal_phase(TaskExecution.ABORTED)
    assert is_terminal_phase(TaskExecution.FAILED)
    assert not is_terminal_phase(TaskExecution.RUNNING)


def test_convert_to_flyte_phase():
    assert convert_to_flyte_phase("FAILED") == TaskExecution.FAILED
    assert convert_to_flyte_phase("TIMEOUT") == TaskExecution.FAILED
    assert convert_to_flyte_phase("TIMEDOUT") == TaskExecution.FAILED
    assert convert_to_flyte_phase("CANCELED") == TaskExecution.FAILED

    assert convert_to_flyte_phase("DONE") == TaskExecution.SUCCEEDED
    assert convert_to_flyte_phase("SUCCEEDED") == TaskExecution.SUCCEEDED
    assert convert_to_flyte_phase("SUCCESS") == TaskExecution.SUCCEEDED

    assert convert_to_flyte_phase("RUNNING") == TaskExecution.RUNNING

    invalid_state = "INVALID_STATE"
    with pytest.raises(Exception, match=f"Unrecognized state: {invalid_state.lower()}"):
        convert_to_flyte_phase(invalid_state)


@patch("flytekit.current_context")
def test_get_agent_secret(mocked_context):
    mocked_context.return_value.secrets.get.return_value = "mocked token"
    assert get_agent_secret("mocked key") == "mocked token"


def test_render_task_template():
    tt = render_task_template(dummy_template, "s3://becket")
    assert tt.container.args == [
        "pyflyte-fast-execute",
        "--additional-distribution",
        "{{ .remote_package_path }}",
        "--dest-dir",
        "{{ .dest_dir }}",
        "--",
        "pyflyte-execute",
        "--inputs",
        "s3://becket/inputs.pb",
        "--output-prefix",
        "s3://becket/output",
        "--raw-output-data-prefix",
        "s3://becket/raw_output",
        "--checkpoint-path",
        "s3://becket/checkpoint_output",
        "--prev-checkpoint",
        "s3://becket/prev_checkpoint",
        "--resolver",
        "flytekit.core.python_auto_container.default_task_resolver",
        "--",
        "task-module",
        "test_agent",
        "task-name",
        "simple_task",
    ]
