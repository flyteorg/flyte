"""
Hive Tasks
----------

Tasks often start with a data gathering step, and often that data is gathered through Hive. Flytekit allows users to run
any kind of Hive query (including queries with multiple statements and staging query commands).

The principal concept to understand with respect to Hive or any other query-engine based task is how Flyte interacts
with the system. That is, when I write a query, how does Flyte become aware of the result? From the user's perspective,
this is done by carefully constructing your query.

When a Hive (or other querying) task runs, two things need to happen: a) The output data needs to be written to a place
accessible to Flyte, and b) Flyte needs to know exactly what that location is.

You get a couple templating args to help make that happen, along with the usual input interpolation that Flyte provides.

* ``.PerRetryUniqueKey`` - This is a string that will be ``[a-zA-Z0-9_]`` and start with a character. It will be unique
  per retry. Feel free to use it to name temp tables.
* ``RawOutputDataPrefix`` - This is the "directory" (S3/GCS output prefix) where Flyte will expect the outputs. You
  should write the outputs to this location.

"""
from flytekit import kwtypes, task, workflow
from flytekit.types.schema import FlyteSchema
from flytekitplugins.hive import HiveConfig, HiveSelectTask, HiveTask

# %%
# This is the world's simplest query. Note that in order for registration to work properly, you'll need to give your
# Hive task a name that's unique across your project/domain for your Flyte installation.
hive_task_no_io = HiveTask(
    name="sql.hive.no_io",
    inputs={},
    task_config=HiveConfig(cluster_label="flyte"),
    query_template="""
        select 1
    """,
    output_schema_type=None,
)


@workflow
def no_io_wf():
    return hive_task_no_io()


# %%
# This is a hive task that demonstrates how you would construct your typical read query. Note where the ``select 1`` is.
hive_task_w_out = HiveTask(
    name="sql.hive.w_out",
    inputs={},
    task_config=HiveConfig(cluster_label="flyte"),
    query_template="""
    CREATE TEMPORARY TABLE {{ .PerRetryUniqueKey }}_tmp AS select 1;
    CREATE EXTERNAL TABLE {{ .PerRetryUniqueKey }} LIKE {{ .PerRetryUniqueKey }}_tmp STORED AS PARQUET;
    ALTER TABLE {{ .PerRetryUniqueKey }} SET LOCATION '{{ .RawOutputDataPrefix }}';
    INSERT OVERWRITE TABLE {{ .PerRetryUniqueKey }}
        SELECT *
        FROM {{ .PerRetryUniqueKey }}_tmp;
    DROP TABLE {{ .PerRetryUniqueKey }};
    """,
    output_schema_type=FlyteSchema,
)

# %%
# .. note::
#
#    There is a helper task that will automatically do the wrapping above. Please be patient as we fill out these docs.
#


@workflow
def with_output_wf() -> FlyteSchema:
    return hive_task_w_out()


# %%
# This just demonstrates the things you can do. Note that when an input is a FlyteSchema, the value filled in will
# be the uri, i.e. where the data is stored.
demo_all = HiveSelectTask(
    name="sql.hive.demo_all",
    inputs=kwtypes(ds=str, earlier_schema=FlyteSchema),
    task_config=HiveConfig(cluster_label="flyte"),
    select_query="""
    SELECT '.PerRetryUniqueKey' as template_key, '{{ .PerRetryUniqueKey }}' as template_value
    UNION
    SELECT '.RawOutputDataPrefix' as template_key, '{{ .RawOutputDataPrefix }}' as template_value
    UNION
    SELECT '.inputs.earlier_schema' as template_key, '{{ .inputs.earlier_schema }}' as template_value
    UNION
    SELECT '.inputs.ds' as template_key, '{{ .inputs.ds }}' as template_value
    """,
    output_schema_type=FlyteSchema,
)


@task
def print_schema(s: FlyteSchema):
    df = s.open().all()
    print(df.to_markdown())


@workflow
def full_hive_demo_wf() -> FlyteSchema:
    s = hive_task_w_out()
    demo_schema = demo_all(ds="2020-01-01", earlier_schema=s)
    print_schema(s=demo_schema)
    return demo_schema
