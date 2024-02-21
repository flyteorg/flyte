import importlib
import typing
from typing import Optional

import cloudpickle
import jsonpickle
from flyteidl.admin.agent_pb2 import (
    CreateTaskResponse,
    DeleteTaskResponse,
    GetTaskResponse,
    Resource,
)
from flyteidl.core.execution_pb2 import TaskExecution

from flytekit import FlyteContextManager
from flytekit.core.type_engine import TypeEngine
from flytekit.extend.backend.base_agent import AgentBase, AgentRegistry
from flytekit.models.literals import LiteralMap
from flytekit.models.task import TaskTemplate
from flytekit.sensor.base_sensor import INPUTS, SENSOR_CONFIG_PKL, SENSOR_MODULE, SENSOR_NAME

T = typing.TypeVar("T")


class SensorEngine(AgentBase):
    name = "Sensor"

    def __init__(self):
        super().__init__(task_type="sensor")

    async def create(
        self, output_prefix: str, task_template: TaskTemplate, inputs: Optional[LiteralMap] = None, **kwargs
    ) -> CreateTaskResponse:
        python_interface_inputs = {
            name: TypeEngine.guess_python_type(lt.type) for name, lt in task_template.interface.inputs.items()
        }
        ctx = FlyteContextManager.current_context()
        if inputs:
            native_inputs = TypeEngine.literal_map_to_kwargs(ctx, inputs, python_interface_inputs)
            task_template.custom[INPUTS] = native_inputs
        return CreateTaskResponse(resource_meta=cloudpickle.dumps(task_template.custom))

    async def get(self, resource_meta: bytes, **kwargs) -> GetTaskResponse:
        meta = cloudpickle.loads(resource_meta)

        sensor_module = importlib.import_module(name=meta[SENSOR_MODULE])
        sensor_def = getattr(sensor_module, meta[SENSOR_NAME])
        sensor_config = jsonpickle.decode(meta[SENSOR_CONFIG_PKL]) if meta.get(SENSOR_CONFIG_PKL) else None

        inputs = meta.get(INPUTS, {})
        cur_phase = (
            TaskExecution.SUCCEEDED
            if await sensor_def("sensor", config=sensor_config).poke(**inputs)
            else TaskExecution.RUNNING
        )
        return GetTaskResponse(resource=Resource(phase=cur_phase, outputs=None))

    async def delete(self, resource_meta: bytes, **kwargs) -> DeleteTaskResponse:
        return DeleteTaskResponse()


AgentRegistry.register(SensorEngine())
