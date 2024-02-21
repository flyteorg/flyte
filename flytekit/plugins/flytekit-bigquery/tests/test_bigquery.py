from collections import OrderedDict

import pytest
from flytekitplugins.bigquery import BigQueryConfig, BigQueryTask
from google.cloud.bigquery import QueryJobConfig
from google.protobuf import json_format
from google.protobuf.struct_pb2 import Struct

from flytekit import StructuredDataset, kwtypes, workflow
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.extend import get_serializable

query_template = "SELECT * FROM `bigquery-public-data.crypto_dogecoin.transactions` WHERE @version = 1 LIMIT 10"


def test_serialization():
    bigquery_task = BigQueryTask(
        name="flytekit.demo.bigquery_task.query",
        inputs=kwtypes(ds=str),
        task_config=BigQueryConfig(
            ProjectID="Flyte", Location="Asia", QueryJobConfig=QueryJobConfig(allow_large_results=True)
        ),
        query_template=query_template,
        output_structured_dataset_type=StructuredDataset,
    )

    @workflow
    def my_wf(ds: str) -> StructuredDataset:
        return bigquery_task(ds=ds)

    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="proj",
        domain="dom",
        version="123",
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
        env={},
    )

    task_spec = get_serializable(OrderedDict(), serialization_settings, bigquery_task)

    assert "SELECT * FROM `bigquery-public-data.crypto_dogecoin.transactions`" in task_spec.template.sql.statement
    assert "@version" in task_spec.template.sql.statement
    assert task_spec.template.sql.dialect == task_spec.template.sql.Dialect.ANSI
    s = Struct()
    s.update({"ProjectID": "Flyte", "Location": "Asia", "allowLargeResults": True})
    assert task_spec.template.custom == json_format.MessageToDict(s)
    assert len(task_spec.template.interface.inputs) == 1
    assert len(task_spec.template.interface.outputs) == 1

    admin_workflow_spec = get_serializable(OrderedDict(), serialization_settings, my_wf)
    assert admin_workflow_spec.template.interface.outputs["o0"].type.structured_dataset_type is not None
    assert admin_workflow_spec.template.outputs[0].var == "o0"
    assert admin_workflow_spec.template.outputs[0].binding.promise.node_id == "n0"
    assert admin_workflow_spec.template.outputs[0].binding.promise.var == "results"


def test_local_exec():
    bigquery_task = BigQueryTask(
        name="flytekit.demo.bigquery_task.query2",
        inputs=kwtypes(ds=str),
        query_template=query_template,
        task_config=BigQueryConfig(ProjectID="Flyte", Location="Asia"),
        output_structured_dataset_type=StructuredDataset,
    )

    assert len(bigquery_task.interface.inputs) == 1
    assert len(bigquery_task.interface.outputs) == 1

    # will not run locally
    with pytest.raises(Exception):
        bigquery_task()
