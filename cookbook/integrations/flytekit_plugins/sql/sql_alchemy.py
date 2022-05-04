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
#   The output of SQLAlchemyTask is a :py:class:`~flytekit.types.schema.FlyteSchema` by default.
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
