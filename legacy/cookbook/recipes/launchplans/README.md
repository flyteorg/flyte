# Creating Launch Plans

[Back to Cookbook Menu](../../)

Launch plans are a simple thin layer on top of workflows that allow users to execute workflows. Please first take a moment to review the [Launch Plans](https://lyft.github.io/flyte/user/concepts/launchplans_schedules.html) documentation.

In order to create a launch plan we first need a workflow. We can use the `ScaleAndRotateWorkflow` created in the workflows chapter.  A default launch plan can be created by just calling the `.create_launch_plan()` function on the Workflow with no arguments.

```python
from workflows import workflows
workflows.ScaleAndRotateWorkflow.create_launch_plan()
```

Users can also specify default input values for some or all of the workflow inputs

```python
scale4x_rotate90degrees_launchplan = workflows.ScaleAndRotateWorkflow.create_launch_plan(
    fixed_inputs={"angle": 90.0, "scale": 4},
)
```

As you use launch plans, you'll find that you want to additional features to control the actual workflow execution. Please see the example in [launchplans.py](launchplans.py) for a preview of notifications, labels and annotations, and refer to [schedules](../multi_schedules/README.md) for information on automatic execution. 
