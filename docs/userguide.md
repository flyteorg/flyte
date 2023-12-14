(userguide)=

# User Guide

If this is your first time using Flyte, check out the {doc}`Getting Started <index>` guide.

This *User Guide*, the {doc}`Tutorials <tutorials>`, and the {doc}`Integrations <integrations>` examples cover all of
the key features of Flyte for data analytics, data science and machine learning practitioners, organized by topic. Each
section below introduces a core feature of Flyte and how you can use it to address specific use cases. Code for all
of the examples can be found in the [flytesnacks repo](https://github.com/flyteorg/flytesnacks).

It comes with a specific environment to make running, documenting
and contributing samples easy. If this is your first time running these examples, follow the
{doc}`environment setup guide <flytesnacks/environment_setup>` to get started.

```{tip}
To learn about how to spin up and manage a Flyte cluster in the cloud, see the
{doc}`Deployment Guides <deployment/index>`.
```

```{note}
Want to contribute an example? Check out the {doc}`Example Contribution Guide <flytesnacks/contribute>`.
```

## Table of Contents

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`🌳 Environment Setup <flytesnacks/environment_setup>`
  - Set up a development environment to run the examples in the user guide.
* - {doc}`🔤 Basics <flytesnacks/examples/basics/index>`
  - Learn about tasks, workflows, launch plans, caching and managing files and directories.
* - {doc}`⌨️ Data Types and IO <flytesnacks/examples/data_types_and_io/index>`
  - Improve pipeline robustness with Flyte's portable and extensible type system.
* - {doc}`🔮 Advanced Composition <flytesnacks/examples/advanced_composition/index>`
  - Implement conditionals, nested and dynamic workflows, map tasks and even recursion!
* - {doc}`🧩 Customizing Dependencies <flytesnacks/examples/customizing_dependencies/index>`
  - Provide custom dependencies to run your Flyte entities.
* - {doc}`🏡 Development Lifecycle <flytesnacks/examples/development_lifecycle/index>`
  - Develop and test locally on the demo cluster.
* - {doc}`⚗️ Testing <flytesnacks/examples/testing/index>`
  - Test tasks and workflows with Flyte's testing utilities.
* - {doc}`🚢 Productionizing <flytesnacks/examples/productionizing/index>`
  - Ship and configure your machine learning pipelines on a production Flyte installation.
* - {doc}`🏗 Extending <flytesnacks/examples/extending/index>`
  - Define custom plugins that aren't currently supported in the Flyte ecosystem.
```

```{toctree}
:maxdepth: -1
:caption: User Guide
:hidden:

Environment Setup <flytesnacks/environment_setup>
Basics <flytesnacks/examples/basics/index>
Data Types and IO <flytesnacks/examples/data_types_and_io/index>
Advanced Composition <flytesnacks/examples/advanced_composition/index>
Customizing Dependencies <flytesnacks/examples/customizing_dependencies/index>
Development Lifecycle <flytesnacks/examples/development_lifecycle/index>
Testing <flytesnacks/examples/testing/index>
Productionizing <flytesnacks/examples/productionizing/index>
Extending <flytesnacks/examples/extending/index>
```
