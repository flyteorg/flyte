import datetime
import enum
import os
import typing
from dataclasses import dataclass

import pandas as pd
from dataclasses_json import DataClassJsonMixin
from typing_extensions import Annotated

from flytekit import kwtypes, task, workflow
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile
from flytekit.types.structured import StructuredDataset

superset_cols = kwtypes(name=str, age=int)
subset_cols = kwtypes(age=int)


@task
def get_subset_df(df: Annotated[pd.DataFrame, superset_cols]) -> Annotated[StructuredDataset, subset_cols]:
    df = pd.concat([df, pd.DataFrame([[30]], columns=["age"])])
    return StructuredDataset(dataframe=df)


@task
def show_sd(in_sd: StructuredDataset):
    pd.set_option("expand_frame_repr", False)
    df = in_sd.open(pd.DataFrame).all()
    print(df)


@dataclass
class MyDataclass(DataClassJsonMixin):
    i: int
    a: typing.List[str]


class Color(enum.Enum):
    RED = "RED"
    GREEN = "GREEN"
    BLUE = "BLUE"


@task
def print_all(
    a: int,
    b: str,
    c: float,
    d: MyDataclass,
    e: typing.List[int],
    f: typing.Dict[str, float],
    g: FlyteFile,
    h: bool,
    i: datetime.datetime,
    j: datetime.timedelta,
    k: Color,
    l: dict,
    m: dict,
    n: typing.List[typing.Dict[str, FlyteFile]],
    o: typing.Dict[str, typing.List[FlyteFile]],
    p: typing.Any,
    q: FlyteDirectory,
):
    print(f"{a}, {b}, {c}, {d}, {e}, {f}, {g}, {h}, {i}, {j}, {k}, {l}, {m}, {n}, {o}, {p}, {q}")


@task
def test_union1(a: typing.Union[int, FlyteFile, typing.Dict[str, float], datetime.datetime, Color]):
    print(a)


@task
def test_union2(a: typing.Union[float, typing.List[int], MyDataclass]):
    print(a)


@workflow
def my_wf(
    a: int,
    b: str,
    c: float,
    d: MyDataclass,
    e: typing.List[int],
    f: typing.Dict[str, float],
    g: FlyteFile,
    h: bool,
    i: datetime.datetime,
    j: datetime.timedelta,
    k: Color,
    l: dict,
    n: typing.List[typing.Dict[str, FlyteFile]],
    o: typing.Dict[str, typing.List[FlyteFile]],
    p: typing.Any,
    q: FlyteDirectory,
    remote: pd.DataFrame,
    image: StructuredDataset,
    m: dict = {"hello": "world"},
) -> Annotated[StructuredDataset, subset_cols]:
    x = get_subset_df(df=remote)  # noqa: shown for demonstration; users should use the same types between tasks
    show_sd(in_sd=x)
    show_sd(in_sd=image)
    print_all(a=a, b=b, c=c, d=d, e=e, f=f, g=g, h=h, i=i, j=j, k=k, l=l, m=m, n=n, o=o, p=p, q=q)
    return x


@task
def task_with_optional(a: typing.Optional[str]) -> str:
    return "default" if a is None else a


@workflow
def wf_with_none(a: typing.Optional[str] = None) -> str:
    return task_with_optional(a=a)


@task
def task_with_env_vars(env_vars: typing.List[str]) -> str:
    collated_env_vars = []
    for env_var in env_vars:
        collated_env_vars.append(os.environ[env_var])
    return ",".join(collated_env_vars)


@workflow
def wf_with_env_vars(env_vars: typing.List[str]) -> str:
    return task_with_env_vars(env_vars=env_vars)
