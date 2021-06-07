"""
Quickstart
----------

In this example we'll show you how to use DoltTable
to annotate dataframe inputs and outputs in your flyte tasks.

"""
import os
import sys
import typing

from flytekitplugins.dolt.schema import DoltConfig, DoltTable
from flytekit import task, workflow
import pandas as pd

doltdb_path = os.path.join(os.path.dirname(__file__), "foo")

rabbits_conf = DoltConfig(
    db_path=doltdb_path,
    tablename="rabbits",
)

@task
def populate_rabbits(a: int) -> DoltTable:
    rabbits = [("George", a), ("Alice", a * 2), ("Sugar Maple", a * 3)]
    df = pd.DataFrame(rabbits, columns=["name", "count"])
    return DoltTable(data=df, config=rabbits_conf)

@task
def unwrap_rabbits(table: DoltTable) -> pd.DataFrame:
    return table.data

@workflow
def wf(a: int) -> pd.DataFrame:
    rabbits = populate_rabbits(a=a)
    df = unwrap_rabbits(table=rabbits)
    return df

if __name__ == "__main__":
    print(f"Running {__file__} main...")
    if len(sys.argv) != 2:
        raise ValueError("Expected 1 argument: a (int)")
    a = int(sys.argv[1])
    result = wf(a=a)
    print(f"Running wf(), returns dataframe\n{result}\n{result.dtypes}")
