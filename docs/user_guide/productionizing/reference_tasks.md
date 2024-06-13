# Reference tasks

```{eval-rst}
.. tags:: Intermediate
```

A {py:func}`flytekit.reference_task` references the Flyte tasks that have already been defined, serialized, and registered. You can reference tasks from other projects and create workflows that use tasks declared by others. These tasks can be in their own containers, python runtimes, flytekit versions, and even different languages.

The following example illustrates how to use reference tasks.

:::{note}
Reference tasks cannot be run locally. You must mock them out.
:::

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/reference_task.py
:caption: productionizing/reference_task.py
:lines: 1-36
```

:::{note}
The macro `{{ registration.version }}` is populated by `flytectl register` during registration.
Generally, it is unnecessary for reference tasks, as it is preferable to bind to a specific version of the task or launch plan.
However, in this example, we are registering both the task `core.flyte_basics.files.normalize_columns` and the workflow that references it.
Therefore, we need the macro to be updated to the version of a specific Flytesnacks release.
This is why `{{ registration.version }}` is used.

A typical reference task would resemble the following:

```python
@reference_task(
     project="flytesnacks",
     domain="development",
     name="core.flyte_basics.files.normalize_columns",
     version="d06cebcfbeabc02b545eefa13a01c6ca992940c8", # If using GIT for versioning OR 0.16.0, if semver
 )
 def normalize_columns(...):
     ...
```
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
