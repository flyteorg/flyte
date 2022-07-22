# Flyte

[Flyte](https://github.com/lyft/flyte) is a container-native, type-safe workflow and pipelines platform optimized for large scale processing and machine learning written in Golang. Workflows can be written in any language, with out of the box support for Python.

# Introduction
Flyte is a fabric that connects disparate computation backends using a type safe data dependency graph. It records all changes to a pipeline, making it possible to rewind time. It also stores
a history of all executions and provides an intuitive UI, CLI and REST/gRPC API to interact with the computation.

Flyte is more than a workflow engine, it provides workflows as a core concepts, but it also provides a single unit of execution - tasks, as a top level concept. Multiple tasks arranged in a data
producer-consumer order creates a workflow. Flyte workflows are pure specification and can be created using any language. Every task can also by any language. We do provide first class support for
python, making it perfect for modern Machine Learning and Data processing pipelines.

The following scenario demonstrates how to install sandbox setup for Flyte, allowing you to create workflow and execute them in UI

You will learn how to:

1. Install sandbox setup for flyte.
2. Register First Workflow
3. How to use Flyteconsole to execute your first workflow

Once the Flyte is in place, it can be extended to support ..... These are discussed in more advanced scenarios.
