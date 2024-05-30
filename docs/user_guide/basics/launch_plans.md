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

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 1
```

We import the workflow from the `workflow.py` file for which we're going to create a launch plan:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 5
```

Create a default launch plan with no inputs during serialization:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 8
```

You can run the launch plan locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 11
```

Create a launch plan and specify the default inputs:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 14-16
```

You can trigger the launch plan locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 19
```

You can override the defaults as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 22
```

It's possible to lock launch plan inputs, preventing them from being overridden during execution:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/launch_plan.py
:caption: basics/launch_plan.py
:lines: 25-27
```

Attempting to modify the inputs will result in an error being raised by Flyte:

:::{note}
You can employ default and fixed inputs in conjunction in a launch plan.
:::

Launch plans can also be used to run workflows on a specific cadence.
For more information, refer to the {ref}`scheduling_launch_plan` documentation.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
