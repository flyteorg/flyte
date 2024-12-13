# Mocking tasks

A lot of the tasks that you write you can run locally, but some of them you will not be able to, usually because they are tasks that depend on a third-party only available on the backend. Hive tasks are a common example, as most users will not have access to the service that executes Hive queries from their development environment. However, it's still useful to be able to locally run a workflow that calls such a task. In these instances, flytekit provides a couple of utilities to help navigate this.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

```{literalinclude} /examples/testing/testing/mocking.py
:caption: testing/mocking.py
:lines: 1-6
```

This is a generic SQL task (and is by default not hooked up to any datastore nor handled by any plugin), and must be mocked:

```{literalinclude} /examples/testing/testing/mocking.py
:caption: testing/mocking.py
:lines: 10-16
```

This is a task that can run locally:

```{literalinclude} /examples/testing/testing/mocking.py
:caption: testing/mocking.py
:pyobject: t1
```

Declare a workflow that chains these two tasks together.

```{literalinclude} /examples/testing/testing/mocking.py
:caption: testing/mocking.py
:pyobject: my_wf
```

Without a mock, calling the workflow would typically raise an exception, but with the `task_mock` construct, which returns a `MagicMock` object, we can override the return value.

```{literalinclude} /examples/testing/testing/mocking.py
:caption: testing/mocking.py
:pyobject: main_1
```

There is another utility as well called `patch` which offers the same functionality, but in the traditional Python patching style, where the first argument is the `MagicMock` object.

```{literalinclude} /examples/testing/testing/mocking.py
:caption: testing/mocking.py
:lines: 45-56
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/testing/
