(getting_started_fundamentals)=

# Flyte Fundamentals

This section of the documentation will take you through the
fundamental concepts of Flyte: tasks, workflows, and launch plans.

You'll learn about the full development lifecycle of creating a project,
registering workflows, and running them on a demo Flyte cluster. These
guides will also walk you through how to visualize artifacts associated with
tasks, optimize them for scale and performance, and extend Flyte for your own
use cases.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`🔀 Tasks, Workflows and LaunchPlans <tasks_and_workflows>`
  - Create tasks as building blocks, compose them into workflows, and schedule
    them with launchplans.
* - {doc}`🗄 Registering Workflows <package_register>`
  - Develop and deploy workflows to a local Flyte demo cluster.
* - {doc}`⏱ Running and Scheduling Workflows <run_schedule>`
  - Execute workflows programmatically and schedule them as cron jobs.
* - {doc}`📊 Visualizing Task Input and Output <visualizing_task_input_and_output>`
  - Create rich, customizable static reports for increased visibility into tasks.
* - {doc}`🏎 Optimizing Tasks <optimizing_tasks>`
  - Make tasks scalable, performant, and robust to unexpected failures.
* - {doc}`🔌 Extending Flyte <extending_flyte>`
  - Customize Flyte types and tasks to fit your needs.
```

```{admonition} Learn more
:class: important

For a comprehensive view of all of Flyte's functionality, see the
{ref}`User Guide <userguide>`, and to learn how to deploy a production Flyte
cluster, see the {ref}`Deployment Guide <deployment>`.
```

```{toctree}
:maxdepth: -1
:hidden:

tasks_and_workflows
package_register
run_schedule
visualizing_task_input_and_output
optimizing_tasks
extending_flyte
```
