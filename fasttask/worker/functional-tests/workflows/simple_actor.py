import os

from flytekit import workflow

from union.actor import ActorEnvironment

actor = ActorEnvironment(
    name="simple-actor",
    replica_count=1,
    ttl_seconds=100,
    container_image=os.getenv("UNION_RUNTIME_TEST_IMAGE"),
)


@actor.task
def add_one(num: int) -> int:
    return num + 1


@workflow
def add_five(num: int) -> int:
    return add_one(num=add_one(num=add_one(num=add_one(num=add_one(num=num)))))
