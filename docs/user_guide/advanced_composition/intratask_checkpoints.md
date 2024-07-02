# Intratask checkpoints

```{eval-rst}
.. tags:: MachineLearning, Intermediate
```

A checkpoint in Flyte serves to recover a task from a previous failure by preserving the task's state before the failure
and resuming from the latest recorded state.

## Why intratask checkpoints?

The inherent design of Flyte, being a workflow engine, allows users to break down operations, programs or ideas
into smaller tasks within workflows. In the event of a task failure, the workflow doesn't need to rerun the
previously completed tasks. Instead, it can retry the specific task that encountered an issue.
Once the problematic task succeeds, it won't be rerun. Consequently, the natural boundaries between tasks act as implicit checkpoints.

However, there are scenarios where breaking a task into smaller tasks is either challenging or undesirable due to the associated overhead.
This is especially true when running a substantial computation in a tight loop.
In such cases, users may consider splitting each loop iteration into individual tasks using dynamic workflows.
Yet, the overhead of spawning new tasks, recording intermediate results, and reconstructing the state can incur additional expenses.

### Use case: Model training

An exemplary scenario illustrating the utility of intra-task checkpointing is during model training.
In situations where executing multiple epochs or iterations with the same dataset might be time-consuming,
setting task boundaries can incur a high bootstrap time and be costly.

Flyte addresses this challenge by providing a mechanism to checkpoint progress within a task execution,
saving it as a file or set of files. In the event of a failure, the checkpoint file can be re-read to
resume most of the state without rerunning the entire task.
This feature opens up possibilities to leverage alternate, more cost-effective compute systems,
such as [AWS spot instances](https://aws.amazon.com/ec2/spot/),
[GCP pre-emptible instances](https://cloud.google.com/compute/docs/instances/preemptible) and others.

These instances offer great performance at significantly lower price points compared to their on-demand or reserved counterparts.
This becomes feasible when tasks are constructed in a fault-tolerant manner.
For tasks running within a short duration, e.g., less than 10 minutes, the likelihood of failure is negligible,
and task-boundary-based recovery provides substantial fault tolerance for successful completion.

However, as the task execution time increases, the cost of re-running it also increases,
reducing the chances of successful completion. This is precisely where Flyte's intra-task checkpointing proves to be highly beneficial.

Here's an example illustrating how to develop tasks that leverage intra-task checkpointing.
It's important to note that Flyte currently offers the low-level API for checkpointing.
Future integrations aim to incorporate higher-level checkpointing APIs from popular training frameworks
like Keras, PyTorch, Scikit-learn, and big-data frameworks such as Spark and Flink, enhancing their fault-tolerance capabilities.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary libraries and set the number of task retries to `3`:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/checkpoint.py
:caption: advanced_composition/checkpoint.py
:lines: 1-4
```

We define a task to iterate precisely `n_iterations`, checkpoint its state, and recover from simulated failures:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/checkpoint.py
:caption: advanced_composition/checkpoint.py
:pyobject: use_checkpoint
```

The checkpoint system offers additional APIs, documented in the code accessible at
[checkpointer code](https://github.com/flyteorg/flytekit/blob/master/flytekit/core/checkpointer.py).

Create a workflow that invokes the task:
The task will automatically undergo retries in the event of a  {ref}`FlyteRecoverableException <flytekit:exception_handling>`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/checkpoint.py
:caption: advanced_composition/checkpoint.py
:pyobject: checkpointing_example
```

The local checkpoint is not utilized here because retries are not supported:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/checkpoint.py
:caption: advanced_composition/checkpoint.py
:lines: 37-42
```

## Run the example on the Flyte cluster

To run the provided workflow on the Flyte cluster, use the following command:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/checkpoint.py \
  checkpointing_example --n_iterations 10
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
