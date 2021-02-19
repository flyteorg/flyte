.. _basic:

Basic Examples: Tasks, Workflows, Files etc
--------------------------------------------
This section provides insight into basic building blocks of Flyte, especially flytekit.
Flytekit is a python SDK for developing flyte workflows and task and can be used generally, whenever stateful computation is
desirable. Flytekit developed workflows and tasks are completely runnable locally, unless they need some advanced backend
functionality like, starting a distributed spark cluster.

In this first getting started section we'll take a look at how to write flyte tasks, string them together to form a workflow,
and then read, manipulate and cache data.