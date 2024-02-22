---
:next-page: environment_setup
:next-page-title: Environment Setup
:prev-page: getting_started/analytics
:prev-page-title: Analytics
---

(userguide)=

# User guide

If this is your first time using Flyte, check out the {doc}`Getting Started <index>` guide.

This _User guide_, the {ref}`Tutorials <tutorials>` and the {ref}`Integrations <integrations>` examples cover all of
the key features of Flyte for data analytics, data science and machine learning practitioners, organized by topic. Each
section below introduces a core feature of Flyte and how you can use it to address specific use cases. Code for all
of the examples can be found in the [flytesnacks repo](https://github.com/flyteorg/flytesnacks).

It comes with a specific environment to make running, documenting
and contributing samples easy. If this is your first time running these examples, follow the
{doc}`environment setup guide <environment_setup>` to get started.

```{tip}
To learn about how to spin up and manage a Flyte cluster in the cloud, see the
{doc}`Deployment Guides <flyte:deployment/index>`.
```

```{note}
Want to contribute or update an example? Check out the {doc}`Contribution Guide <../flytesnacks/contribute>`.
```

## Table of contents

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`🌳 Environment Setup <environment_setup>`
  - Set up a development environment to run the examples in the user guide.
* - {doc}`🔤 Basics <basics/index>`
  - Learn about tasks, workflows, launch plans, caching and managing files and directories.
* - {doc}`⌨️ Data Types and IO <data_types_and_io/index>`
  - Improve pipeline robustness with Flyte's portable and extensible type system.
* - {doc}`🔮 Advanced Composition <advanced_composition/index>`
  - Implement conditionals, nested and dynamic workflows, map tasks and even recursion!
* - {doc}`🧩 Customizing Dependencies <customizing_dependencies/index>`
  - Provide custom dependencies to run your Flyte entities.
* - {doc}`🏡 Development Lifecycle <development_lifecycle/index>`
  - Develop and test locally on the demo cluster.
* - {doc}`⚗️ Testing <testing/index>`
  - Test tasks and workflows with Flyte's testing utilities.
* - {doc}`🚢 Productionizing <productionizing/index>`
  - Ship and configure your machine learning pipelines on a production Flyte installation.
* - {doc}`🏗 Extending <extending/index>`
  - Define custom plugins that aren't currently supported in the Flyte ecosystem.
```

```{toctree}
:maxdepth: -1
:name: user_guide_toc
:hidden:

environment_setup
basics/index
data_types_and_io/index
advanced_composition/index
customizing_dependencies/index
development_lifecycle/index
testing/index
productionizing/index
extending/index
```