import typing

import grpc
from flyteidl.admin.agent_pb2 import (
    CreateTaskRequest,
    CreateTaskResponse,
    DeleteTaskRequest,
    DeleteTaskResponse,
    GetAgentRequest,
    GetAgentResponse,
    GetTaskRequest,
    GetTaskResponse,
    ListAgentsRequest,
    ListAgentsResponse,
)
from flyteidl.service.agent_pb2_grpc import AgentMetadataServiceServicer, AsyncAgentServiceServicer
from prometheus_client import Counter, Summary

from flytekit import logger
from flytekit.exceptions.system import FlyteAgentNotFound
from flytekit.extend.backend.base_agent import AgentRegistry, mirror_async_methods
from flytekit.models.literals import LiteralMap
from flytekit.models.task import TaskTemplate

metric_prefix = "flyte_agent_"
create_operation = "create"
get_operation = "get"
delete_operation = "delete"

# Follow the naming convention. https://prometheus.io/docs/practices/naming/
request_success_count = Counter(
    f"{metric_prefix}requests_success_total",
    "Total number of successful requests",
    ["task_type", "operation"],
)
request_failure_count = Counter(
    f"{metric_prefix}requests_failure_total",
    "Total number of failed requests",
    ["task_type", "operation", "error_code"],
)
request_latency = Summary(
    f"{metric_prefix}request_latency_seconds",
    "Time spent processing agent request",
    ["task_type", "operation"],
)
input_literal_size = Summary(f"{metric_prefix}input_literal_bytes", "Size of input literal", ["task_type"])


def agent_exception_handler(func: typing.Callable):
    async def wrapper(
        self,
        request: typing.Union[CreateTaskRequest, GetTaskRequest, DeleteTaskRequest],
        context: grpc.ServicerContext,
        *args,
        **kwargs,
    ):
        if isinstance(request, CreateTaskRequest):
            task_type = request.template.type
            operation = create_operation
            if request.inputs:
                input_literal_size.labels(task_type=task_type).observe(request.inputs.ByteSize())
        elif isinstance(request, GetTaskRequest):
            task_type = request.task_type
            operation = get_operation
        elif isinstance(request, DeleteTaskRequest):
            task_type = request.task_type
            operation = delete_operation
        else:
            context.set_code(grpc.StatusCode.UNIMPLEMENTED)
            context.set_details("Method not implemented!")
            return

        try:
            with request_latency.labels(task_type=task_type, operation=operation).time():
                res = await func(self, request, context, *args, **kwargs)
            request_success_count.labels(task_type=task_type, operation=operation).inc()
            return res
        except FlyteAgentNotFound:
            error_message = f"Cannot find agent for task type: {task_type}."
            logger.error(error_message)
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(error_message)
            request_failure_count.labels(task_type=task_type, operation=operation, error_code="404").inc()
        except Exception as e:
            error_message = f"failed to {operation} {task_type} task with error {e}."
            logger.error(error_message)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(error_message)
            request_failure_count.labels(task_type=task_type, operation=operation, error_code="500").inc()

    return wrapper


class AsyncAgentService(AsyncAgentServiceServicer):
    @agent_exception_handler
    async def CreateTask(self, request: CreateTaskRequest, context: grpc.ServicerContext) -> CreateTaskResponse:
        tmp = TaskTemplate.from_flyte_idl(request.template)
        inputs = LiteralMap.from_flyte_idl(request.inputs) if request.inputs else None
        agent = AgentRegistry.get_agent(tmp.type)

        logger.info(f"{tmp.type} agent start creating the job")
        return await mirror_async_methods(
            agent.create, output_prefix=request.output_prefix, task_template=tmp, inputs=inputs
        )

    @agent_exception_handler
    async def GetTask(self, request: GetTaskRequest, context: grpc.ServicerContext) -> GetTaskResponse:
        agent = AgentRegistry.get_agent(request.task_type)
        logger.info(f"{agent.task_type} agent start checking the status of the job")
        return await mirror_async_methods(agent.get, resource_meta=request.resource_meta)

    @agent_exception_handler
    async def DeleteTask(self, request: DeleteTaskRequest, context: grpc.ServicerContext) -> DeleteTaskResponse:
        agent = AgentRegistry.get_agent(request.task_type)
        logger.info(f"{agent.task_type} agent start deleting the job")
        return await mirror_async_methods(agent.delete, resource_meta=request.resource_meta)


class AgentMetadataService(AgentMetadataServiceServicer):
    async def GetAgent(self, request: GetAgentRequest, context: grpc.ServicerContext) -> GetAgentResponse:
        return GetAgentResponse(agent=AgentRegistry._METADATA[request.name])

    async def ListAgents(self, request: ListAgentsRequest, context: grpc.ServicerContext) -> ListAgentsResponse:
        agents = [agent for agent in AgentRegistry._METADATA.values()]
        return ListAgentsResponse(agents=agents)
