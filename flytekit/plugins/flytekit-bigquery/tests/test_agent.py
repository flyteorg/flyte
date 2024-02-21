import json
from dataclasses import asdict
from datetime import timedelta
from unittest import mock

from flyteidl.core.execution_pb2 import TaskExecution
from flytekitplugins.bigquery.agent import Metadata

import flytekit.models.interface as interface_models
from flytekit.extend.backend.base_agent import AgentRegistry
from flytekit.interfaces.cli_identifiers import Identifier
from flytekit.models import literals, task, types
from flytekit.models.core.identifier import ResourceType
from flytekit.models.task import Sql, TaskTemplate


@mock.patch("google.cloud.bigquery.job.QueryJob")
@mock.patch("google.cloud.bigquery.Client")
def test_bigquery_agent(mock_client, mock_query_job):
    job_id = "dummy_id"
    mock_instance = mock_client.return_value
    mock_query_job_instance = mock_query_job.return_value
    mock_query_job_instance.state.return_value = "SUCCEEDED"
    mock_query_job_instance.job_id.return_value = job_id

    class MockDestination:
        def __init__(self):
            self.project = "dummy_project"
            self.dataset_id = "dummy_dataset"
            self.table_id = "dummy_table"

    class MockJob:
        def __init__(self):
            self.state = "SUCCEEDED"
            self.job_id = job_id
            self.destination = MockDestination()

    setattr(MockJob, "errors", None)

    mock_instance.get_job.return_value = MockJob()
    mock_instance.query.return_value = MockJob()
    mock_instance.cancel_job.return_value = MockJob()

    agent = AgentRegistry.get_agent("bigquery_query_job_task")

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
        "Location": "us-central1",
        "ProjectID": "dummy_project",
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
        custom=task_config,
        metadata=task_metadata,
        interface=interfaces,
        type="bigquery_query_job_task",
        sql=Sql("SELECT 1"),
    )

    metadata_bytes = json.dumps(
        asdict(Metadata(job_id="dummy_id", project="dummy_project", location="us-central1"))
    ).encode("utf-8")
    assert agent.create("/tmp", dummy_template, task_inputs).resource_meta == metadata_bytes
    res = agent.get(metadata_bytes)
    assert res.resource.phase == TaskExecution.SUCCEEDED
    assert (
        res.resource.outputs.literals["results"].scalar.structured_dataset.uri
        == "bq://dummy_project:dummy_dataset.dummy_table"
    )
    assert res.log_links[0].name == "BigQuery Console"
    assert (
        res.log_links[0].uri
        == "https://console.cloud.google.com/bigquery?project=dummy_project&j=bq:us-central1:dummy_id&page=queryresults"
    )
    agent.delete(metadata_bytes)
    mock_instance.cancel_job.assert_called()
