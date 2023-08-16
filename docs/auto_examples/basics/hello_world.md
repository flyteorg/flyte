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


# Hello World


```{tags} Basic
```

This simple workflow calls a task that returns "Hello World" and then just sets that as the final output of the workflow.

```{code-cell}
:lines_to_next_cell: 1

from flytekit import task, workflow
```

You can change the signature of the workflow to take in an argument like this:

```{code-cell}
@task
def say_hello() -> str:
    return "hello world"
```

You can treat the outputs of a task as you normally would a Python function. Assign the output to two variables
and use them in subsequent tasks as normal. See {py:func}`flytekit.workflow`
You can change the signature of the workflow to take in an argument like this:

```{code-cell}
@workflow
def my_wf() -> str:
    res = say_hello()
    return res
```

Execute the Workflow, simply by invoking it like a function and passing in
the necessary parameters

```{note}

One thing to remember, currently we only support `Keyword arguments`. So
every argument should be passed in the form `arg=value`. Failure to do so
will result in an error
````

```{code-cell}
:lines_to_next_cell: 2

if __name__ == "__main__":
    print(f"Running my_wf() {my_wf()}")
```

In the next few examples you'll learn more about the core ideas of Flyte, which are tasks, workflows, and launch
plans.
