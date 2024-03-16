import asyncio
import os
import typing
from functools import partial
from pathlib import Path

import pandas as pd

from flytekit import task, workflow
from flytekit.configuration import Config
from flytekit.experimental import EagerException, eager
from flytekit.remote import FlyteRemote
from flytekit.types.directory import FlyteDirectory
from flytekit.types.file import FlyteFile
from flytekit.types.structured import StructuredDataset

remote = FlyteRemote(
    config=Config.for_sandbox(),
    default_project="flytesnacks",
    default_domain="development",
)


eager_partial = partial(eager, remote=remote)


@task
def add_one(x: int) -> int:
    return x + 1


@task
def double(x: int) -> int:
    return x * 2


@task
def gt_0(x: int) -> bool:
    return x > 0


@task
def raises_exc(x: int) -> int:
    if x == 0:
        raise TypeError
    return x


@task
def create_structured_dataset() -> StructuredDataset:
    df = pd.DataFrame({"a": [1, 2, 3]})
    return StructuredDataset(dataframe=df)


@task
def create_file() -> FlyteFile:
    fname = "/tmp/flytekit_test_file"
    with open(fname, "w") as fh:
        fh.write("some data\n")
    return FlyteFile(path=fname)


@task
def create_directory() -> FlyteDirectory:
    dirname = "/tmp/flytekit_test_dir"
    Path(dirname).mkdir(exist_ok=True, parents=True)
    with open(os.path.join(dirname, "file"), "w") as tmp:
        tmp.write("some data\n")
    return FlyteDirectory(path=dirname)


@eager_partial
async def simple_eager_wf(x: int) -> int:
    out = await add_one(x=x)
    return await double(x=out)


@eager_partial
async def conditional_eager_wf(x: int) -> int:
    if await gt_0(x=x):
        return -1
    return 1


@eager_partial
async def try_except_eager_wf(x: int) -> int:
    try:
        return await raises_exc(x=x)
    except EagerException:
        return -1


@eager_partial
async def gather_eager_wf(x: int) -> typing.List[int]:
    results = await asyncio.gather(*[add_one(x=x) for _ in range(10)])
    return results


@eager_partial
async def nested_eager_wf(x: int) -> int:
    out = await simple_eager_wf(x=x)
    return await double(x=out)


@workflow
def wf_with_eager_wf(x: int) -> int:
    out = simple_eager_wf(x=x)
    return double(x=out)


@workflow
def subworkflow(x: int) -> int:
    return add_one(x=x)


@eager_partial
async def eager_wf_with_subworkflow(x: int) -> int:
    out = await subworkflow(x=x)
    return await double(x=out)


@eager_partial
async def eager_wf_structured_dataset() -> int:
    dataset = await create_structured_dataset()
    df = dataset.open(pd.DataFrame).all()
    return int(df["a"].sum())


@eager_partial
async def eager_wf_flyte_file() -> str:
    file = await create_file()
    file.download()
    with open(file.path) as f:
        data = f.read().strip()
    return data


@eager_partial
async def eager_wf_flyte_directory() -> str:
    directory = await create_directory()
    directory.download()
    with open(os.path.join(directory.path, "file")) as f:
        data = f.read().strip()
    return data


@eager(remote=remote, local_entrypoint=True)
async def eager_wf_local_entrypoint(x: int) -> int:
    out = await add_one(x=x)
    return await double(x=out)


if __name__ == "__main__":
    print(asyncio.run(simple_eager_wf(x=1)))
