(shell_task)=

# Shell tasks

```{eval-rst}
.. tags:: Basic
```

To execute bash scripts within Flyte, you can utilize the {py:class}`~flytekit.extras.tasks.shell.ShellTask` class.
This example includes three shell tasks to execute bash commands.

```{note}
To clone and run the example code on this page, see the [Flytesnacks repo][flytesnacks].
```

First, import the necessary libraries:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/shell_task.py
:caption: basics/shell_task.py
:lines: 1-8
```

With the required imports in place, you can proceed to define a shell task.
To create a shell task, provide a name for it, specify the bash script to be executed,
and define inputs and outputs if needed:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/shell_task.py
:caption: basics/shell_task.py
:lines: 13-55
```

Here's a breakdown of the parameters of the `ShellTask`:

- The `inputs` parameter allows you to specify the types of inputs that the task will accept
- The `output_locs` parameter is used to define the output locations, which can be `FlyteFile` or `FlyteDirectory`
- The `script` parameter contains the actual bash script that will be executed
  (`{inputs.x}`, `{outputs.j}`, etc. will be replaced with the actual input and output values).
- The `debug` parameter is helpful for debugging purposes

We define a task to instantiate `FlyteFile` and `FlyteDirectory`.
A `.gitkeep` file is created in the FlyteDirectory as a placeholder to ensure the directory exists:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/shell_task.py
:caption: basics/shell_task.py
:pyobject: create_entities
```

We create a workflow to define the dependencies between the tasks:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/shell_task.py
:caption: basics/shell_task.py
:pyobject: shell_task_wf
```

You can run the workflow locally:

```{rli} https://raw.githubusercontent.com/flyteorg/flytesnacks/69dbe4840031a85d79d9ded25f80397c6834752d/examples/basics/basics/shell_task.py
:caption: basics/shell_task.py
:lines: 85-86
```

[flytesnacks]: https://github.com/flyteorg/flytesnacks/tree/master/examples/basics/
