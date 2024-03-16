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

(launch_plan)=

# Launch plans

```{eval-rst}
.. tags:: Basic
```

Launch plans link a partial or complete list of inputs required to initiate a workflow,
accompanied by optional run-time overrides like notifications, schedules and more.
They serve various purposes:

- Schedule the same workflow multiple times, with optional predefined inputs.
- Run a specific workflow but with altered notifications.
- Share a workflow with predefined inputs, allowing another user to initiate an execution.
- Share a workflow with the option for the other user to override certain inputs.
- Share a workflow, ensuring specific inputs remain unchanged.

Launch plans are the only means for invoking workflow executions.
When a workflow is serialized and registered, a _default launch plan_ is generated.
This default launch plan can bind default workflow inputs and runtime options defined
in the project's flytekit configuration (such as user role).

To begin, import the necessary libraries.

```{code-cell}
from flytekit import LaunchPlan, current_context
```

+++ {"lines_to_next_cell": 0}

We import the workflow from the `workflow.py` file for which we're going to create a launch plan.

```{code-cell}
from .workflow import simple_wf
```

+++ {"lines_to_next_cell": 0}

Create a default launch plan with no inputs during serialization.

```{code-cell}
default_lp = LaunchPlan.get_default_launch_plan(current_context(), simple_wf)
```

+++ {"lines_to_next_cell": 0}

You can run the launch plan locally as follows:

```{code-cell}
default_lp(x=[-3, 0, 3], y=[7, 4, -2])
```

+++ {"lines_to_next_cell": 0}

Create a launch plan and specify the default inputs.

```{code-cell}
simple_wf_lp = LaunchPlan.create(
    name="simple_wf_lp", workflow=simple_wf, default_inputs={"x": [-3, 0, 3], "y": [7, 4, -2]}
)
```

+++ {"lines_to_next_cell": 0}

You can trigger the launch plan locally as follows:

```{code-cell}
simple_wf_lp()
```

+++ {"lines_to_next_cell": 0}

You can override the defaults as follows:

```{code-cell}
simple_wf_lp(x=[3, 5, 3], y=[-3, 2, -2])
```

+++ {"lines_to_next_cell": 0}

It's possible to lock launch plan inputs, preventing them from being overridden during execution.

```{code-cell}
simple_wf_lp_fixed_inputs = LaunchPlan.get_or_create(
    name="fixed_inputs", workflow=simple_wf, fixed_inputs={"x": [-3, 0, 3]}
)
```

Attempting to modify the inputs will result in an error being raised by Flyte.

:::{note}
You can employ default and fixed inputs in conjunction in a launch plan.
:::

Launch plans can also be used to run workflows on a specific cadence.
For more information, refer to the {ref}`scheduling_launch_plan` documentation.
