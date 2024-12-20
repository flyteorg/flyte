from flytekit import task, workflow
from flytekit.image_spec import ImageSpec

image_spec = ImageSpec(packages=["numpy", "pandas"], apt_packages=["git"], registry="", builder="test")


@task(container_image=image_spec)
def t2() -> str:
    return "flyte"


@task(container_image=image_spec)
def t1() -> str:
    return "flyte"


@workflow
def wf():
    t1()
    t2()
