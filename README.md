![Flyte Logo](images/flyte_lockup_gradient_on_light.png "Flyte Logo")

[![Current Release](https://img.shields.io/github/release/lyft/flyte.svg)](https://github.com/lyft/flyte/releases/latest)
[![Build Status](https://travis-ci.org/lyft/flyte.svg?branch=master)](https://travis-ci.org/lyft/flyte)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
![Commit activity](https://img.shields.io/github/commit-activity/w/lyft/flyte.svg?style=plastic)
![Commit since last release](https://img.shields.io/github/commits-since/lyft/flyte/latest.svg?style=plastic)
![GitHub milestones Completed](https://img.shields.io/github/milestones/closed/lyft/flyte?style=plastic)
![GitHub next milestone percentage](https://img.shields.io/github/milestones/progress-percent/lyft/flyte/2?style=plastic)
![Twitter Follow](https://img.shields.io/twitter/follow/flyteorg?label=Follow&style=social)
[![Slack Status](https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social)](https://docs.google.com/forms/d/e/1FAIpQLSf8bNuyhy7rkm77cOXPHIzCm3ApfL7Tdo7NUs6Ej2NOGQ1PYw/viewform?pli=1)

Flyte is a container-native, type-safe workflow and pipelines platform optimized for large scale processing and machine learning written in Golang. Workflows can be written in any language, with out of the box support for Python. 

# Homepage
https://flyte.org
Docs: https://lyft.github.io/flyte

# Introduction
Flyte is a fabric that connects disparate computation backends using a type safe data dependency graph. It records all changes to a pipeline, making it possible to rewind time. It also stores
a history of all executions and provides an intuitive UI, CLI and REST/gRPC API to interact with the computation.

Flyte is more than a workflow engine, it provides workflows as a core concepts, but it also provides a single unit of execution - tasks, as a top level concept. Multiple tasks arranged in a data
producer-consumer order creates a workflow. Flyte workflows are pure specification and can be created using any language. Every task can also by any language. We do provide first class support for
python, making it perfect for modern Machine Learning and Data processing pipelines.

# Features
 - Used at Scale in production by 500+ users at Lyft with more than 400k workflows a month and more than 20+ million container executions per month
 - Centralized Inventory of Tasks, Workflows and Executions
 - gRPC / REST interface to define and executes tasks and workflows
 - Type safe construction of pipelines, each task has an interface which is characterized by its input and outputs. Thus illegal construction of pipelines fails during declaration rather than at
   runtime
 - Types that help in creating machine learning and data processing pipelines like - Blobs (images, arbitrary files), Directories, Schema (columnar structured data), collections, maps etc
 - Memoization and Lineage tracking
 - Workflows features
  * Multiple Schedules for every workflow
  * Parallel step execution
  * Extensible Backend to add customized plugin experiences
  * Arbitrary container execution
  * Branching
  * Inline Subworkflows (a workflow can be embeded within one node of the top level workflow)
  * Distributed Remote Child workflows (a remote workflow can be triggered and statically verified at compile time)
  * Array Tasks (map some function over a large dataset, controlled execution of 1000's of containers)
  * Dynamic Workflow creation and execution - with runtime type safety
  * Container side plugins with first class support in python
 - Maintain an inventory of tasks and workflows
 - Record history of all executions and executions (as long as they follow convention) are completely repeatable
 - Multi Cloud support (AWS, GCP and others)
 - Extensible core
 - Modularized
 - Automated notifications to Slack, Email, Pagerduty
 - Deep observability
 - Multi K8s cluster support
 - Comes with many system supported out of the box on K8s like Spark etc.
 - Snappy Console
 - Python CLI
 - Written in Golang and optimized for performance

## Coming Soon
 - Single Task Execution support
 - Reactive pipelines
 - More integrations

# Current Usage 
- Lyft Rideshare
- Lyft L5 autonomous
- Juno

# Component Repos 
Repo | Language | Purpose
--- | --- | ---
[flyte](https://github.com/lyft/flyte) | Kustomize,RST | deployment, documentation, issues
[flyteidl](https://github.com/lyft/flyteidl) | Protobuf | interface definitions
[flytepropeller](https://github.com/lyft/flytepropeller) | Go | execution engine
[flyteadmin](https://github.com/lyft/flyteadmin) | Go | control plane
[flytekit](https://github.com/lyft/flytekit) | Python | python SDK and tools
[flyteconsole](https://github.com/lyft/flyteconsole) | Typescript | admin console
[datacatalog](https://github.com/lyft/datacatalog) | Go  | manage input & output artifacts
[flyteplugins](https://github.com/lyft/flyteplugins) | Go  | flyte plugins
[flytestdlib](https://github.com/lyft/flytestdlib) |  Go | standard library
[flytesnacks](https://github.com/lyft/flytesnacks) | Python | examples, tips, and tricks

# Production K8s Operators

Repo | Language | Purpose
--- | --- | ---
[Spark](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator) | Go | Apache Spark batch
[Flink](https://github.com/lyft/flinkk8soperator) | Go | Apache Flink streaming
