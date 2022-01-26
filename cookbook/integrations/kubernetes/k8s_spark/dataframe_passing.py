"""
.. _intermediate_spark_dataframes_passing:

Converting a Spark DataFrame to a Pandas DataFrame
==================================================

This example shows how users can return a spark.Dataset from a task and consume it as a pandas.DataFrame.
If the dataframe does not fit in memory, it will result in a runtime failure.
"""
import flytekit
import pandas
from flytekit import kwtypes, task, workflow
from flytekit.types.schema import FlyteSchema
from flytekitplugins.spark import Spark

# %%
# .. _df_my_schema_definition:
#
# Define my_schema
# -----------------
# This section defines a simple schema type with 2 columns, `name: str` and `age: int`
my_schema = FlyteSchema[kwtypes(name=str, age=int)]

# %%
# ``create_spark_df`` is a spark task that runs within a spark cotext (and relies on having a spark cluster up and running). This task generates a spark DataFrame whose schema matches the predefined :any:`df_my_schema_definition`
# 
# Notice that the task simply returns a pyspark.DataFrame object, even though the return type specifies  :any:`df_my_schema_definition`
# The flytekit type-system will automatically convert the pyspark.DataFrame to Flyte Schema object.
# FlyteSchema object is an abstract representation of a DataFrame, that can conform to multiple different dataframe formats.


@task(
    task_config=Spark(
        spark_conf={
            "spark.driver.memory": "1000M",
            "spark.executor.memory": "1000M",
            "spark.executor.cores": "1",
            "spark.executor.instances": "2",
            "spark.driver.cores": "1",
        }
    ),
    cache_version="1",
)
def create_spark_df() -> my_schema:
    """
    This spark program returns a spark dataset that conforms to the defined schema. Failure to do so should result
    in a runtime error. TODO: runtime error enforcement
    """
    sess = flytekit.current_context().spark_session
    return sess.createDataFrame(
        [("Alice", 5), ("Bob", 10), ("Charlie", 15),], my_schema.column_names(),
    )


# %%
# The task ``sum_of_all_ages`` receives a parameter of type :any:`df_my_schema_definition`. It is important to note that there is no
# expectation that the schema is a pandas dataframe or a spark dataframe, but just a generic schema object. The Flytekit schema object
# can be read into multiple formats using the ``open()`` method. Default conversion is to :py:class:`pandas.DataFrame`
# Refer to :py:class:`flytekit.types.schema.FlyteSchema` for more details.
#
@task(cache_version="1")
def sum_of_all_ages(s: my_schema) -> int:
    """
    The schema is passed into this task. Schema is just a reference to the actually object and has almost no overhead.
    Only performing an ``open`` on the schema will cause the data to be loaded into memory (also downloaded if this being
    run in a remote setting)
    """
    # This by default returns a pandas.DataFrame object. ``open`` can be parameterized to return other dataframe types
    reader = s.open()
    # supported dataframes
    df: pandas.DataFrame = reader.all()
    return int(df["age"].sum())


# %%
# The schema workflow allows connecting the ``create_spark_df`` with  ``sum_of_all_ages`` because the return type of the first task and the parameter type for the second task match
@workflow
def my_smart_schema() -> int:
    """
    This workflow shows how a simple schema can be created in spark and passed to a python function and accessed as a
    pandas.DataFrame. Flyte Schemas are abstract data frames and not really tied to a specific memory representation.
    """
    df = create_spark_df()
    return sum_of_all_ages(s=df)


# %%
# This program can be executed locally and it should work as expected. This greatly simplifies using disparate DataFrame technologies for the end user.
# New DataFrame technologies can also be dynamically loaded in flytekit's TypeEngine.
if __name__ == "__main__":
    """
    This program can be run locally
    """
    print(f"Running {__file__} main...")
    print(f"Running my_smart_schema()-> {my_smart_schema()}")
