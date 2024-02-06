---
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

(getting_started_tasks_and_workflows)=

# Tasks, Workflows and LaunchPlans

In {doc}`"Getting started with workflow development" <../getting_started_with_workflow_development/index>`, we got a basic sense
of how Flyte works by creating and running a workflow made up of a few tasks.
In this guide, you'll learn more about how tasks and workflows fit into the Flyte
programming model.

## Tasks

Flyte tasks are the core building blocks of larger, more complex workflows.

### Tasks are Containerized Blocks of Compute

You can think of a Flyte task as a containerized block of compute. When a task
runs on a Flyte backend, it's isolated within its own container, separate from
all other tasks. Consider this simple one:

```{code-cell} ipython3
from typing import List
from flytekit import task

@task
def mean(values: List[float]) -> float:
    return sum(values) / len(values)
```

As you can see, a task is just a regular Python function that's decorated
with {py:func}`@task <flytekit.task>`. We can run this function just like any
other Python function:

```{code-cell} ipython3
mean(values=[float(i) for i in range(1, 11)])
```

```{important}
There are three important things to note here:

- Most of the Flyte tasks you'll ever write can be executed locally.
- Tasks and workflows must be invoked with keyword arguments.
- When a task runs on a Flyte cluster, it runs on a
  [Kubernetes Pod](https://kubernetes.io/docs/concepts/workloads/pods/), where
  Flyte orchestrates what task to run at what time in the context of a workflow.
```

### Tasks are Strongly Typed

You might also notice that the `mean` function signature is type-annotated with
Python type hints. Flyte uses these annotations to check the input and output
types of the task when it's compiled or invoked.

Under the hood, Flyte uses its own type system that translates values to and from
Flyte types and the SDK language types, in this case Python. The Flyte type
system uses Python type annotations to make sure that the data passing through
tasks and workflows are compatible with the explicitly stated types that we
define through a function signature.

So if we call the `mean` function with the wrong types, we get an error:

```{code-cell} ipython3
try:
    mean(values="hi")
except Exception as e:
    print(e)
```

This may not seem like much for this simple example, but as you start dealing
with more complex data types and pipelines, Flyte's type system becomes
invaluable for catching bugs early.

Flyte's type system is also used for caching, data lineage tracking, and
automatic serialization and deserialization of data as it's passed from one task
to another. You can learn more about it in the {ref}`User Guide <data_types_and_io>`.

## Workflows

Workflows compose multiple tasks – or other workflows – into meaningful steps
of computation to produce some useful set of outputs or outcomes.

Suppose the `mean` task is just one building block of a larger computation.
This is where Flyte workflows can help us manage the added complexity.

### Workflows Build Execution Graphs

Suppose that we want to mean-center and standard-deviation-scale a set of
values. In addition to a `mean` function, we also need to compute standard
deviation and implement the centering and scaling logic.

Let's go ahead and implement those as tasks:

```{code-cell} ipython3
from math import sqrt
from flytekit import workflow


@task
def standard_deviation(values: List[float], mu: float) -> float:
    variance = sum([(x - mu) ** 2 for x in values])
    return sqrt(variance)

@task
def standard_scale(values: List[float], mu: float, sigma: float) -> List[float]:
    return [(x - mu) / sigma for x in values]
```

Then we put all the pieces together into a workflow, which is a function
that's decorated with {py:func}`@workflow <flytekit.workflow>`:

```{code-cell} ipython3
@workflow
def standard_scale_workflow(values: List[float]) -> List[float]:
    mu = mean(values=values)
    sigma = standard_deviation(values=values, mu=mu)
    return standard_scale(values=values, mu=mu, sigma=sigma)
```

Just like tasks, workflows are executable in a regular Python runtime:

```{code-cell} ipython3
standard_scale_workflow(values=[float(i) for i in range(1, 11)])
```

(workflows_versus_task_syntax)=

### Workflows versus Tasks Under the Hood

