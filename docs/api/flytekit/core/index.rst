=============
Core Flytekit
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

.. important::

   Tasks and Workflows can both be locally run, assuming the relevant tasks are capable of local execution.
   This is useful for unit testing.


Branching and Conditionals
==========================

Branches and conditionals can be expressed explicitly in Flyte. These conditions are evaluated
in the flyte engine and hence should be used for control flow. ``dynamic workflows`` can be used to perform custom conditional logic not supported by flytekit

.. autosummary::
   :nosignatures:
   :toctree: generated/

   conditional


Customizing Tasks & Workflows
==============================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   TaskMetadata - Wrapper object that allows users to specify Task
   Resources - Things like CPUs/Memory, etc.
   WorkflowFailurePolicy - Customizes what happens when a workflow fails.
   PodTemplate - Custom PodTemplate for a task.

Dynamic and Nested Workflows
==============================
See the :py:mod:`Dynamic <flytekit.core.dynamic_workflow_task>` module for more information.

.. autosummary::
   :nosignatures:
   :toctree: generated/

   dynamic

Signaling
=========

.. autosummary::
   :nosignatures:
   :toctree: generated/

   approve
   sleep
   wait_for_input

Scheduling
============================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   CronSchedule
   FixedRate

Notifications
============================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   Email
   PagerDuty
   Slack

Reference Entities
====================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   get_reference_entity
   LaunchPlanReference
   TaskReference
   WorkflowReference
   reference_task
   reference_workflow
   reference_launch_plan

Core Task Types
=================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   SQLTask
   ContainerTask
   PythonFunctionTask
   PythonInstanceTask
   LaunchPlan

Secrets and SecurityContext
============================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   Secret
   SecurityContext


Common Flyte IDL Objects
=========================

.. autosummary::
   :nosignatures:
   :toctree: generated/

   AuthRole
   Labels
   Annotations
   WorkflowExecutionPhase
   Blob
   BlobMetadata
   Literal
   Scalar
   LiteralType
   BlobType

Task Utilities
==============

.. autosummary::
   :nosignatures:
   :toctree: generated/

   HashMethod

Artifacts
=========

.. autosummary::
   :nosignatures:
   :toctree: generated/

   Artifact

Documentation
=============

.. autosummary::
   :nosignatures:
   :toctree: generated/

   Description
   Documentation
   SourceCode
