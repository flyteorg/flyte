"""
=====================
Remote Access
=====================

.. currentmodule:: flytekit.remote

This module provides utilities for performing operations on tasks, workflows, launchplans, and executions, for example,
the following code fetches and executes a workflow:

.. code-block:: python

    # create a remote object from flyte config and environment variables
    FlyteRemote(config=Config.auto())
    FlyteRemote(config=Config.auto(config_file=....))
    FlyteRemote(config=Config(....))

    # Or if you need to specify a custom cert chain
    # (options and compression are also respected keyword arguments)
    FlyteRemote(private_key=your_private_key_bytes, root_certificates=..., certificate_chain=...)

    # fetch a workflow from the flyte backend
    remote = FlyteRemote(...)
    flyte_workflow = remote.fetch_workflow(name="my_workflow", version="v1")

    # execute the workflow, wait=True will return the execution object after it's completed
    workflow_execution = remote.execute(flyte_workflow, inputs={"a": 1, "b": 10}, wait=True)

    # inspect the execution's outputs
    print(workflow_execution.outputs)

.. _remote-entrypoint:

Entrypoint
==========

.. autosummary::
   :template: custom.rst
   :toctree: generated/
   :nosignatures:

   ~remote.FlyteRemote
   ~remote.Options

.. _remote-flyte-entities:

Entities
========

.. autosummary::
   :template: custom.rst
   :toctree: generated/
   :nosignatures:

   ~entities.FlyteTask
   ~entities.FlyteWorkflow
   ~entities.FlyteLaunchPlan

.. _remote-flyte-entity-components:

Entity Components
=================

.. autosummary::
   :template: custom.rst
   :toctree: generated/
   :nosignatures:

   ~entities.FlyteNode
   ~entities.FlyteTaskNode
   ~entities.FlyteWorkflowNode

.. _remote-flyte-execution-objects:

Execution Objects
=================

.. autosummary::
   :template: custom.rst
   :toctree: generated/
   :nosignatures:

   ~executions.FlyteWorkflowExecution
   ~executions.FlyteTaskExecution
   ~executions.FlyteNodeExecution

"""

from flytekit.remote.entities import (
    FlyteBranchNode,
    FlyteLaunchPlan,
    FlyteNode,
    FlyteTask,
    FlyteTaskNode,
    FlyteWorkflow,
    FlyteWorkflowNode,
)
from flytekit.remote.executions import FlyteNodeExecution, FlyteTaskExecution, FlyteWorkflowExecution
from flytekit.remote.remote import FlyteRemote
