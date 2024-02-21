from collections import OrderedDict

import pytest
from flytekitplugins.athena import AthenaConfig, AthenaTask

from flytekit import kwtypes, workflow
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.extend import get_serializable
from flytekit.types.schema import FlyteSchema


def test_serialization():
    athena_task = AthenaTask(
        name="flytekit.demo.athena_task.query",
        inputs=kwtypes(ds=str),
        task_config=AthenaConfig(database="mnist", catalog="my_catalog", workgroup="my_wg"),
        query_template="""
            insert overwrite directory '{{ .rawOutputDataPrefix }}' stored as parquet
            select *
            from blah
            where ds = '{{ .Inputs.ds }}'
        """,
        # the schema literal's backend uri will be equal to the value of .raw_output_data
        output_schema_type=FlyteSchema,
    )

    @workflow
    def my_wf(ds: str) -> FlyteSchema:
        return athena_task(ds=ds)

    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="proj",
        domain="dom",
        version="123",
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
        env={},
    )
    task_spec = get_serializable(OrderedDict(), serialization_settings, athena_task)
    assert "{{ .rawOutputDataPrefix" in task_spec.template.custom["statement"]
    assert "insert overwrite directory" in task_spec.template.custom["statement"]
    assert "mnist" == task_spec.template.custom["schema"]
    assert "my_catalog" == task_spec.template.custom["catalog"]
    assert "my_wg" == task_spec.template.custom["routingGroup"]
    assert len(task_spec.template.interface.inputs) == 1
    assert len(task_spec.template.interface.outputs) == 1

    admin_workflow_spec = get_serializable(OrderedDict(), serialization_settings, my_wf)
    assert admin_workflow_spec.template.interface.outputs["o0"].type.schema is not None
    assert admin_workflow_spec.template.outputs[0].var == "o0"
    assert admin_workflow_spec.template.outputs[0].binding.promise.node_id == "n0"
    assert admin_workflow_spec.template.outputs[0].binding.promise.var == "results"


def test_local_exec():
    athena_task = AthenaTask(
        name="flytekit.demo.athena_task.query2",
        inputs=kwtypes(ds=str),
        query_template="""
            insert overwrite directory '{{ .rawOutputDataPrefix }}' stored as parquet
            select *
            from blah
            where ds = '{{ .Inputs.ds }}'
        """,
        # the schema literal's backend uri will be equal to the value of .raw_output_data
        output_schema_type=FlyteSchema,
    )

    assert len(athena_task.interface.inputs) == 1
    assert len(athena_task.interface.outputs) == 1

    # will not run locally
    with pytest.raises(Exception):
        athena_task()
