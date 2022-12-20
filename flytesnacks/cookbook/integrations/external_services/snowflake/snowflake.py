"""
Snowflake Query
###############

This example shows how to use a Flyte SnowflakeTask to execute a query.
"""

from flytekit import kwtypes, workflow
from flytekitplugins.snowflake import SnowflakeConfig, SnowflakeTask

# %%
# This is the world's simplest query. Note that in order for registration to work properly, you'll need to give your
# Snowflake task a name that's unique across your project/domain for your Flyte installation.
snowflake_task_no_io = SnowflakeTask(
    name="sql.snowflake.no_io",
    inputs={},
    query_template="SELECT 1",
    output_schema_type=None,
    task_config=SnowflakeConfig(
        account="ha63105.us-central1.gcp",
        database="SNOWFLAKE_SAMPLE_DATA",
        schema="TPCH_SF1000",
        warehouse="COMPUTE_WH",
    ),
)


@workflow
def no_io_wf():
    return snowflake_task_no_io()


# %%
# Of course, in real world applications we are usually more interested in using Snowflake to query a dataset.
# In this case we use SNOWFLAKE_SAMPLE_DATA which is default dataset in snowflake service.
# `here <https://docs.snowflake.com/en/user-guide/sample-data.html>`__
# The data is formatted according to this schema:
#
# +----------------------------------------------+
# | C_CUSTKEY (int)                              |
# +----------------------------------------------+
# | C_NAME (string)                              |
# +----------------------------------------------+
# | C_ADDRESS (string)                           |
# +----------------------------------------------+
# | C_NATIONKEY (int)                            |
# +----------------------------------------------+
# | C_PHONE (string)                             |
# +----------------------------------------------+
# | C_ACCTBAL (float)                            |
# +----------------------------------------------+
# | C_MKTSEGMENT (string)                        |
# +----------------------------------------------+
# | C_COMMENT (string)                           |
# +----------------------------------------------+
#
# Let's look out how we can parameterize our query to filter results for a specific country, provided as a user input
# specifying a nation key.

snowflake_task_templatized_query = SnowflakeTask(
    name="sql.snowflake.w_io",
    # Define inputs as well as their types that can be used to customize the query.
    inputs=kwtypes(nation_key=int),
    task_config=SnowflakeConfig(
        account="ha63105.us-central1.gcp",
        database="SNOWFLAKE_SAMPLE_DATA",
        schema="TPCH_SF1000",
        warehouse="COMPUTE_WH",
    ),
    query_template="SELECT * from CUSTOMER where C_NATIONKEY =  {{ .inputs.nation_key }} limit 100",
)


@workflow
def full_snowflake_wf(nation_key: int):
    return snowflake_task_templatized_query(nation_key=nation_key)


# %%
# Check query result on snowflake console: ``https://<account>.snowflakecomputing.com/console#/monitoring/queries/detail``
#
# For example, https://ha63105.us-central1.gcp.snowflakecomputing.com/console#/monitoring/queries/detail
