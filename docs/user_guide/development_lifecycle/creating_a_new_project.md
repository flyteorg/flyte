---
jupytext:
  cell_metadata_filter: all
  formats: md:myst
  main_language: python
  notebook_metadata_filter: all
  text_representation:
    extension: .md
    format_name: myst
    format_version: 0.13
    jupytext_version: 1.16.1
kernelspec:
  display_name: Python 3
  language: python
  name: python3
---

# Creating a new project

Creates project to be used as a home for the flyte resources of tasks and workflows.
Refer to the [flytectl API reference](https://docs.flyte.org/projects/flytectl/en/stable/gen/flytectl_create_project.html)
for more details.

```{eval-rst}
.. prompt:: bash

  flytectl create project --id "my-flyte-project-name" --labels "my-label=my-project-label"  --description "my-flyte-project-name" --name "my-flyte-project-name"
```
