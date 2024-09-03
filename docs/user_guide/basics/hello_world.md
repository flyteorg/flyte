# Hello, World!

```{eval-rst}
.. tags:: Basic
```

Let's write a Flyte {py:func}`~flytekit.workflow` that invokes a
{py:func}`~flytekit.task` to generate the output "Hello, World!".

Flyte tasks are the core building blocks of larger, more complex workflows.
Workflows compose multiple tasks – or other workflows –
into meaningful steps of computation to produce some useful set of outputs or outcomes.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

To begin, import `task` and `workflow` from the `flytekit` library:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/hello_world.py
:caption: basics/hello_world.py
:lines: 1
```

Define a task that produces the string "Hello, World!".
Simply using the `@task` decorator to annotate the Python function:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/hello_world.py
:caption: basics/hello_world.py
:pyobject: say_hello
```

You can handle the output of a task in the same way you would with a regular Python function.
Store the output in a variable and use it as a return value for a Flyte workflow:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/hello_world.py
:caption: basics/hello_world.py
:pyobject: hello_world_wf
```

Run the workflow by simply calling it like a Python function:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/hello_world.py
:caption: basics/hello_world.py
:lines: 19-20
```

Next, let's delve into the specifics of {ref}`tasks <task>`,
{ref}`workflows <workflow>` and {ref}`launch plans <launch_plan>`.

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
