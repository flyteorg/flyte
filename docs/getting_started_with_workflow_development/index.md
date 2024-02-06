(getting_started_workflow_development)=
# Getting started with workflow development

Machine learning engineers, data engineers, and data analysts often represent the processes that consume, transform, and output data with directed acyclic graphs (DAGs). In this section, you will learn how to create a Flyte project to contain the workflow code that implements your DAG, as well as the configuration files needed to package the code to run on a local or remote Flyte cluster.

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`Installing development tools <installing_development_tools>`
  - Install the tools needed to create Flyte projects and run workflows and tasks.
* - {doc}`Creating a Flyte project <../getting_started_with_workflow_development/creating_a_flyte_project>`
  - Create a Flyte project that contains workflow code and essential configuration files.
* - {doc}`Flyte project components <flyte_project_components>`
  - Understand the directory structure, configuration files, and code in a Flyte project.
* - {doc}`Running a workflow locally <running_a_workflow_locally>`
  - Execute a workflow in a local Python environment or in a local Flyte cluster.
```

```{toctree}
:maxdepth: -1
:hidden:

installing_development_tools
creating_a_flyte_project
flyte_project_components
running_a_workflow_locally
```
