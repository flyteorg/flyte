"""
Using Schemas
-------------

.. tags:: DataFrame, Basic

This example explains how an untyped schema is passed between tasks using :py:class:`pandas.DataFrame`.
Flytekit makes it possible for users to directly return or accept a :py:class:`pandas.DataFrame`, which are automatically
converted into flyte's abstract representation of a schema object

"""
import pandas
from flytekit import task, workflow

# %%
# Flytekit allows users to directly use pandas.dataframe in their tasks as long as they import
# Note: # noqa: F401. This is to ignore pylint complaining about unused imports
from flytekit.types import schema  # noqa: F401


# %%
# This task generates a pandas.DataFrame and returns it. The Dataframe itself will be serialized to an intermediate
# format like parquet before passing between tasks
@task
def get_df(a: int) -> pandas.DataFrame:
    """
    Generate a sample dataframe
    """
    return pandas.DataFrame(data={"col1": [a, 2], "col2": [a, 4]})


# %%
# This task shows an example of transforming a dataFrame
@task
def add_df(df: pandas.DataFrame) -> pandas.DataFrame:
    """
    Append some data to the dataframe.
    NOTE: this may result in runtime failures if the columns do not match
    """
    return df.append(pandas.DataFrame(data={"col1": [5, 10], "col2": [5, 10]}))


# %%
# The workflow shows that passing DataFrame's between tasks is as simple as passing dataFrames in memory
@workflow
def df_wf(a: int) -> pandas.DataFrame:
    """
    Pass data between the dataframes
    """
    df = get_df(a=a)
    return add_df(df=df)


# %%
# The entire program can be run locally
if __name__ == "__main__":
    print(f"Running {__file__} main...")
    print(f"Running df_wf(a=42) {df_wf(a=42)}")
