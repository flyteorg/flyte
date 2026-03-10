import asyncio
import signal
import sys
from typing import List
import os

import flyte
from write_run_metadata import write_run_metadata

# Maximum time (seconds) to wait for the run to reach a terminal state.
RUN_TIMEOUT_SECONDS = int(os.getenv("RUN_TIMEOUT_SECONDS", "600"))

env_name ="hello_world"

env = flyte.TaskEnvironment(name=env_name)

logs = "This is a sample log"


@env.task
async def double(x: int) -> int:
    return x * 2


@env.task
async def root_wf(x: int) -> List[int]:
    print(x)
    vals = []
    with flyte.group("double-list-1"):
        for x in range(x):
            vals.append(double(x))

        o1 = await asyncio.gather(*vals)

    vals = []
    with flyte.group("double-list-2"):
        for x in range(x):
            vals.append(double(x))

        o2 = await asyncio.gather(*vals)

    return o1 + o2


@env.task
async def group_dynamic(x: int) -> List[int]:
    print(logs)
    vals = {"even": [], "odd": []}
    for x in range(x):
        if x % 2 == 0:
            group = "even"
        else:
            group = "odd"
        vals[group].append(double(x))

    with flyte.group("even-group"):
        o1 = await asyncio.gather(*vals["even"])
    with flyte.group("odd-group"):
        o2 = await asyncio.gather(*vals["odd"])

    # ensure this runs longs enough to produce metrics
    await asyncio.sleep(25)

    return o1 + o2


if __name__ == "__main__":
    api_key = os.getenv('UNION_API_KEY')
    if api_key is None:
      raise Exception('UNION_API_KEY has not been provided')
    # These values (org, project, domain) can be updated in buildkite
    org = os.getenv('TEST_ORG', 'playground')
    test_project = os.getenv('TEST_PROJECT', 'e2e')
    test_domain = os.getenv('TEST_DOMAIN', 'development')

    flyte.init(api_key=api_key, org=org, project=test_project, domain=test_domain, image_builder="remote")
    run = flyte.with_runcontext().run(group_dynamic, x=10)
    print(run.name)
    print(run.url)

    # Wait for the run to finish so downstream e2e tests don't race against
    # tasks that are still queueing / waiting for resources.
    # run.wait() blocks and polls the server for status updates internally.
    # We wrap it with SIGALRM to enforce a timeout since wait() has no
    # timeout parameter.
    def _timeout_handler(signum, frame):
        print(
            f"ERROR: Run {run.name} did not complete within {RUN_TIMEOUT_SECONDS}s",
            file=sys.stderr,
        )
        sys.exit(1)

    signal.signal(signal.SIGALRM, _timeout_handler)
    signal.alarm(RUN_TIMEOUT_SECONDS)

    print(f"Waiting for run {run.name} to complete (timeout={RUN_TIMEOUT_SECONDS}s)…")
    run.wait(quiet=True)
    signal.alarm(0)  # cancel the alarm

    print(f"Run completed — phase={run.action.phase}")

    write_run_metadata({
        "env": env_name,
        "name": "run_with_groups",
        "id": run.name,
        "url": run.url,
        "display_name": "group_dynamic",
        "inputs": 10,
        "logs": logs
    })