Although Flyte workflow syntax looks like Python code, it's actually a
[domain-specific language (DSL)](https://en.wikipedia.org/wiki/Domain-specific_language)
for building execution graphs where tasks – and other workflows – serve as the
building blocks.

This means that the workflow function body only supports a subset of Python's
semantics:

- In workflows, you shouldn't use non-deterministic operations like
  `rand.random`, `time.now()`, etc. These functions will be invoked at compile
  time and your workflows will not behave as you expect them to.
- Within workflows, the inputs of workflow and the outputs of tasks function are promises under the hood,
  so you can't access and operate on them like typical Python function outputs.
  _You can only pass promises into tasks, workflows, and other Flyte constructs_.
- Regular Python conditionals won't work as intended in workflows: you need to
  use the {ref}`conditional <conditional>` construct.

In contrast to workflow code, the code within tasks is actually executed by a
Python interpreter when it's run locally or inside a container when run on a
Flyte cluster.

### Workflows Deal with Promises

A promise is essentially a placeholder for a value that hasn't been
materialized yet. To show you what this means concretely, let's re-define
the workflow above but let's also print the output of one of the tasks:

```{code-cell} ipython3
@workflow
def standard_scale_workflow_with_print(values: List[float]) -> List[float]:
    mu = mean(values=values)
    print(mu)  # this is not the actual float value!
    sigma = standard_deviation(values=values, mu=mu)
    return standard_scale(values=values, mu=mu, sigma=sigma)
```

We didn't even execute the workflow and we're already seeing the value of `mu`,
which is a promise. So what's happening here?

When we decorate `standard_scale_workflow_with_print` with `@workflow`, Flyte
compiles an execution graph that's defined inside the function body, so
_it doesn't actually run the computations yet_. Therefore, when Flyte compiles a
workflow, the outputs of task calls are actually promises and not regular python
values.

### Workflows are Strongly Typed Too

Since both tasks and workflows are strongly typed, Flyte can actually catch
type errors! When we learn more about packaging and registering in the next few
guides, we'll see that Flyte can also catch compile-time errors even before
you running any code!

For now, however, we can run the workflow locally to see that we'll get an
error if we introduce a bug in the `standard_scale` task.

```{code-cell} ipython3
@task
def buggy_standard_scale(values: List[float], mu: float, sigma: float) -> float:
    """
    🐞 The implementation and output type of this task is incorrect! It should
    be List[float] instead of a sum of all the scaled values.
    """
    return sum([(x - mu) / sigma for x in values])

@workflow
def buggy_standard_scale_workflow(values: List[float]) -> List[float]:
    mu = mean(values=values)
    sigma = standard_deviation(values=values, mu=mu)
    return buggy_standard_scale(values=values, mu=mu, sigma=sigma)

try:
    buggy_standard_scale_workflow(values=[float(i) for i in range(1, 11)])
except Exception as e:
    print(e)
```

### Workflows can be Embedded in Other Workflows

When a workflow uses another workflow as part of the execution graph, we call
the inner workflow a **subworkflow**. Subworkflows are strongly typed and can
be invoked just like tasks when defining the outer workflow.

For example, we can embed `standard_scale_workflow` inside
`workflow_with_subworkflow`, which uses a `generate_data` task to supply the
data for scaling:

```{code-cell} ipython3
import random

@task
def generate_data(num_samples: int, seed: int) -> List[float]:
    random.seed(seed)
    return [random.random() for _ in range(num_samples)]

@workflow
def workflow_with_subworkflow(num_samples: int, seed: int) -> List[float]:
    data = generate_data(num_samples=num_samples, seed=seed)
    return standard_scale_workflow(values=data)

workflow_with_subworkflow(num_samples=10, seed=3)
```

```{important}
Learn more about subworkflows in the {ref}`User Guide <subworkflow>`.
```

### Specifying Dependencies without Passing Data

You can also specify dependencies between tasks and subworkflows without passing
data from the upstream entity to the downstream entity using the `>>` right shift
operator:

```{code-block} python
@workflow
def wf():
    promise1 = task1()
    promise2 = task2()
    promise3 = subworkflow()
    promise1 >> promise2
    promise2 >> promise3
```

In this workflow, `task1` will execute before `task2`, but it won't pass any of
its data to `task2`. Similarly, `task2` will execute before `subworkflow`.

```{important}
Learn more about chaining flyte entities in the {ref}`User Guide <chain_flyte_entities>`.
```

## Launch plans

A Flyte {py:class}`~flytekit.LaunchPlan` is a partial or complete binding of
inputs necessary to launch a workflow. You can think of it like
the {py:func}`~functools.partial` function in the Python standard library where
you can define default (overridable) and fixed (non-overridable) inputs.

```{note}
Additionally, `LaunchPlan`s provides an interface for specifying run-time
overrides such as notifications, schedules, and more.
```

Create a launch plan like so:

```{code-cell} ipython3
from flytekit import LaunchPlan

standard_scale_launch_plan = LaunchPlan.get_or_create(
    standard_scale_workflow,
    name="standard_scale_lp",
    default_inputs={"values": [3.0, 4.0, 5.0]}
)
```

### Invoking LaunchPlans Locally

You can run a `LaunchPlan` locally. This is, using the local Python interpreter (REPL). It will use the `default_inputs` dictionary
whenever it's invoked:

```{code-cell} ipython3
standard_scale_launch_plan()
```

Of course, these defaults can be overridden:

```{code-cell} ipython3
standard_scale_launch_plan(values=[float(x) for x in range(20, 30)])
```

Later, you'll learn how to run a launch plan on a `cron` schedule, but for
the time being you can think of them as a way for you to templatize workflows
for some set of related use cases, such as model training with a fixed dataset
for reproducibility purposes.

### LaunchPlans can be Embedded in Workflows

Similar to subworkflows, launch plans can be used in a workflow definition:

```{code-cell} ipython3
@workflow
def workflow_with_launchplan(num_samples: int, seed: int) -> List[float]:
    data = generate_data(num_samples=num_samples, seed=seed)
    return standard_scale_launch_plan(values=data)

workflow_with_launchplan(num_samples=10, seed=3)
```

The main difference between subworkflows and launch plans invoked in workflows is
that the latter will kick off a new workflow execution on the Flyte cluster with
its own execution name, while the former will execute the workflow in the
context of the parent workflow's execution context.

```{important}
Learn more about subworkflows in the {ref}`User Guide <launch_plan>`.
```

## What's Next?

So far we've been working with small code snippets and self-contained scripts.
Next, we'll see how to organize a Flyte project that follows software
engineering best practices, including organizing code into meaningful modules, defining third-party dependencies, and creating a container image for making our workflows reproducible.
