"""Submits a tiny task to a running devbox and asserts it reaches SUCCEEDED.

Used by .github/workflows/flyte-binary-v2.yml as the post-build integration
gate: if this script exits non-zero, the just-built devbox image is broken.

The SDK reads connection info from $HOME/.flyte/config.yaml. Storage for
fast-registration uploads is handled server-side by the devbox's DataProxy +
rustfs, so the SDK doesn't need explicit S3 credentials. The worker image
comes from $FLYTE_WORKER_IMAGE so CI can pre-pull it into k3s before
submission and keep this script aligned with whatever tag the workflow
loaded.
"""
import os
import sys

import flyte

WORKER_IMAGE = os.environ["FLYTE_WORKER_IMAGE"]

env = flyte.TaskEnvironment(
    name="devbox_ci_smoke",
    image=WORKER_IMAGE,
)


@env.task
def add_one(x: int) -> int:
    return x + 1


def main() -> int:
    flyte.init_from_config()
    run = flyte.run(add_one, x=41)
    print(f"run.result={run.result!r}")
    if run.result != 42:
        print(f"FAIL: expected 42, got {run.result!r}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
