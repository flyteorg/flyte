import time
from flytekit import task, workflow, LaunchPlan, reference_launch_plan, CronSchedule, ConcurrencyPolicy, ConcurrencyLimitBehavior, Resources

##
# Workflows here are crafted specifically for reproducing
# Before the fix was deployed, all of the starve_* workflows here would
# end up in a situation where the reference launchplan would have completed
# execution but the parent's node for the the launchplan would show up as running
# for an extended duration (until another node completes).
# In addition node layout here triggers evaluations in a sequence that exposes
# some issues inherent with how the max-parallelism is currently implemented in flyte.

# ---------- helper tasks -------------------------------------------------
@task(
    cache=False,
    requests=Resources(cpu="0.25", mem="250Mi"),
    limits=Resources(cpu="0.25", mem="500Mi")
)
def sleep_task(name: str, sleep_seconds: int):
    print(f"[{name}] sleeping {sleep_seconds}s …")
    time.sleep(sleep_seconds)
    print(f"[{name}] done")

@task(
    cache=False,
    requests=Resources(cpu="0.25", mem="250Mi"),
    limits=Resources(cpu="0.25", mem="500Mi")
)
def sleepy(sleep_seconds: int):
    print(f"Sleeping for {sleep_seconds} seconds...")
    time.sleep(sleep_seconds)
    print("Finished sleeping.")


# ---------- reference launch-plans ---------------------------------------
@reference_launch_plan(
    project="shardool-project",
    domain="development",
    name="another_sleepy_lp",
    version="level0-v.12",
)
def lp_fast(sleep_seconds: int): ...

@reference_launch_plan(
    project="shardool-project",
    domain="development",
    name="another_sleepy_lp",
    version="level0-v.12",
)
def lp_slow(sleep_seconds: int): ...

# ---------- Repo Workflow and LaunchPlan with low max_parallelism ----------

@workflow
def starve_lp_parent_max():
    gate = sleep_task(name="gate", sleep_seconds=60)

    # 1) Top Level fast tasks
    fast_tasks = []
    for i in range(1):
        t = sleepy(sleep_seconds=60)
        fast_tasks.append(t)

    # 2) 2nd level slow tasks (fanned out downstream from top level fast)
    slow_tasks_level2 = []
    for i in range(2):
        t = sleepy(sleep_seconds=60*60)
        slow_tasks_level2.append(t)

    # 3) Additional slow tasks to take up parallelism slots
    slow_tasks = []
    for i in range(3):
        t = sleepy(sleep_seconds=60*60)
        slow_tasks.append(t)

    # 4) Fast reference launchplan <- This is where we expect the evaluations to be delayed when the issue reproduces
    fast_lp = lp_fast(sleep_seconds=180)

    # 5) Edge definitions
    for t in fast_tasks:
        gate >> t

    for i, t in enumerate(slow_tasks_level2):
        fast_tasks[i//2] >> t

    for t in slow_tasks:
        gate >> t

    # 6) The fast LP is directly downstream from the gate task.
    gate >> fast_lp

    # downstream of fast_lp, not necessary for reproduction.
    slow_lp = lp_slow(sleep_seconds=600)
    fast_lp >> slow_lp

lp = LaunchPlan.get_or_create(
    name="starve_lp_parent_lp_max",
    workflow=starve_lp_parent_max,
    max_parallelism=5,
    concurrency=ConcurrencyPolicy(
        max_concurrency=5,
        behavior=ConcurrencyLimitBehavior.SKIP),
    schedule=CronSchedule(schedule="*/15 * * * *")
)

@workflow
def starve_lp_parent_fixed():
    gate = sleep_task(name="gate", sleep_seconds=60)

    # 1) Top Level fast task1
    fast_task1 = sleep_task(name="fast_task1", sleep_seconds=60)

    # 2) 2nd level fast tasks (fanned out downstream from top level fast)
    fast_tasks = []
    for i in range(6):
        t = sleep_task(name=f"fast_task_{i}", sleep_seconds=60)
        fast_tasks.append(t)

    slow_lp = lp_slow(sleep_seconds=60*45)

    # 3) Additional slow task to take up parallelism slots
    # This is task is required to reproduce the issue.
    slow_task1 = sleep_task(name="slow_task1", sleep_seconds=60*60)

    # 4) Fast reference launchplan <- This is where we expect the evaluations to be delayed when the issue reproduces
    fast_lp = lp_fast(sleep_seconds=60*7)

    # 5) Edge definitions
    gate >> fast_task1

    for t in fast_tasks:
        fast_task1 >> t

    for t in fast_tasks:
        t >> slow_lp

    gate >> slow_task1
    slow_task1

    # 6) The fast LP is directly downstream from the gate task.
    gate >> fast_lp

lp = LaunchPlan.get_or_create(
    name="starve_lp_parent_lp_fixed",
    workflow=starve_lp_parent_fixed,
    max_parallelism=5,
    concurrency=ConcurrencyPolicy(
        max_concurrency=5,
        behavior=ConcurrencyLimitBehavior.SKIP),
    schedule=CronSchedule(schedule="*/15 * * * *")
)

@workflow
def starve_lp_parent_nested():
    starve_lp_parent_fixed()

lp = LaunchPlan.get_or_create(
    name="starve_lp_parent_lp_nested",
    workflow=starve_lp_parent_nested,
    max_parallelism=5,
    concurrency=ConcurrencyPolicy(
        max_concurrency=5,
        behavior=ConcurrencyLimitBehavior.SKIP),
    schedule=CronSchedule(schedule="*/15 * * * *")
)
