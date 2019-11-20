.. _introduction:

Introduction
============
Flyte is a structured programming and distributed processing platform created at Lyft that enables highly concurrent, scalable and maintainable workflows for machine learning and data processing. Flyte consists of tasks as fundamental building blocks (like functions in functional programming) that can be linked together in a directed acyclic graph to form workflows (control flow specification). Flyte uses protocol buffers as the specification language to specify these workflows and tasks. The actual implementation of the task can be in any language as the tasks themselves are executed as containers. Flyte comes with Flytekit, a python SDK to develop applications on Flyte (authoring workflows, tasks etc). Want to learn more, Read ON!



.. toctree::
   :maxdepth: 2
   :caption: Flyte Introduction
   :name: introductiontoc

   whatis
   docs_overview
