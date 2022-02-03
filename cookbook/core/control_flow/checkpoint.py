"""
Intratask Checkpoints
------------

.. note::

  This feature is available since flytekit version 0.30.0b6+ and needs a Flyte backend version of atleast 0.19.0+

A checkpoint allows a task to recover from a previous failure by recording the state of a task before the failure and
resuming from the latest recorded state.

Why Intra-task checkpoints?
^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Flyte at its core is a workflow engine, and workflows provide a way to logically break up an operation / program / idea
into smaller sized tasks. If a task fails, the workflow does not need to run the previously completed tasks, but can
simply retry the task at hand. Eventually, when the task succeeds, it will never be run again. Thus, task boundaries
naturally serve as checkpoints.

There are cases where it is not easy or desirable to break a task into further smaller tasks, as running a task
adds overhead. This is especially true when running a large computation in a tight-loop. This these cases, users can
split each loop iteration into its own task using dynamic workflows, but the overhead of spawning new tasks, recording
intermediate results, and reconstituting the state can be extremely expensive.

Model-training Use Case
==================

An example of this case is model training. Running multiple epochs or different iterations with the same
dataset can take a long time, but the bootstrap time may be high and creating task boundaries can be expensive.

To tackle this, Flyte offers a way to checkpoint progress within a task execution as a file or a set of files. These
checkpoints can be written synchronously or asynchronously. In case of failure, the checkpoint file can be re-read to resume
most of the state without re-running the entire task. This opens up the opportunity to use alternate compute systems with
lower guarantees like - `AWS Spot Instances <https://aws.amazon.com/ec2/spot/>`__, `GCP Pre-emptible Instances <https://cloud.google.com/compute/docs/instances/preemptible>`__ etc.

These instances offer great performance at much lower price-points as compared to their on-demand or reserved alternatives.
This is possible if you construct your tasks in a fault-tolerant way. For most cases, when the task runs for a short duration,
e.g. less than 10 minutes, the potential of failure is not significant and task-boundary-based recovery offers
significant fault-tolerance to ensure successful completion.

But, as the time for a task increases, the cost of re-running the task increases and reduces the chance of successful
completion. This is where Flyte's intra-task checkpointing truly shines. This document provides an example of how to
develop tasks which utilize intra-task checkpointing. This only provides the low-level API. We intend to integrate
higher level checkpointing API's available in popular training frameworks like Keras, Pytorch, Scikit learn and
big-Data frameworks like Spark, Flink to super charge their fault-tolerance.
"""

from flytekit import task, workflow, current_context
from flytekit.exceptions.user import FlyteRecoverableException


RETRIES=3


# %%
# This task shows how in case of failure checkpoints can help resume. This is only an example task and shows the api for
# the checkpointer. The checkpoint system exposes other api's and for a more detailed understanding refer to
# :ref:py:class:`flytekit.Checkpoint`
#
# The goal of this method is to actually return `a+4` It does this in 3 retries of the task, by recovering from previous
# failures. For each failure it increments the value by 1
@task(retries=RETRIES)
def use_checkpoint(n_iterations: int) -> int:
    cp = current_context().checkpoint
    prev = cp.read()
    start = 0
    if prev:
        start = int(prev.decode())

    # We will create a failure interval so that we can create failures ever n iterations and then succeed within
    # configured retries
    failure_interval = n_iterations * 1.0 / RETRIES
    i = 0
    for i in range(start, n_iterations):
        # simulate a deterministic failure, for demonstration. We want to show how it eventually completes within
        # the given retries
        if i > start and i % failure_interval == 0:
            raise FlyteRecoverableException(f"Failed at iteration {start}, failure_interval {failure_interval}")
        # save progress state. It is also entirely possible save state every few intervals.
        cp.write(f"{i + 1}".encode())

    return i


# %%
# The workflow in this case simply calls the task. The task itself will be retried for the failure :ref:py:class:`FlyteRecoverableException`
@workflow
def example(n_iterations: int) -> int:
    return use_checkpoint(n_iterations=n_iterations)


#%%
# Locally the checkpoint is stored, but since, retries are not supported, the checkpoint is not really used.
if __name__ == "__main__":
    try:
        example(n_iterations=10)
    except RuntimeError as e:
        # Locally an exception is expected as retries are not performed.
        pass
