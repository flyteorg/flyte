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
    jupytext_version: 1.14.7
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

+++ {"lines_to_next_cell": 0}

(imperative_wf_style)=

# Imperative Workflows

```{eval-rst}
.. tags:: Basic
```

Workflows are typically created and specified by decorating a function with the `@workflow` decorator. This will
run through the body of the function at compile time, using the subsequent calls of the underlying tasks to determine
and record the workflow structure. This is the declarative style and makes sense when a human is writing it up by hand.

For cases where workflows are constructed programmatically, an imperative style makes more sense. For example, when tasks have already been defined, their order and dependencies have
been specified in text format of some kind (perhaps you're converting from a legacy system), and your goal is to orchestrate those tasks.

```{code-cell}
import typing

from flytekit import Workflow, task
```

+++ {"lines_to_next_cell": 0}

Assume we have the following tasks, and they are meant to represent more complicated tasks. `t1` has simple scalar I/O, `t2` is
a pure side effect task (though we typically don't recommend these, they are inevitable), and `t3` takes in a list
as an input.

```{code-cell}
@task
def t1(a: str) -> str:
    return a + " world"


@task
def t2():
    print("side effect")


@task
def t3(a: typing.List[str]) -> str:
    """
    This is a pedagogical demo that happens to do a reduction step. Flyte is higher-order orchestration
    platform, not a map-reduce framework and is not meant to supplant Spark et. al.
    """
    return ",".join(a)
```

+++ {"lines_to_next_cell": 0}

Start by creating an imperative style workflow, which is aliased to just `Workflow` from `flytekit`

```{code-cell}
:lines_to_next_cell: 2

wf = Workflow(name="my.imperative.workflow.example")
```

+++ {"lines_to_next_cell": 0}

Inputs have to be added to the workflow before they can be used. Add them by specifying the name and the type.

```{code-cell}
wf.add_workflow_input("in1", str)
```

+++ {"lines_to_next_cell": 0}

Next associate a task, and pass in the workflow level input.

```{code-cell}
node_t1 = wf.add_entity(t1, a=wf.inputs["in1"])
```

+++ {"lines_to_next_cell": 0}

Create a workflow output linked to the output of that task.

```{code-cell}
wf.add_workflow_output("output_from_t1", node_t1.outputs["o0"])
```

+++ {"lines_to_next_cell": 0}

To add a task that has no inputs or outputs, just add the entity. We don't need to capture the resulting node
because we have no use for it.

```{code-cell}
wf.add_entity(t2)
```

+++ {"lines_to_next_cell": 0}

We can also pass in a list to a task. Also creating a workflow input returns an object that can be used as an
alternate way of linking workflow inputs. Here, `t3` uses both workflow inputs.

```{code-cell}
wf_in2 = wf.add_workflow_input("in2", str)
node_t3 = wf.add_entity(t3, a=[wf.inputs["in1"], wf_in2])
```

+++ {"lines_to_next_cell": 0}

You can also create a workflow output as a list from multiple task outputs

```{code-cell}
wf.add_workflow_output(
    "output_list",
    [node_t1.outputs["o0"], node_t3.outputs["o0"]],
    python_type=typing.List[str],
)


if __name__ == "__main__":
    print(wf)
    print(wf(in1="hello", in2="foo"))
```
