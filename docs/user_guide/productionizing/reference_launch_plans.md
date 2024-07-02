# Reference launch plans

```{eval-rst}
.. tags:: Intermediate
```

A {py:func}`flytekit.reference_launch_plan` references previously defined, serialized, and registered Flyte launch plans. You can reference launch plans from other projects and create workflows that use launch plans declared by others.

The following example illustrates how to use reference launch plans.

:::{note}
Reference launch plans cannot be run locally. You must mock them out.
:::

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/productionizing/productionizing/reference_launch_plan.py
:caption: productionizing/reference_launch_plan.py
:lines: 1-36
```

It's important to verify that the workflow interface corresponds to that of the referenced workflow.

:::{note}
The macro `{{ registration.version }}` is populated by `flytectl register` during registration.
Generally, it is unnecessary for reference launch plans, as it is preferable to bind to a specific version of the task or launch plan.
However, in this example, we are registering both the launch plan `core.flyte_basics.files.normalize_csv_file` and the workflow that references it.
Therefore, we need the macro to be updated to the version of a specific Flytesnacks release.
This is why `{{ registration.version }}` is used.

A typical reference launch plan would resemble the following:

```python
@reference_launch_plan(
    project="flytesnacks",
    domain="development",
    name="core.flyte_basics.files.normalize_csv_file",
    version="d06cebcfbeabc02b545eefa13a01c6ca992940c8", # If using GIT for versioning OR 0.16.0, if semver
)
def normalize_csv_file(...):
    ...
```
:::

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/productionizing/
