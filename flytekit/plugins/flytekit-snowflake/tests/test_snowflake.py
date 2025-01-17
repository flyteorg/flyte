from collections import OrderedDict

import pytest
from flytekitplugins.snowflake import SnowflakeConfig, SnowflakeTask

from flytekit import kwtypes, workflow
from flytekit.configuration import Image, ImageConfig, SerializationSettings
from flytekit.extend import get_serializable
from flytekit.types.schema import FlyteSchema

query_template = """
            insert overwrite directory '{{ .rawOutputDataPrefix }}' stored as parquet
            select *
            from my_table
            where ds = '{{ .Inputs.ds }}'
        """


def test_serialization():
    snowflake_task = SnowflakeTask(
        name="flytekit.demo.snowflake_task.query",
        inputs=kwtypes(ds=str),
        task_config=SnowflakeConfig(
            account="snowflake", warehouse="my_warehouse", schema="my_schema", database="my_database"
        ),
        query_template=query_template,
        # the schema literal's backend uri will be equal to the value of .raw_output_data
        output_schema_type=FlyteSchema,
    )

    @workflow
    def my_wf(ds: str) -> FlyteSchema:
        return snowflake_task(ds=ds)

    default_img = Image(name="default", fqn="test", tag="tag")
    serialization_settings = SerializationSettings(
        project="proj",
        domain="dom",
        version="123",
        image_config=ImageConfig(default_image=default_img, images=[default_img]),
        env={},
    )

    task_spec = get_serializable(OrderedDict(), serialization_settings, snowflake_task)

    assert "{{ .rawOutputDataPrefix" in task_spec.template.sql.statement
    assert "insert overwrite directory" in task_spec.template.sql.statement
    assert task_spec.template.sql.dialect == task_spec.template.sql.Dialect.ANSI
    assert "snowflake" == task_spec.template.config["account"]
    assert "my_warehouse" == task_spec.template.config["warehouse"]
    assert "my_schema" == task_spec.template.config["schema"]
    assert "my_database" == task_spec.template.config["database"]
    assert len(task_spec.template.interface.inputs) == 1
    assert len(task_spec.template.interface.outputs) == 1

    admin_workflow_spec = get_serializable(OrderedDict(), serialization_settings, my_wf)
    assert admin_workflow_spec.template.interface.outputs["o0"].type.schema is not None
    assert admin_workflow_spec.template.outputs[0].var == "o0"
    assert admin_workflow_spec.template.outputs[0].binding.promise.node_id == "n0"
    assert admin_workflow_spec.template.outputs[0].binding.promise.var == "results"


def test_local_exec():
    snowflake_task = SnowflakeTask(
        name="flytekit.demo.snowflake_task.query2",
        inputs=kwtypes(ds=str),
        query_template="select 1\n",
        # the schema literal's backend uri will be equal to the value of .raw_output_data
        output_schema_type=FlyteSchema,
    )

    assert len(snowflake_task.interface.inputs) == 1
    assert snowflake_task.query_template == "select 1"
    assert len(snowflake_task.interface.outputs) == 1

    # will not run locally
    with pytest.raises(Exception):
        snowflake_task()


def test_sql_template():
    snowflake_task = SnowflakeTask(
        name="flytekit.demo.snowflake_task.query2",
        inputs=kwtypes(ds=str),
        query_template="""select 1 from\t
         custom where column = 1""",
        output_schema_type=FlyteSchema,
    )
    assert snowflake_task.query_template == "select 1 from custom where column = 1"
