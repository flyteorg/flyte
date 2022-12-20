"""
Intratask Checkpoints
---------------------

.. tags:: MachineLearning, Intermediate

.. note::

  This feature is available from Flytekit version 0.30.0b6+ and needs a Flyte backend version of at least 0.19.0+.

A checkpoint recovers a task from a previous failure by recording the state of a task before the failure and
resuming from the latest recorded state.

Why Intra-task Checkpoints?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Flyte, at its core, is a workflow engine. Workflows provide a way to break up an operation/program/idea
logically into smaller tasks. If a task fails, the workflow does not need to run the previously completed tasks. It can
simply retry the task that failed. Eventually, when the task succeeds, it will not run again. Thus, task boundaries
naturally serve as checkpoints.

There are cases where it is not easy or desirable to break a task into smaller tasks, because running a task
adds to the overhead. This is true when running a large computation in a tight-loop. In such cases, users can
split each loop iteration into its own task using :ref:`dynamic workflows <Dynamic Workflows>`, but the overhead of spawning new tasks, recording
intermediate results, and reconstituting the state can be expensive.

Model-training Use Case
=======================

An example of this case is model training. Running multiple epochs or different iterations with the same
dataset can take a long time, but the bootstrap time may be high and creating task boundaries can be expensive.

To tackle this, Flyte offers a way to checkpoint progress within a task execution as a file or a set of files. These
checkpoints can be written synchronously or asynchronously. In case of failure, the checkpoint file can be re-read to resume
most of the state without re-running the entire task. This opens up the opportunity to use alternate compute systems with
lower guarantees like `AWS Spot Instances <https://aws.amazon.com/ec2/spot/>`__, `GCP Pre-emptible Instances <https://cloud.google.com/compute/docs/instances/preemptible>`__, etc.

These instances offer great performance at much lower price-points as compared to their on-demand or reserved alternatives.
This is possible if you construct the tasks in a fault-tolerant manner. In most cases, when the task runs for a short duration,
e.g., less than 10 minutes, the potential of failure is insignificant and task-boundary-based recovery offers
significant fault-tolerance to ensure successful completion.

But as the time for a task increases, the cost of re-running it increases, and reduces the chances of successful
completion. This is where Flyte's intra-task checkpointing truly shines.

Let's look at an example of how to develop tasks which utilize intra-task checkpointing. It only provides the low-level API, though. We intend to integrate
higher-level checkpointing APIs available in popular training frameworks like Keras, Pytorch, Scikit-learn, and
big-data frameworks like Spark and Flink to supercharge their fault-tolerance.
"""

from flytekit import current_context, task, workflow
from flytekit.exceptions.user import FlyteRecoverableException

RETRIES = 3


# %%
# This task shows how checkpoints can help resume execution in case of a failure. This is an example task and shows the API for
# the checkpointer. The checkpoint system exposes other APIs. For a detailed understanding, refer to the `checkpointer code <https://github.com/flyteorg/flytekit/blob/master/flytekit/core/checkpointer.py>`__.
#
# The goal of this method is to loop for exactly n_iterations, checkpointing state and recovering from simualted failures.
@task(retries=RETRIES)
def use_checkpoint(n_iterations: int) -> int:
    cp = current_context().checkpoint
    prev = cp.read()
    start = 0
    if prev:
        start = int(prev.decode())

    # create a failure interval so we can create failures for across 'n' iterations and then succeed after
    # configured retries
    failure_interval = n_iterations // RETRIES
    i = 0
    for i in range(start, n_iterations):
        # simulate a deterministic failure, for demonstration. We want to show how it eventually completes within
        # the given retries
        if i > start and i % failure_interval == 0:
            raise FlyteRecoverableException(
                f"Failed at iteration {i}, failure_interval {failure_interval}"
            )
        # save progress state. It is also entirely possible save state every few intervals.
        cp.write(f"{i + 1}".encode())

    return i


# %%
# The workflow here simply calls the task. The task itself
# will be retried for the :ref:`FlyteRecoverableException <flytekit:exception_handling>`.
#
@workflow
def example(n_iterations: int) -> int:
    return use_checkpoint(n_iterations=n_iterations)


# %%
# The checkpoint is stored locally, but it is not used since retries are not supported.
if __name__ == "__main__":
    try:
        example(n_iterations=10)
    except RuntimeError as e:  # noqa : F841
        # no retries are performed, so an exception is expected when run locally.
        pass
