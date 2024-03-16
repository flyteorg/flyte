import typing

from flytekit import workflow
from flytekit.configuration import Config
from flytekit.remote.remote import FlyteRemote

r = FlyteRemote(config=Config.auto(), default_project="p1", default_domain="d1")
t1 = r.fetch_task(name="task1", version="tst")
t2 = r.fetch_task(name="task2", version="tst")


@workflow
def hello_wf(a: int) -> typing.Tuple[float, bool]:
    x = t1()
    y = t2()
    y >> x
    return x, y
