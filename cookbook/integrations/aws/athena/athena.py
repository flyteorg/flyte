"""
Athena Query
############

This example shows how to use a Flyte AthenaTask to execute a query.
"""

from flytekit import kwtypes, task, workflow
from flytekit.types.schema import FlyteSchema
from flytekitplugins.athena import AthenaConfig, AthenaTask

# %%
# This is the world's simplest query. Note that in order for registration to work properly, you'll need to give your
# Athena task a name that's unique across your project/domain for your Flyte installation.
athena_task_no_io = AthenaTask(
    name="sql.athena.no_io",
    inputs={},
    query_template="""
        select 1
    """,
    output_schema_type=None,
    task_config=AthenaConfig(database="mnist"),
)


@workflow
def no_io_wf():
    return athena_task_no_io()


# %%
# Of course, in real world applications we are usually more interested in using Athena to query a dataset.
# In this case we've populated our vaccinations table with the publicly available dataset
# `here <https://www.kaggle.com/gpreda/covid-world-vaccination-progress>`__.
# For a primer on how upload a dataset, checkout of the official
# `AWS docs <https://docs.aws.amazon.com/quicksight/latest/user/create-a-data-set-athena.html>`__.
# The data is formatted according to this schema:
#
# +----------------------------------------------+
# | country (string)                             |
# +----------------------------------------------+
# | iso_code (string)                            |
# +----------------------------------------------+
# | date (string)                                |
# +----------------------------------------------+
# | total_vaccinations (string)                  |
# +----------------------------------------------+
# | people_vaccinated (string)                   |
# +----------------------------------------------+
# | people_fully_vaccinated (string)             |
# +----------------------------------------------+
# | daily_vaccinations_raw (string)              |
# +----------------------------------------------+
# | daily_vaccinations (string)                  |
# +----------------------------------------------+
# | total_vaccinations_per_hundred (string)      |
# +----------------------------------------------+
# | people_vaccinated_per_hundred (string)       |
# +----------------------------------------------+
# | people_fully_vaccinated_per_hundred (string) |
# +----------------------------------------------+
# | daily_vaccinations_per_million (string)      |
# +----------------------------------------------+
# | vaccines (string)                            |
# +----------------------------------------------+
# | source_name (string)                         |
# +----------------------------------------------+
# | source_website (string)                      |
# +----------------------------------------------+
#
# Let's look out how we can parameterize our query to filter results for a specific country, provided as a user input
# specifying a country iso code.
# We'll produce a FlyteSchema that we can use in downstream flyte tasks for further analysis or manipulation.
# Note that we cache this output data so we don't have to re-run the query in future workflow iterations
# should we decide to change how we manipulate data downstream

athena_task_templatized_query = AthenaTask(
    name="sql.athena.w_io",
    # Define inputs as well as their types that can be used to customize the query.
    inputs=kwtypes(iso_code=str),
    task_config=AthenaConfig(database="vaccinations"),
    query_template="""
    SELECT * FROM vaccinations where iso_code like  {{ .inputs.iso_code }}
    """,
    # While we define a generic schema as the output here, Flyte supports more strongly typed schemas to provide
    # better compile-time checks for task compatibility. Refer to :py:class:`flytekit.FlyteSchema` for more details
    output_schema_type=FlyteSchema,
    # Cache the output data so we don't have to re-run the query in future workflow iterations
    # should we decide to change how we manipulate data downstream.
    # For more information about caching, visit :ref:`Task Caching <task_cache>`
    cache=True,
    cache_version="1.0",
)


# %%
# Now we (trivially) clean up and interact with the data produced from the above Athena query in a separate Flyte task.
@task
def manipulate_athena_schema(s: FlyteSchema) -> FlyteSchema:
    df = s.open().all()
    return df[df.total_vaccinations.notnull()]


@workflow
def full_athena_wf(country_iso_code: str) -> FlyteSchema:
    demo_schema = athena_task_templatized_query(iso_code=country_iso_code)
    return manipulate_athena_schema(s=demo_schema)
