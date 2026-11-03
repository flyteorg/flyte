"""Run pinned flyte-sdk examples against the local devbox."""

import asyncio
import logging
import os
import sys
import tempfile
import time

import flyte
from flyte.models import ActionPhase

RUN_TIMEOUT_SECS = 900


async def submit(task_fn, name, **kwargs):
    run = await flyte.with_runcontext(log_level=logging.DEBUG).run.aio(task_fn, **kwargs)
    print(f"[{name}] submitted run={run.name} url={run.url}", flush=True)
    return run


async def wait_for_result(name, run):
    # run.wait() can return before the run is terminal when the devbox's
    # watch stream ends early, so re-enter it until the action is done.
    deadline = time.monotonic() + RUN_TIMEOUT_SECS
    await run.wait.aio(quiet=True)
    while not run.action.done():
        if time.monotonic() > deadline:
            raise TimeoutError(f"[{name}] run {run.name} not terminal after {RUN_TIMEOUT_SECS}s")
        await asyncio.sleep(5)
        await run.wait.aio(quiet=True)

    # details() may serve a pre-terminal cache, so the watched phase is the
    # source of truth; details are only fetched for the error message.
    phase = run.action.phase
    if phase != ActionPhase.SUCCEEDED:
        detail = await run.action.details()
        message = detail.error_info.message if detail.error_info else "<no error info>"
        raise RuntimeError(f"[{name}] finished in phase {phase}: {message}")
    print(f"[{name}] succeeded", flush=True)


async def main(upload_file):
    # Import every example before the first submit: the SDK caches the code
    # bundle of loaded modules, so late imports would be missing from it.
    from examples.basics.file_local import process_file
    from examples.basics.hello import main as hello_main
    from examples.basics.no_outputs import no_outputs_task

    # The example manifest: (name, entry task, inputs). Add a line (and the
    # matching import above) to cover another example.
    examples = [
        # Control plane smoke: task->task calls and flyte.map fanout.
        ("basics/hello", hello_main, {"x_list": list(range(1, 11))}),
        # No-output edge case.
        ("basics/no_outputs", no_outputs_task, {}),
        # Data plane: File upload via the object store, read inside the task.
        ("basics/file_local", process_file, {"file": upload_file}),
    ]

    # Submit everything up front, then wait in parallel: total wall-clock is
    # the slowest run, not the sum (pod scheduling and image pulls dominate).
    runs = [(name, await submit(task_fn, name, **kwargs)) for name, task_fn, kwargs in examples]
    results = await asyncio.gather(
        *(wait_for_result(name, run) for name, run in runs),
        return_exceptions=True,
    )

    failures = [r for r in results if isinstance(r, BaseException)]
    for failure in failures:
        print(f"FAILED: {failure}", flush=True)
    if failures:
        raise SystemExit(1)
    print(f"All {len(runs)} example runs succeeded", flush=True)


if __name__ == "__main__":
    os.chdir(sys.argv[1])
    sys.path.insert(0, os.getcwd())

    flyte.init_from_config()

    from flyte.io import File

    with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as tmp:
        tmp.write("Hello from devbox CI!")
        tmp_path = tmp.name
    uploaded = File.from_local_sync(tmp_path)

    asyncio.run(main(uploaded))
