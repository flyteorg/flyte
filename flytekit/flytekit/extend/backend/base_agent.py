import asyncio
import inspect
import signal
import sys
import time
import typing
from abc import ABC
from collections import OrderedDict
from functools import partial
from types import FrameType, coroutine

from flyteidl.admin.agent_pb2 import (
    Agent,
    CreateTaskResponse,
    DeleteTaskResponse,
    GetTaskResponse,
)
from flyteidl.core import literals_pb2
from flyteidl.core.execution_pb2 import TaskExecution
from flyteidl.core.tasks_pb2 import TaskTemplate
from rich.progress import Progress

import flytekit
from flytekit import FlyteContext, PythonFunctionTask, logger
from flytekit.configuration import ImageConfig, SerializationSettings
from flytekit.core import utils
from flytekit.core.base_task import PythonTask
from flytekit.core.type_engine import TypeEngine
from flytekit.exceptions.system import FlyteAgentNotFound
from flytekit.exceptions.user import FlyteUserException
from flytekit.models.literals import LiteralMap


class AgentBase(ABC):
    """
    This is the base class for all agents. It defines the interface that all agents must implement.
    The agent service will be run either locally or in a pod, and will be responsible for
    invoking agents. The propeller will communicate with the agent service
    to create tasks, get the status of tasks, and delete tasks.

    All the agents should be registered in the AgentRegistry. Agent Service
    will look up the agent based on the task type. Every task type can only have one agent.
    """

    name = "Base Agent"

    def __init__(self, task_type: str, **kwargs):
        self._task_type = task_type

    @property
    def task_type(self) -> str:
        """
        task_type is the name of the task type that this agent supports.
        """
        return self._task_type

    def create(
        self,
        output_prefix: str,
        task_template: TaskTemplate,
        inputs: typing.Optional[LiteralMap] = None,
        **kwargs,
    ) -> CreateTaskResponse:
        """
        Return a Unique ID for the task that was created. It should return error code if the task creation failed.
        """
        raise NotImplementedError

    def get(self, resource_meta: bytes, **kwargs) -> GetTaskResponse:
        """
        Return the status of the task, and return the outputs in some cases. For example, bigquery job
        can't write the structured dataset to the output location, so it returns the output literals to the propeller,
        and the propeller will write the structured dataset to the blob store.
        """
        raise NotImplementedError

    def delete(self, resource_meta: bytes, **kwargs) -> DeleteTaskResponse:
        """
        Delete the task. This call should be idempotent.
        """
        raise NotImplementedError


class AgentRegistry(object):
    """
    This is the registry for all agents.
    The agent service will look up the agent registry based on the task type.
    The agent metadata service will look up the agent metadata based on the agent name.
    """

    _REGISTRY: typing.Dict[str, AgentBase] = {}
    _METADATA: typing.Dict[str, Agent] = {}

    @staticmethod
    def register(agent: AgentBase):
        if agent.task_type in AgentRegistry._REGISTRY:
            raise ValueError(f"Duplicate agent for task type {agent.task_type}")
        AgentRegistry._REGISTRY[agent.task_type] = agent

        if agent.name in AgentRegistry._METADATA:
            agent_metadata = AgentRegistry._METADATA[agent.name]
            agent_metadata.supported_task_types.append(agent.task_type)
        else:
            agent_metadata = Agent(name=agent.name, supported_task_types=[agent.task_type])
            AgentRegistry._METADATA[agent.name] = agent_metadata

        logger.info(f"Registering an agent for task type: {agent.task_type}, name: {agent.name}")

    @staticmethod
    def get_agent(task_type: str) -> AgentBase:
        if task_type not in AgentRegistry._REGISTRY:
            raise FlyteAgentNotFound(f"Cannot find agent for task type: {task_type}.")
        return AgentRegistry._REGISTRY[task_type]

    @staticmethod
    def get_agent_metadata(name: str) -> Agent:
        if name not in AgentRegistry._METADATA:
            raise FlyteAgentNotFound(f"Cannot find agent for name: {name}.")
        return AgentRegistry._METADATA[name]


def mirror_async_methods(func: typing.Callable, **kwargs) -> typing.Coroutine:
    if inspect.iscoroutinefunction(func):
        return func(**kwargs)
    args = [v for _, v in kwargs.items()]
    return asyncio.get_running_loop().run_in_executor(None, func, *args)


def convert_to_flyte_phase(state: str) -> TaskExecution.Phase:
    """
    Convert the state from the agent to the phase in flyte.
    """
    state = state.lower()
    # timedout is the state of Databricks job. https://docs.databricks.com/en/workflows/jobs/jobs-2.0-api.html#runresultstate
    if state in ["failed", "timeout", "timedout", "canceled"]:
        return TaskExecution.FAILED
    elif state in ["done", "succeeded", "success"]:
        return TaskExecution.SUCCEEDED
    elif state in ["running"]:
        return TaskExecution.RUNNING
    raise ValueError(f"Unrecognized state: {state}")


def is_terminal_phase(phase: TaskExecution.Phase) -> bool:
    """
    Return true if the phase is terminal.
    """
    return phase in [TaskExecution.SUCCEEDED, TaskExecution.ABORTED, TaskExecution.FAILED]


def get_agent_secret(secret_key: str) -> str:
    return flytekit.current_context().secrets.get(secret_key)


