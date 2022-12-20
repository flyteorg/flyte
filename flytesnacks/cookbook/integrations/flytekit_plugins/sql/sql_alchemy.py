"""
SQLAlchemy
----------

SQLAlchemy is the Python SQL toolkit and Object Relational Mapper that gives application developers the full power and flexibility of SQL.

Flyte provides an easy-to-use interface to utilize SQLAlchemy to connect to various SQL Databases. In this example, we use a Postgres DB to understand how SQLAlchemy can be used within Flyte.

This task will run with a pre-built container, and thus users needn't build one.
You can simply implement the task, then register and execute it immediately.

First, install the Flyte Sqlalchemy plugin:

.. prompt:: bash $

    pip install flytekitplugins-sqlalchemy

"""

# %%
# Let's first import the libraries.
from flytekit import kwtypes, task, workflow
from flytekit.types.schema import FlyteSchema
from flytekitplugins.sqlalchemy import SQLAlchemyConfig, SQLAlchemyTask

# %%
# First we define a ``SQLALchemyTask``, which returns the first ``n`` records from the ``rna`` table of the
# `RNA central database <https://rnacentral.org/help/public-database>`__ . Since this database is public, we can
# hard-code the database URI, including the user and password in a string.
#
# .. note::
#
#   The output of SQLAlchemyTask is a ``FlyteSchema`` by default.
#
# .. caution::
#
#    **Never** store passwords for proprietary or sensitive databases! If you need to store and access secrets in a task,
#    Flyte provides a convenient API. See :ref:`sphx_glr_auto_core_containerization_use_secrets.py` for more details.

DATABASE_URI = (
    "postgresql://reader:NWDMCE5xdipIjRrp@hh-pgsql-public.ebi.ac.uk:5432/pfmegrnargs"
)

# Here we define the schema of the expected output of the query, which we then re-use in the `get_mean_length` task.
DataSchema = FlyteSchema[kwtypes(sequence_length=int)]

sql_task = SQLAlchemyTask(
    "rna",
    query_template="""
        select len as sequence_length from rna
        where len >= {{ .inputs.min_length }}
        and len <= {{ .inputs.max_length }}
        limit {{ .inputs.limit }}
    """,
    inputs=kwtypes(min_length=int, max_length=int, limit=int),
    output_schema_type=DataSchema,
    task_config=SQLAlchemyConfig(uri=DATABASE_URI),
)


# %%
# Next, we define a task that computes the mean length of sequences in the subset of RNA sequences that our query
# returned.
# Note for those running this in your live Flyte backend via ``pyflyte run``.  ``run`` will use the default flytekit
# image if one is not specified.  The default flytekit image does not have the sqlalchemy flytekit plugin installed.
# To correctly kick off an execution of this task, you'll need to use the following command.
#
# .. code::
#
#   pyflyte --config ~/.flyte/your-config.yaml run --destination-dir /app --remote --image ghcr.io/flyteorg/flytekit:py3.8-sqlalchemy-latest integrations/flytekit_plugins/sql/sql_alchemy.py my_wf --min_length 3 --max_length 100 --limit 50
#
# Note also we added the ``destination-dir`` argument, since by default ``pyflyte run`` copies code into ``/root`` which
# is not what that image's workdir is set to.
@task
def get_mean_length(data: DataSchema) -> float:
    dataframe = data.open().all()
    return dataframe["sequence_length"].mean().item()


# %%
# Finally, we put everything together into a workflow:
@workflow
def my_wf(min_length: int, max_length: int, limit: int) -> float:
    return get_mean_length(
        data=sql_task(min_length=min_length, max_length=max_length, limit=limit)
    )


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(my_wf(min_length=50, max_length=200, limit=5))
