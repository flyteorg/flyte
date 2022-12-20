"""
Dolt Branches
-------------

In this example, we'll show how to use DoltTable along with Dolt's ``Branch`` feature.

"""
import os
import sys
import typing

import pandas as pd
from dolt_integrations.core import NewBranch
from flytekit import task, workflow
from flytekitplugins.dolt.schema import DoltConfig, DoltTable

# %%
# A Simple Workflow
# ^^^^^^^^^^^^^^^^^
# We will run a simple data workflow:
#
# 1. Create a ``users`` table with ``name`` and ``count`` columns.
# 2. Filter the ``users`` table for users with ``count > 5``.
# 3. Record the filtered users' names in a ``big_users`` table.

# %%
# Database Configuration
# ======================
# Let's define our database configuration.
# Our ``DoltConfig`` references a ``foo`` folder containing
# our database. Use either a ``tablename`` or a ``sql`` select
# statement to fetch data.

doltdb_path = os.path.join(os.path.dirname(__file__), "foo")


def generate_confs(a: int) -> typing.Tuple[DoltConfig, DoltConfig, DoltConfig]:
    users_conf = DoltConfig(
        db_path=doltdb_path, tablename="users", branch_conf=NewBranch(f"run/a_is_{a}")
    )

    query_users = DoltTable(
        config=DoltConfig(
            db_path=doltdb_path,
            sql="select * from users where `count` > 5",
            branch_conf=NewBranch(f"run/a_is_{a}"),
        ),
    )

    big_users_conf = DoltConfig(
        db_path=doltdb_path,
        tablename="big_users",
        branch_conf=NewBranch(f"run/a_is_{a}"),
    )

    return users_conf, query_users, big_users_conf


# %%
# .. tip ::
#   A ``DoltTable`` is an  extension of ``DoltConfig`` that wraps a ``pandas.DataFrame`` -- accessible via the ``DoltTable.data``
#   attribute at execution time.

# %%
# Type Annotating Tasks and Workflows
# ===================================
# We can turn our data processing pipeline into a Flyte workflow
# by decorating functions with the :py:func:`~flytekit.task` and :py:func:`~flytekit.workflow` decorators.
# Annotating the inputs and outputs of those functions with Dolt schemas
# indicates how to save and load data between tasks.
#
# The ``DoltTable.data`` attribute loads dataframes for input arguments.
# Return types of ``DoltTable`` save the ``data`` to the
# Dolt database given a connection configuration.


@task
def get_confs(a: int) -> typing.Tuple[DoltConfig, DoltTable, DoltConfig]:
    return generate_confs(a)


@task
def populate_users(a: int, conf: DoltConfig) -> DoltTable:
    users = [("George", a), ("Alice", a * 2), ("Stephanie", a * 3)]
    df = pd.DataFrame(users, columns=["name", "count"])
    return DoltTable(data=df, config=conf)


@task
def filter_users(
    a: int, all_users: DoltTable, filtered_users: DoltTable, conf: DoltConfig
) -> DoltTable:
    usernames = filtered_users.data[["name"]]
    return DoltTable(data=usernames, config=conf)


@task
def count_users(users: DoltTable) -> int:
    return users.data.shape[0]


@workflow
def wf(a: int) -> int:
    user_conf, query_conf, big_user_conf = get_confs(a=a)
    users = populate_users(a=a, conf=user_conf)
    big_users = filter_users(
        a=a, all_users=users, filtered_users=query_conf, conf=big_user_conf
    )
    big_user_cnt = count_users(users=big_users)
    return big_user_cnt


if __name__ == "__main__":
    print(f"Running {__file__} main...")
    if len(sys.argv) != 2:
        raise ValueError("Expected 1 argument: a (int)")
    a = int(sys.argv[1])
    result = wf(a=a)
    print(f"Running wf(), returns int\n{result}\n{type(result)}")

# %%
# We will run this workflow twice:
#
# .. prompt:: $
#
#   python branch_example.py 2
#
# .. prompt:: $
#
#   python branch_example.py 3
#
# Which creates distinct branches for our two ``a`` values:
#
# .. prompt:: $
#
#   cd foo
#   dolt branch
