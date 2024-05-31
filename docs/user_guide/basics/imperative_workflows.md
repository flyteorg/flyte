(imperative_workflow)=

# Imperative workflows

```{eval-rst}
.. tags:: Basic
```

Workflows are commonly created by applying the `@workflow` decorator to Python functions.
During compilation, this involves processing the function's body and utilizing subsequent calls to
underlying tasks to establish and record the workflow structure. This approach is known as declarative
and is suitable when manually drafting the workflow.

However, in cases where workflows are constructed programmatically, an imperative style is more appropriate.
For instance, if tasks have been defined already, their sequence and dependencies might have been specified
in textual form (perhaps during a transition from a legacy system).
In such scenarios, you want to orchestrate these tasks.
This is where Flyte's imperative workflows come into play, allowing you to programmatically construct workflows.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the necessary dependencies:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 1
```

We import the `slope` and `intercept` tasks from the `workflow.py` file:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 4
```

Create an imperative workflow:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 7
```

Add the workflow inputs to the imperative workflow:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 11-12
```

::: {note}
If you want to assign default values to the workflow inputs,
you can create a {ref}`launch plan <launch_plan>`.
:::

Add the tasks that need to be triggered from within the workflow:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 16-19
```

Lastly, add the workflow output:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 23
```

You can execute the workflow locally as follows:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/imperative_workflow.py
:caption: basics/imperative_workflow.py
:lines: 27-28
```

:::{note}
You also have the option to provide a list of inputs and
retrieve a list of outputs from the workflow:

```python
wf_input_y = imperative_wf.add_workflow_input("y", list[str])
node_t3 = wf.add_entity(some_task, a=[wf.inputs["x"], wf_input_y])
```

```python
wf.add_workflow_output(
    "list_of_outputs",
    [node_t1.outputs["o0"], node_t2.outputs["o0"]],
    python_type=list[str],
)
```
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
