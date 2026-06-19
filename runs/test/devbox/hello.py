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
    try:
        flyte.init_from_config()
        run = flyte.run(add_one, x=41)
    except Exception as e:
        # The SDK wraps storage errors with a generic message; walk the chain
        # so CI logs show the real cause (network, signing, etc.).
        cur, depth = e, 0
        while cur is not None and depth < 10:
            print(f"  [{depth}] {type(cur).__name__}: {cur}", file=sys.stderr)
            cur = cur.__cause__ or cur.__context__
            depth += 1
        raise
    print(f"run.result={run.result!r}")
    if run.result != 42:
        print(f"FAIL: expected 42, got {run.result!r}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
