import tempfile

import cloudpickle
import pytest
from flyteidl.admin.agent_pb2 import DeleteTaskResponse
from flyteidl.core.execution_pb2 import TaskExecution

import flytekit.models.interface as interface_models
from flytekit.extend.backend.base_agent import AgentRegistry
from flytekit.models import literals, types
from flytekit.sensor import FileSensor
from flytekit.sensor.base_sensor import SENSOR_MODULE, SENSOR_NAME
from tests.flytekit.unit.extend.test_agent import get_task_template


@pytest.mark.asyncio
async def test_sensor_engine():
    interfaces = interface_models.TypedInterface(
        {
            "path": interface_models.Variable(types.LiteralType(types.SimpleType.STRING), "description1"),
        },
        {},
    )
    tmp = get_task_template("sensor")
    tmp._custom = {
        SENSOR_MODULE: FileSensor.__module__,
        SENSOR_NAME: FileSensor.__name__,
    }
    file = tempfile.NamedTemporaryFile()

    tmp._interface = interfaces

    task_inputs = literals.LiteralMap(
        {
            "path": literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(string_value=file.name))),
        },
    )
    agent = AgentRegistry.get_agent("sensor")

    res = await agent.create("/tmp", tmp, task_inputs)

    metadata_bytes = cloudpickle.dumps(tmp.custom)
    assert res.resource_meta == metadata_bytes
    res = await agent.get(metadata_bytes)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    res = await agent.delete(metadata_bytes)
    assert res == DeleteTaskResponse()
