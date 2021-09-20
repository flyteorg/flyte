"""
SQLAlchemy
----------

SQLAlchemy is the Python SQL toolkit and Object Relational Mapper that gives application developers the full power and flexibility of SQL.

Flyte provides an easy-to-use interface to utilize SQLAlchemy to connect to various SQL Databases. In this example, we use a Postgres DB to understand how SQLAlchemy can be used within Flyte.

This task will run with a pre-built container, and thus users needn't build one.
You can simply implement the task, then register and execute it immediately.

.. note::
    This example works locally only. For remote, you need to input a valid URI.
"""

# %%
# Let's first import the libraries.
import typing

import pandas
import psycopg2
from flytekit import kwtypes, task, workflow
from flytekitplugins.sqlalchemy import SQLAlchemyConfig, SQLAlchemyTask


# %%
# Next, we define a ``create_db`` function to create a Postgres DB using `psycopg2 <https://pypi.org/project/psycopg2/>`__.
def create_db(
    conn_info: typing.Dict[str, str],
) -> None:
    psql_connection_string = (
        f"user={conn_info['user']} password={conn_info['password']}"
    )
    conn = psycopg2.connect(psql_connection_string)
    cur = conn.cursor()

    conn.autocommit = True
    sql_query = f"CREATE DATABASE {conn_info['database']}"

    try:
        cur.execute(sql_query)
    except Exception as e:
        print(f"{type(e).__name__}: {e}")
        print(f"Query: {cur.query}")
        cur.close()
    else:
        conn.autocommit = False


# %%
# We define a ``run_query`` function to create a table and insert data entries into it.
def run_query(
    sql_query: str,
    conn: psycopg2.extensions.connection,
    cur: psycopg2.extensions.cursor,
) -> None:
    try:
        cur.execute(sql_query)
    except Exception as e:
        print(f"{type(e).__name__}: {e}")
        print(f"Query: {cur.query}")
        conn.rollback()
        cur.close()
    else:
        conn.commit()


local_pg_server = f"postgresql://localhost/flights"


# %%
# We define a helper function.
def pg_server() -> str:
    conn_info = {"database": "flights", "user": "postgres", "password": "postgres"}

    # create database
    create_db(conn_info)

    # connect to the database
    connection = psycopg2.connect(**conn_info)
    cursor = connection.cursor()

    # run queries
    table_sql = """
        CREATE TABLE test (
            origin VARCHAR NOT NULL,
            destination VARCHAR NOT NULL,
            duration INTEGER NOT NULL
        );
    """
    run_query(table_sql, connection, cursor)

    first_entry_sql = """
        INSERT INTO test (origin, destination, duration)
        VALUES ('New York', 'London', 415);
    """
    run_query(first_entry_sql, connection, cursor)

    second_entry_sql = """
        INSERT INTO test (origin, destination, duration)
        VALUES ('Shanghai', 'Paris', 760);
    """
    run_query(second_entry_sql, connection, cursor)

    cursor.close()
    connection.close()

    return local_pg_server


# %%
# Finally, we fetch the number of data entries whose durations lie within the user-specified durations.
#
# .. note::
#
#   The output of SQLAlchemyTask is a :py:class:`~flytekit.types.schema.FlyteSchema` by default.
@task
def get_length(df: pandas.DataFrame) -> int:
    return len(df)


sql_task = SQLAlchemyTask(
    "fetch_flight_data",
    query_template="""
            select * from test
            where duration >= {{ .inputs.lower_duration_cap }}
            and duration <= {{ .inputs.upper_duration_cap }}
        """,
    inputs=kwtypes(lower_duration_cap=int, upper_duration_cap=int),
    task_config=SQLAlchemyConfig(uri="postgresql://localhost/flights"),
)


@workflow
def my_wf(lower_duration_cap: int, upper_duration_cap: int) -> int:
    return get_length(
        df=sql_task(
            lower_duration_cap=lower_duration_cap, upper_duration_cap=upper_duration_cap
        )
    )


if __name__ == "__main__":
    print(f"Starting pg server at {pg_server()}")
    print(f"Running {__file__} main...")
    print(f"Starting Postgres DB at {pg_server()}")
    print(my_wf(lower_duration_cap=300, upper_duration_cap=800))
