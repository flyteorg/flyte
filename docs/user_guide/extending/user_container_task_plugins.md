(user_container)=

# User container task plugins

```{eval-rst}
.. tags:: Extensibility, Contribute, Intermediate
```

A user container task plugin runs a user-defined container that has the user code.

This tutorial will walk you through writing your own sensor-style plugin that allows users to wait for a file to land in the object store. Remember that if you follow the flyte/flytekit constructs, you will automatically make your plugin portable across all cloud platforms that Flyte supports.

## Sensor plugin

A sensor plugin waits for some event to happen before marking the task as success. You need not worry about the timeout as that will be handled by the flyte engine itself when running in production.

### Plugin API

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```python
sensor = WaitForObjectStoreFile(metadata=metadata(timeout="1H", retries=10))

@workflow
def wait_and_run(path: str) -> int:
    # To demonstrate how to create outputs, we will also
    # return the output from the sensor. The output will be the
    # same as the path
    path = sensor(path=path)
    return do_next(path=path)
```

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/extending/extending/user_container.py
:caption: extending/user_container.py
:lines: 1-6
```

### Plugin structure

As illustrated above, to achieve this structure we need to create a class named `WaitForObjectStoreFile`, which
derives from {py:class}`flytekit.PythonFunctionTask` as follows.

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/extending/extending/user_container.py
:caption: extending/user_container.py
:pyobject: WaitForObjectStoreFile
```

#### Config objects

Flytekit routes to the right plugin based on the type of `task_config` class if using the `@task` decorator.
Config is very useful for cases when you want to customize the behavior of the plugin or pass the config information
to the backend plugin; however, in this case there's no real configuration. The config object can be any class that your
plugin understands.

:::{note}
Observe that the base class is Generic; it is parameterized with the desired config class.
:::

:::{note}
To create a task decorator-based plugin, `task_config` is required.
In this example, we are creating a named class plugin, and hence, this construct does not need a plugin.
:::

Refer to the [spark plugin](https://github.com/flyteorg/flytekit/tree/master/plugins/flytekit-spark) for an example of a config object.


### Actual usage

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/extending/extending/user_container.py
:caption: extending/user_container.py
:lines: 54-69
```

And of course, you can run the workflow locally using your own new shiny plugin!

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/extending/extending/user_container.py
:caption: extending/user_container.py
:lines: 73-78
```

The key takeaways of a user container task plugin are:

- The task object that gets serialized at compile-time is recreated using the user's code at run time.
- At platform-run-time, the user-decorated function is executed.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/extending/
