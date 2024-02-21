import json
from dataclasses import asdict
from datetime import timedelta
from unittest import mock
from unittest.mock import MagicMock

import pytest
from flyteidl.admin.agent_pb2 import DeleteTaskResponse
from flyteidl.core.execution_pb2 import TaskExecution
from flytekitplugins.snowflake.agent import Metadata

import flytekit.models.interface as interface_models
from flytekit import lazy_module
from flytekit.extend.backend.base_agent import AgentRegistry
from flytekit.interfaces.cli_identifiers import Identifier
from flytekit.models import literals, task, types
from flytekit.models.core.identifier import ResourceType
from flytekit.models.task import Sql, TaskTemplate


@mock.patch("flytekitplugins.snowflake.agent.get_private_key", return_value="pb")
@pytest.mark.asyncio
async def test_snowflake_agent(mock_get_private_key):
    query_status_mock = MagicMock()
    query_status_mock.name = "SUCCEEDED"

    # Configure the mock connection to return the mock status object
    snowflake_connector = lazy_module("snowflake.connector")
    snowflake_connector.connect = MagicMock()
    mock_conn_instance = snowflake_connector.connect.return_value
    mock_conn_instance.get_query_status_throw_if_error.return_value = query_status_mock

    agent = AgentRegistry.get_agent("snowflake")

    task_id = Identifier(
        resource_type=ResourceType.TASK, project="project", domain="domain", name="name", version="version"
    )

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

    task_config = {
        "user": "dummy_user",
        "account": "dummy_account",
        "database": "dummy_database",
        "schema": "dummy_schema",
        "warehouse": "dummy_warehouse",
        "table": "dummy_table",
    }

    int_type = types.LiteralType(types.SimpleType.INTEGER)
    interfaces = interface_models.TypedInterface(
        {
            "a": interface_models.Variable(int_type, "description1"),
            "b": interface_models.Variable(int_type, "description2"),
        },
        {},
    )
    task_inputs = literals.LiteralMap(
        {
            "a": literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=1))),
            "b": literals.Literal(scalar=literals.Scalar(primitive=literals.Primitive(integer=1))),
        },
    )

    dummy_template = TaskTemplate(
        id=task_id,
        custom=None,
        config=task_config,
        metadata=task_metadata,
        interface=interfaces,
        type="snowflake",
        sql=Sql("SELECT 1"),
    )

    metadata = Metadata(
        user="dummy_user",
        account="dummy_account",
        table="dummy_table",
        database="dummy_database",
        schema="dummy_schema",
        warehouse="dummy_warehouse",
        query_id="dummy_query_id",
    )

    res = await agent.create("/tmp", dummy_template, task_inputs)
    metadata.query_id = Metadata(**json.loads(res.resource_meta.decode("utf-8"))).query_id
    metadata_bytes = json.dumps(asdict(metadata)).encode("utf-8")
    assert res.resource_meta == metadata_bytes

    res = await agent.get(metadata_bytes)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    assert (
        res.resource.outputs.literals["results"].scalar.structured_dataset.uri
        == "snowflake://dummy_user:dummy_account/dummy_warehouse/dummy_database/dummy_schema/dummy_table"
    )

    delete_response = await agent.delete(metadata_bytes)

    # Assert the response
    assert isinstance(delete_response, DeleteTaskResponse)

    # Verify that the expected methods were called on the mock cursor
    mock_cursor = mock_conn_instance.cursor.return_value
    mock_cursor.fetchall.assert_called_once()

    mock_cursor.execute.assert_called_once_with(f"SELECT SYSTEM$CANCEL_QUERY('{metadata.query_id}')")
    mock_cursor.fetchall.assert_called_once()

    # Verify that the connection was closed
    mock_cursor.close.assert_called_once()
    mock_conn_instance.close.assert_called_once()
