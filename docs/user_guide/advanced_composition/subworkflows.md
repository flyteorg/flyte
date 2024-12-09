(subworkflow)=

# Subworkflows

```{eval-rst}
.. tags:: Intermediate
```

Subworkflows share similarities with {ref}`launch plans <Launch plans>`, as both enable users to initiate one workflow from within another.
The distinction lies in the analogy: think of launch plans as "pass by pointer" and subworkflows as "pass by value."

## When to use subworkflows?

Subworkflows offer an elegant solution for managing parallelism between a workflow and its launched sub-flows,
as they execute within the same context as the parent workflow.
Consequently, all nodes of a subworkflow adhere to the overall constraints imposed by the parent workflow.

Consider this scenario: when workflow `A` is integrated as a subworkflow of workflow `B`,
running workflow `B` results in the entire graph of workflow `A` being duplicated into workflow `B` at the point of invocation.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

Here's an example illustrating the calculation of slope, intercept and the corresponding y-value:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py
:caption: advanced_composition/subworkflow.py
:lines: 1-35
```

The `slope_intercept_wf` computes the slope and intercept of the regression line.
Subsequently, the `regression_line_wf` triggers `slope_intercept_wf` and then computes the y-value.

To execute the workflow locally, use the following:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py
:caption: advanced_composition/subworkflow.py
:lines: 39-40
```

It's possible to nest a workflow that contains a subworkflow within another workflow.
Workflows can be easily constructed from other workflows, even if they function as standalone entities.
Each workflow in this module has the capability to exist and run independently:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py
:caption: advanced_composition/subworkflow.py
:pyobject: nested_regression_line_wf
```

You can run the nested workflow locally as well:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py
:caption: advanced_composition/subworkflow.py
:lines: 52-53
```

## External workflow

When launch plans are employed within a workflow to initiate the execution of a pre-defined workflow,
a new external execution is triggered. This results in a distinct execution ID and can be identified
as a separate entity.

These external invocations of a workflow, initiated using launch plans from a parent workflow,
are termed as external workflows. They may have separate parallelism constraints since the context is not shared.

:::{tip}
If your deployment uses {ref}`multiple Kubernetes clusters <flyte:deployment-deployment-multicluster>`,
external workflows may offer a way to distribute the workload of a workflow across multiple clusters.
:::

Here's an example that illustrates the concept of external workflows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py
:caption: advanced_composition/subworkflow.py
:lines: 61-71
```

:::{figure} https://raw.githubusercontent.com/flyteorg/static-resources/main/flytesnacks/user_guide/flyte_external_workflow_execution.png
:alt: External workflow execution
:class: with-shadow
:::

In the console screenshot above, note that the launch plan execution ID differs from that of the workflow.

You can run a workflow containing an external workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py
:caption: advanced_composition/subworkflow.py
:lines: 75-76
```

## Run the example on a Flyte cluster

To run the provided workflows on a Flyte cluster, use the following commands:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py \
  regression_line_wf
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py \
  nested_regression_line_wf
```

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/subworkflow.py \
  nested_regression_line_lp
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
