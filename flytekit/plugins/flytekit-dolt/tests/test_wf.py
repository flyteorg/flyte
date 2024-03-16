import os
import shutil
import subprocess
import tempfile
import typing
from pathlib import Path

import doltcli as dolt
import pandas
import pytest
from dolt_integrations.core import NewBranch
from flytekitplugins.dolt.schema import DoltConfig, DoltTable

from flytekit import task, workflow


@pytest.fixture(scope="module")
def dolt_install():
    """
    Every pytest instance downloads the most recent `dolt` binary,
    sets it as the dolt path in Dolt's Python package, and then deletes
    it afterward.
    """
    tmp_dir = tempfile.TemporaryDirectory()
    dir_path = os.path.dirname(os.path.realpath(__file__))
    install_script = os.path.join(dir_path, "..", "..", "flytekit-dolt", "scripts", "install.sh")

    try:
        dolt_path = os.path.join(tmp_dir.name, "dolt")
        subprocess.run(install_script, env={"INSTALL_PATH": tmp_dir.name})
        for attr, value in [("user.name", "First Last"), ("user.email", "first.last@example.com")]:
            subprocess.run([dolt_path, "config", "--global", "--add", attr, value])
        dolt.set_dolt_path(dolt_path)
        yield dolt_path
    finally:
        shutil.rmtree(tmp_dir.name)
        shutil.rmtree(Path.home() / ".dolt")


@pytest.fixture(scope="function")
def doltdb_path(tmp_path, dolt_install):
    with tempfile.TemporaryDirectory() as d:
        db_path = os.path.join(d, "foo")
        yield db_path


@pytest.fixture(scope="function")
def dolt_config(doltdb_path):
    yield DoltConfig(
        db_path=doltdb_path,
        tablename="foo",
    )


@pytest.fixture(scope="function")
def db(doltdb_path):
    try:
        db = dolt.Dolt.init(doltdb_path)
        db.sql("create table bar (name text primary key, count bigint)")
        db.sql("insert into bar values ('Dilly', 3)")
        db.sql("select dolt_commit('-am', 'Initialize bar table')")
        yield db
    finally:
        pass


def test_dolt_table_write(db, dolt_config):
    @task
    def my_dolt(a: int) -> DoltTable:
        df = pandas.DataFrame([("Alice", a)], columns=["name", "count"])
        return DoltTable(data=df, config=dolt_config)

    @workflow
    def my_wf(a: int) -> DoltTable:
        return my_dolt(a=a)

    x = my_wf(a=5)
    assert x
    assert (x.data == pandas.DataFrame([("Alice", 5)], columns=["name", "count"])).all().all()


def test_dolt_table_read(db, dolt_config):
    @task
    def my_dolt(t: DoltTable) -> str:
        df = t.data
        return df.name.values[0]

    @workflow
    def my_wf(t: DoltTable) -> str:
        return my_dolt(t=t)

    dolt_config.tablename = "bar"
    x = my_wf(t=DoltTable(config=dolt_config))
    assert x == "Dilly"


def test_dolt_table_read_task_config(db, dolt_config):
    @task
    def my_dolt(t: DoltTable) -> str:
        df = t.data
        return df.name.values[0]

    @task
    def my_table() -> DoltTable:
        dolt_config.tablename = "bar"
        t = DoltTable(config=dolt_config)
        return t

    @workflow
    def my_wf() -> str:
        t = my_table()
        return my_dolt(t=t)

    x = my_wf()
    assert x == "Dilly"


def test_dolt_table_read_mixed_config(db, dolt_config):
    @task
    def my_dolt(t: DoltTable) -> str:
        df = t.data
        return df.name.values[0]

    @task
    def my_table(conf: DoltConfig) -> DoltTable:
        return DoltTable(config=conf)

    @workflow
    def my_wf(conf: DoltConfig) -> str:
        t = my_table(conf=conf)
        return my_dolt(t=t)

    dolt_config.tablename = "bar"
    x = my_wf(conf=dolt_config)

    assert x == "Dilly"


def test_dolt_sql_read(db, dolt_config):
    @task
    def my_dolt(t: DoltTable) -> str:
        df = t.data
        return df.name.values[0]

    @workflow
    def my_wf(t: DoltTable) -> str:
        return my_dolt(t=t)

    dolt_config.tablename = None
    dolt_config.sql = "select * from bar"
    x = my_wf(t=DoltTable(config=dolt_config))
    assert x == "Dilly"


def test_branching(db, doltdb_path):
    def generate_confs(a: int) -> typing.Tuple[DoltConfig, DoltConfig, DoltConfig]:
        branch_conf = NewBranch(f"run/a_is_{a}")
        users_conf = DoltConfig(
            db_path=doltdb_path,
            tablename="users",
            branch_conf=branch_conf,
        )

        query_users = DoltTable(
            config=DoltConfig(
                db_path=doltdb_path,
                sql="select * from users where `count` > 5",
                branch_conf=branch_conf,
            ),
        )

        big_users_conf = DoltConfig(
            db_path=doltdb_path,
            tablename="big_users",
            branch_conf=branch_conf,
        )

        return users_conf, query_users, big_users_conf

    @task
    def get_confs(a: int) -> typing.Tuple[DoltConfig, DoltTable, DoltConfig]:
        return generate_confs(a)

    @task
    def populate_users(a: int, conf: DoltConfig) -> DoltTable:
        users = [("George", a), ("Alice", a * 2), ("Stephanie", a * 3)]
        df = pandas.DataFrame(users, columns=["name", "count"])
        return DoltTable(data=df, config=conf)

    @task
    def filter_users(a: int, all_users: DoltTable, filtered_users: DoltTable, conf: DoltConfig) -> DoltTable:
        usernames = filtered_users.data[["name"]]
        return DoltTable(data=usernames, config=conf)

    @task
    def count_users(users: DoltTable) -> int:
        return users.data.shape[0]

    @workflow
    def wf(a: int) -> int:
        user_conf, query_conf, big_user_conf = get_confs(a=a)
        users = populate_users(a=a, conf=user_conf)
        big_users = filter_users(a=a, all_users=users, filtered_users=query_conf, conf=big_user_conf)
        big_user_cnt = count_users(users=big_users)
        return big_user_cnt

    assert wf(a=2) == 1
    assert wf(a=3) == 2

    res = db.sql("select * from big_users as of HASHOF('run/a_is_3')", result_format="csv")
    names = set([x["name"] for x in res])
    assert names == {"Alice", "Stephanie"}

    res = db.sql("select * from big_users as of HASHOF('run/a_is_2')", result_format="csv")
    names = set([x["name"] for x in res])
    assert names == {"Stephanie"}