class AsyncAgentExecutorMixin:
    """
    This mixin class is used to run the agent task locally, and it's only used for local execution.
    Task should inherit from this class if the task can be run in the agent.
    It can handle asynchronous tasks and synchronous tasks.
    Asynchronous tasks are tasks that take a long time to complete, such as running a query.
    Synchronous tasks run quickly and can return their results instantly. Sending a prompt to ChatGPT and getting a response, or retrieving some metadata from a backend system.
    """

    _clean_up_task: coroutine = None
    _agent: AgentBase = None
    _entity: PythonTask = None

    def execute(self, **kwargs) -> typing.Any:
        ctx = FlyteContext.current_context()
        ss = ctx.serialization_settings or SerializationSettings(ImageConfig())
        output_prefix = ctx.file_access.get_random_remote_directory()

        from flytekit.tools.translator import get_serializable

        self._entity = typing.cast(PythonTask, self)
        task_template = get_serializable(OrderedDict(), ss, self._entity).template
        self._agent = AgentRegistry.get_agent(task_template.type)

        res = asyncio.run(self._create(task_template, output_prefix, kwargs))

        # If the task is synchronous, the agent will return the output from the resource literals.
        if res.HasField("resource"):
            if res.resource.phase != TaskExecution.SUCCEEDED:
                raise FlyteUserException(f"Failed to run the task {self._entity.name}")
            return LiteralMap.from_flyte_idl(res.resource.outputs)

        res = asyncio.run(self._get(resource_meta=res.resource_meta))

        if res.resource.phase != TaskExecution.SUCCEEDED:
            raise FlyteUserException(f"Failed to run the task {self._entity.name}")

        # Read the literals from a remote file, if agent doesn't return the output literals.
        if task_template.interface.outputs and len(res.resource.outputs.literals) == 0:
            local_outputs_file = ctx.file_access.get_random_local_path()
            ctx.file_access.get_data(f"{output_prefix}/output/outputs.pb", local_outputs_file)
            output_proto = utils.load_proto_from_file(literals_pb2.LiteralMap, local_outputs_file)
            return LiteralMap.from_flyte_idl(output_proto)

        return LiteralMap.from_flyte_idl(res.resource.outputs)

    async def _create(
        self, task_template: TaskTemplate, output_prefix: str, inputs: typing.Dict[str, typing.Any] = None
    ) -> CreateTaskResponse:
        ctx = FlyteContext.current_context()

        # Convert python inputs to literals
        literals = inputs or {}
        for k, v in inputs.items():
            literals[k] = TypeEngine.to_literal(ctx, v, type(v), self._entity.interface.inputs[k].type)
        literal_map = LiteralMap(literals)

        if isinstance(self, PythonFunctionTask):
            # Write the inputs to a remote file, so that the remote task can read the inputs from this file.
            path = ctx.file_access.get_random_local_path()
            utils.write_proto_to_file(literal_map.to_flyte_idl(), path)
            ctx.file_access.put_data(path, f"{output_prefix}/inputs.pb")
            task_template = render_task_template(task_template, output_prefix)

        res = await mirror_async_methods(
            self._agent.create,
            output_prefix=output_prefix,
            task_template=task_template,
            inputs=literal_map,
        )

        signal.signal(signal.SIGINT, partial(self.signal_handler, res.resource_meta))  # type: ignore
        return res

    async def _get(self, resource_meta: bytes) -> GetTaskResponse:
        phase = TaskExecution.RUNNING

        progress = Progress(transient=True)
        task = progress.add_task(f"[cyan]Running Task {self._entity.name}...", total=None)
        task_phase = progress.add_task("[cyan]Task phase: RUNNING, Phase message: ", total=None, visible=False)
        task_log_links = progress.add_task("[cyan]Log Links: ", total=None, visible=False)
        with progress:
            while not is_terminal_phase(phase):
                progress.start_task(task)
                time.sleep(1)
                res = await mirror_async_methods(self._agent.get, resource_meta=resource_meta)
                if self._clean_up_task:
                    await self._clean_up_task
                    sys.exit(1)

                phase = res.resource.phase
                progress.update(
                    task_phase,
                    description=f"[cyan]Task phase: {TaskExecution.Phase.Name(phase)}, Phase message: {res.resource.message}",
                    visible=True,
                )
                log_links = ""
                for link in res.log_links:
                    log_links += f"{link.name}: {link.uri}\n"
                if log_links:
                    progress.update(task_log_links, description=f"[cyan]{log_links}", visible=True)

        return res

    def signal_handler(self, resource_meta: bytes, signum: int, frame: FrameType) -> typing.Any:
        if self._clean_up_task is None:
            co = mirror_async_methods(self._agent.delete, resource_meta=resource_meta)
            self._clean_up_task = asyncio.create_task(co)


def render_task_template(tt: TaskTemplate, file_prefix: str) -> TaskTemplate:
    args = tt.container.args
    for i in range(len(args)):
        tt.container.args[i] = args[i].replace("{{.input}}", f"{file_prefix}/inputs.pb")
        tt.container.args[i] = args[i].replace("{{.outputPrefix}}", f"{file_prefix}/output")
        tt.container.args[i] = args[i].replace("{{.rawOutputDataPrefix}}", f"{file_prefix}/raw_output")
        tt.container.args[i] = args[i].replace("{{.checkpointOutputPrefix}}", f"{file_prefix}/checkpoint_output")
        tt.container.args[i] = args[i].replace("{{.prevCheckpointPrefix}}", f"{file_prefix}/prev_checkpoint")
    return tt
