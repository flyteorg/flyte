.. _basic:

Basics
------
In this section, you'll learn how to use the basic building blocks of Flyte using ``flytekit``.
``flytekit`` is a python SDK for developing flyte workflows and task and can be used generally, whenever stateful
computation is desirable. ``flytekit``-developed workflows and tasks are completely runnable locally, unless they need
some advanced backend functionality like starting a distributed spark cluster.

In this section we'll take a look at how to write flyte tasks, compose them together to form a workflow,
and then read, manipulate and cache data.
