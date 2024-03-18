---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

# Waiting for external inputs

*New in Flyte 1.3.0*

There are use cases where you may want a workflow execution to pause, only to continue
when some time has passed or when it receives some inputs that are external to
the workflow execution inputs. You can think of these as execution-time inputs,
since they need to be supplied to the workflow after it's launched. Examples of
this use case would be:

1. **Model Deployment**: A hyperparameter-tuning workflow that
   trains `n` models, where a human needs to inspect a report before approving
   the model for downstream deployment to some serving layer.
2. **Data Labeling**: A workflow that iterates through an image dataset,
   presenting individual images to a human annotator for them to label.
3. **Active Learning**: An [active learning](<https://en.wikipedia.org/wiki/Active_learning_(machine_learning)>)
   workflow that trains a model, shows examples for a human annotator to label
   based which examples it's least/most certain about or would provide the most
   information to the model.

These use cases can be achieved in Flyte with the {func}`~flytekit.sleep`,
{func}`~flytekit.wait_for_input`, and {func}`~flytekit.approve` workflow nodes.
Although all of the examples above are human-in-the-loop processes, these
constructs allow you to pass inputs into a workflow from some arbitrary external
process (👩 human or 🤖 machine) in order to continue.

:::{important}
These functions can only be used inside {func}`@workflow <flytekit.workflow>`-decorated
functions, {func}`@dynamic <flytekit.dynamic>`-decorated functions, or
{ref}`imperative workflows <imperative_workflow>`.
:::

## Pause executions with the `sleep` node

The simplest case is when you want your workflow to {py:func}`~flytekit.sleep`
for some specified amount of time before continuing.

Though this type of node may not be used often in a production setting,
you might want to use it, for example, if you want to simulate a delay in
your workflow to mock out the behavior of some long-running computation.

```{code-cell}
from datetime import timedelta

from flytekit import sleep, task, workflow


@task
def long_running_computation(num: int) -> int:
    """A mock task pretending to be a long-running computation."""
    return num


@workflow
def sleep_wf(num: int) -> int:
    """Simulate a "long-running" computation with sleep."""

    # increase the sleep duration to actually make it long-running
    sleeping = sleep(timedelta(seconds=10))
    result = long_running_computation(num=num)
    sleeping >> result
    return result
```

+++ {"lines_to_next_cell": 0}

As you can see above, we define a simple `add_one` task and a `sleep_wf`
workflow. We first create a `sleeping` and `result` node, then
order the dependencies with the `>>` operator such that the workflow sleeps
for 10 seconds before kicking off the `result` computation. Finally, we
return the `result`.

:::{note}
You can learn more about the `>>` chaining operator
{ref}`here <chain_flyte_entities>`.
:::

Now that you have a general sense of how this works, let's move onto the
{func}`~flytekit.wait_for_input` workflow node.

## Supply external inputs with `wait_for_input`

With the {py:func}`~flytekit.wait_for_input` node, you can pause a
workflow execution that requires some external input signal. For example,
suppose that you have a workflow that publishes an automated analytics report,
but before publishing it you want to give it a custom title. You can achieve
this by defining a `wait_for_input` node that takes a `str` input and
finalizes the report:

```{code-cell}
import typing

from flytekit import wait_for_input


@task
def create_report(data: typing.List[float]) -> dict:  # o0
    """A toy report task."""
    return {
        "mean": sum(data) / len(data),
        "length": len(data),
        "max": max(data),
        "min": min(data),
    }


@task
def finalize_report(report: dict, title: str) -> dict:
    return {"title": title, **report}


@workflow
def reporting_wf(data: typing.List[float]) -> dict:
    report = create_report(data=data)
    title_input = wait_for_input("title", timeout=timedelta(hours=1), expected_type=str)
    return finalize_report(report=report, title=title_input)
```

Let's breakdown what's happening in the code above:

- In `reporting_wf` we first create the raw `report`
- Then, we define a `title` node that will wait for a string to be provided
  through the Flyte API, which can be done through the Flyte UI or through
  `FlyteRemote` (more on that later). This node will time out after 1 hour.
- Finally, we pass the `title_input` promise into `finalize_report`, which
  attaches the custom title to the report.

:::{note}
The `create_report` task is just toy example. In a realistic example, this
report might be an html file or set of visualizations. This can be rendered
in the Flyte UI with {ref}`Flyte Decks <decks>`.
:::

As mentioned in the beginning of this page, this construct can be used for
selecting the best-performing model in cases where there isn't a clear single
metric to determine the best model, or if you're doing data labeling using
a Flyte workflow.

## Continue executions with `approve`

Finally, the {py:func}`~flytekit.approve` workflow node allows you to wait on
an explicit approval signal before continuing execution. Going back to our
report-publishing use case, suppose that we want to block the publishing of
a report for some reason (e.g. if they don't appear to be valid):

```{code-cell}
from flytekit import approve


@workflow
def reporting_with_approval_wf(data: typing.List[float]) -> dict:
    report = create_report(data=data)
    title_input = wait_for_input("title", timeout=timedelta(hours=1), expected_type=str)
    final_report = finalize_report(report=report, title=title_input)

    # approve the final report, where the output of approve is the final_report
    # dictionary.
    return approve(final_report, "approve-final-report", timeout=timedelta(hours=2))
```

+++ {"lines_to_next_cell": 0}

The `approve` node will pass the `final_report` promise through as the
output of the workflow, provided that the `approve-final-report` gets an
approval input via the Flyte UI or Flyte API.

You can also use the output of the `approve` function as a promise, feeding
it to a subsequent task. Let's create a version of our report-publishing
workflow where the approval happens after `create_report`:

```{code-cell}
@workflow
def approval_as_promise_wf(data: typing.List[float]) -> dict:
    report = create_report(data=data)
    title_input = wait_for_input("title", timeout=timedelta(hours=1), expected_type=str)

    # wait for report to run so that the user can view it before adding a custom
    # title to the report
    report >> title_input

    final_report = finalize_report(
        report=approve(report, "raw-report-approval", timeout=timedelta(hours=2)),
        title=title_input,
    )
    return final_report
```

+++ {"lines_to_next_cell": 0}

## Working with conditionals

The node constructs by themselves are useful, but they become even more
useful when we combine them with other Flyte constructs, like {ref}`conditionals <conditional>`.

To illustrate this, let's extend the report-publishing use case so that we
produce an "invalid report" output in case we don't approve the final report:

```{code-cell}
:lines_to_next_cell: 2

from flytekit import conditional


@task
def invalid_report() -> dict:
    return {"invalid_report": True}


@workflow
def conditional_wf(data: typing.List[float]) -> dict:
    report = create_report(data=data)
    title_input = wait_for_input("title-input", timeout=timedelta(hours=1), expected_type=str)

    # Define a "review-passes" wait_for_input node so that a human can review
    # the report before finalizing it.
    review_passed = wait_for_input("review-passes", timeout=timedelta(hours=2), expected_type=bool)
    report >> review_passed

    # This conditional returns the finalized report if the review passes,
    # otherwise it returns an invalid report output.
    return (
        conditional("final-report-condition")
        .if_(review_passed.is_true())
        .then(finalize_report(report=report, title=title_input))
        .else_()
        .then(invalid_report())
    )
```

On top of the `approved` node, which we use in the `conditional` to
determine which branch to execute, we also define a `disapprove_reason`
gate node, which will be used as an input to the `invalid_report` task.

## Sending inputs to `wait_for_input` and `approve` nodes

Assuming that you've registered the above workflows on a Flyte cluster that's
been started with {ref}`flytectl demo start <getting_started_running_workflow_local_cluster>`,
there are two ways of using `wait_for_input` and `approve` nodes:

### Using the Flyte UI

If you launch the `reporting_wf` workflow on the Flyte UI, you'll see a
**Graph** view of the workflow execution like this:

```{image} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/wait_for_input_graph.png
:alt: reporting workflow wait for input graph
```

Clicking on the {fa}`play-circle,style=far` icon of the `title` task node or the
**Resume** button on the sidebar will create a modal form that you can use to
provide the custom title input.

```{image} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/wait_for_input_form.png
:alt: reporting workflow wait for input form
```

### Using `FlyteRemote`

For many cases it's enough to use Flyte UI to provide inputs/approvals on
gate nodes. However, if you want to pass inputs to `wait_for_input` and
`approve` nodes programmatically, you can use the
{py:meth}`FlyteRemote.set_signal <flytekit.remote.remote.FlyteRemote.set_signal>`
method. Using the `gate_node_with_conditional_wf` workflow, the example
below allows you to set values for `title-input` and `review-passes` nodes.

```python
import typing
from flytekit.remote.remote import FlyteRemote
from flytekit.configuration import Config

remote = FlyteRemote(
    Config.for_sandbox(),
    default_project="flytesnacks",
    default_domain="development",
)

# First kick off the wotrkflow
flyte_workflow = remote.fetch_workflow(
    name="core.control_flow.waiting_for_external_inputs.conditional_wf"
)

# Execute the workflow
execution = remote.execute(flyte_workflow, inputs={"data": [1.0, 2.0, 3.0, 4.0, 5.0]})

# Get a list of signals available for the execution
signals = remote.list_signals(execution.id.name)

# Set a signal value for the "title" node. Make sure that the "title-input"
# node is in the `signals` list above
remote.set_signal("title-input", execution.id.name, "my report")

# Set signal value for the "review-passes" node. Make sure that the "review-passes"
# node is in the `signals` list above
remote.set_signal("review-passes", execution.id.name, True)
```
