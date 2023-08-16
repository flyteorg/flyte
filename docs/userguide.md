---
:next-page: environment_setup
:next-page-title: Environment Setup
:prev-page: getting_started/analytics
:prev-page-title: Analytics
---

(userguide)=

# User Guide

If this is your first time using Flyte, check out the {doc}`Getting Started <index>` guide.

This *User Guide*, the {doc}`Tutorials <tutorials>`, and the {doc}`Integrations <integrations>` examples cover all of
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
Want to contribute an example? Check out the {doc}`Example Contribution Guide <contribute>`.
```

## Table of Contents

```{list-table}
:header-rows: 0
:widths: 20 30

* - {doc}`🌳 Environment Setup <_repos/flytesnacks/docs/environment_setup>`
  - Setup your development environment to run the User Guide examples.
* - {doc}`🔤 Flyte Basics <auto_examples/basics/index>`
  - Learn about tasks, workflows, launch plans, caching, and working with files and directories.
* - {doc}`🚰 Control Flow <auto_examples/control_flow/index>`
  - Implement conditionals, nested and dynamic workflows, map tasks, and even recursion!
* - {doc}`⌨️ Type System <auto_examples/type_system/index>`
  - Improve pipeline robustness with Flyte's portable and extensible type system.
* - {doc}`⚗️ Testing <auto_examples/testing/index>`
  - Test tasks and workflows with Flyte's testing utilities.
* - {doc}`📦 Containerization <auto_examples/containerization/index>`
  - Easily manage the complexity of configuring the containers that run Flyte tasks.
* - {doc}`🐳 Image Spec <auto_examples/image_spec/index>`
  - Build a container image without a Dockerfile.
* - {doc}`🎮 Remote Access <auto_examples/remote_access/index>`
  - Register, inspect, and monitor tasks and workflows on a Flyte backend.
* - {doc}`🚢  Production Config <auto_examples/deployment/index>`
  - Ship and configure your machine learning pipelines on a production Flyte installation.
* - {doc}`🏗 Extending Flyte <auto_examples/extend_flyte/index>`
  - Define custom plugins that aren't currently supported in the Flyte ecosystem.
```

```{toctree}
:maxdepth: -1
:caption: User Guide
:hidden:

Environment Setup <_repos/flytesnacks/docs/environment_setup>
Basics <auto_examples/basics/index>
Control Flow <auto_examples/control_flow/index>
Type System <auto_examples/type_system/index>
Testing <auto_examples/testing/index>
Containerization <auto_examples/containerization/index>
Image Spec <auto_examples/image_spec/index>
Remote Access <auto_examples/remote_access/index>
Production Config <auto_examples/deployment/index>
Extending Flyte <auto_examples/extend_flyte/index>
Example Contribution Guide <contribute>
```
