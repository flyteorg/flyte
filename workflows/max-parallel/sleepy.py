import time

from flytekit import LaunchPlan, task, workflow, Resources

# For use as a reference workflow that blocks for the specified duration
# Example registration:

@task(
    cache=False,
    requests=Resources(cpu="0.25", mem="250Mi"),
    limits=Resources(cpu="0.25", mem="500Mi")
)
def another_sleepy_task(sleep_seconds: int = 1):
    print(f"Sleeping for {sleep_seconds} seconds...")
    time.sleep(sleep_seconds)
    print("Finished sleeping.")

@workflow
def another_sleepy_wf(sleep_seconds: int):
    another_sleepy_task(sleep_seconds=sleep_seconds)

lp = LaunchPlan.get_or_create(
    name="another_sleepy_lp",
    workflow=another_sleepy_wf,
)
