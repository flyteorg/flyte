=============
Flytekit SDK
=============

.. currentmodule:: flytekit

This package contains all of the most common abstractions you'll need to write Flyte workflows and extend Flytekit.

Basic Authoring
===============

These are the essentials needed to get started writing tasks and workflows.

.. autosummary::
   :nosignatures:
   :toctree: generated/

   task
   workflow
   kwtypes
   current_context
   ExecutionParameters
   FlyteContext
   map_task
   ~core.workflow.ImperativeWorkflow
   ~core.node_creation.create_node
   ~core.promise.NodeOutput
   FlyteContextManager
