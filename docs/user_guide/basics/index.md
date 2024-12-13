# Basics

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
shell_tasks
named_outputs
```
