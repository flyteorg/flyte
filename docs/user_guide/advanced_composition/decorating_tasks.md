(decorating_tasks)=

# Decorating tasks

```{eval-rst}
.. tags:: Intermediate
```

You can easily change how tasks behave by using decorators to wrap your task functions.

In order to make sure that your decorated function contains all the type annotation and docstring
information that Flyte needs, you will need to use the built-in {py:func}`~functools.wraps` decorator.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import the required dependencies.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:lines: 1-4
```

Create a logger to monitor the execution's progress.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:lines: 7
```

## Using a single decorator

We define a decorator that logs the input and output details for a decorated task.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:pyobject: log_io
```

We create a task named `t1` that is decorated with `log_io`.

:::{note}
The order of invoking the decorators is important. `@task` should always be the outer-most decorator.
:::

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:pyobject: t1
```

(stacking_decorators)=

## Stacking multiple decorators

You can also stack multiple decorators on top of each other as long as `@task` is the outer-most decorator.

We define a decorator that verifies if the output from the decorated function is a positive number before it's returned.
If this assumption is violated, it raises a `ValueError` exception.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:pyobject: validate_output
```

:::{note}
The output of the `validate_output` task uses {py:func}`~functools.partial` to implement parameterized decorators.
:::

We define a function that uses both the logging and validator decorators.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:pyobject: t2
```

Finally, we compose a workflow that calls `t1` and `t2`.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py
:caption: advanced_composition/decorating_tasks.py
:lines: 53-59
```

## Run the example on the Flyte cluster

To run the provided workflow on the Flyte cluster, use the following command:

```
pyflyte run --remote \
  https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/advanced_composition/advanced_composition/decorating_tasks.py \
  decorating_task_wf --x 10
```

In this example, you learned how to modify the behavior of tasks via function decorators using the built-in
{py:func}`~functools.wraps` decorator pattern. To learn more about how to extend Flyte at a deeper level, for
example creating custom types, custom tasks or backend plugins,
see {ref}`Extending Flyte <plugins_extend>`.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/advanced_composition/
