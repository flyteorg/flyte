# Basics

.. include:: /examples

This section introduces you to the basic building blocks of Flyte
using `flytekit`. `flytekit` is a Python SDK for developing Flyte workflows and
tasks, and can be used generally, whenever stateful computation is desirable.
`flytekit` workflows and tasks are completely runnable locally, unless they need
some advanced backend functionality like starting a distributed Spark cluster.

Here, you will learn how to write Flyte tasks, assemble them into workflows,
run bash scripts, and document workflows.

```{toctree}
:maxdepth: -1
:name: basics_toc
:hidden:

hello_world
tasks
workflows
launch_plans
imperative_workflows
documenting_workflows
/examples/basics/basics/basic_interactive_mode.ipynb
shell_tasks
/examples/basics/basics/basic_interactive_mode
https://raw.githubusercontent.com/Mecoli1219/flytesnacks/85966d139a9a3ccdf3323124563960fdd5f5844a/examples/basics/basics/basic_interactive_mode.ipynb
examples/basics/basics/basic_interactive_mode.ipynb
test
examples/basics/basics/basic_interactive_mode
named_outputs
```
