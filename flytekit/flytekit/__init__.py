"""
=====================
Core Flytekit
=====================

.. currentmodule:: flytekit

This package contains all of the most common abstractions you'll need to write Flyte workflows and extend Flytekit.

Basic Authoring
===============

These are the essentials needed to get started writing tasks and workflows.

.. autosummary::
   :nosignatures:
   :template: custom.rst
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
   :template: custom.rst
   :toctree: generated/

   conditional


Customizing Tasks & Workflows
==============================

.. autosummary::
   :nosignatures:
   :template: custom.rst
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
   :template: custom.rst
   :toctree: generated/

   dynamic

Signaling
=========

.. autosummary::
   :nosignatures:
   :template: custom.rst
   :toctree: generated/

   approve
   sleep
   wait_for_input

Scheduling
============================

.. autosummary::
   :nosignatures:
   :template: custom.rst
   :toctree: generated/

   CronSchedule
   FixedRate

Notifications
============================

.. autosummary::
   :nosignatures:
   :template: custom.rst
   :toctree: generated/

   Email
   PagerDuty
   Slack

Reference Entities
====================

.. autosummary::
   :nosignatures:
   :template: custom.rst
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
   :template: custom.rst
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
   :template: custom.rst
   :toctree: generated/

   Secret
   SecurityContext


Common Flyte IDL Objects
=========================

.. autosummary::
   :nosignatures:
   :template: custom.rst
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
   :template: custom.rst
   :toctree: generated/

   HashMethod

Documentation
=============

.. autosummary::
   :nosignatures:
   :template: custom.rst
   :toctree: generated/

   Description
   Documentation
   SourceCode

"""
import os
import sys
from typing import Generator

from rich import traceback

from flytekit.lazy_import.lazy_module import lazy_module

if sys.version_info < (3, 10):
    from importlib_metadata import entry_points
else:
    from importlib.metadata import entry_points

from flytekit._version import __version__
from flytekit.core.base_sql_task import SQLTask
from flytekit.core.base_task import SecurityContext, TaskMetadata, kwtypes
from flytekit.core.checkpointer import Checkpoint
from flytekit.core.condition import conditional
from flytekit.core.container_task import ContainerTask
from flytekit.core.context_manager import ExecutionParameters, FlyteContext, FlyteContextManager
from flytekit.core.dynamic_workflow_task import dynamic
from flytekit.core.gate import approve, sleep, wait_for_input
from flytekit.core.hash import HashMethod
from flytekit.core.launch_plan import LaunchPlan, reference_launch_plan
from flytekit.core.map_task import map_task
from flytekit.core.notification import Email, PagerDuty, Slack
from flytekit.core.pod_template import PodTemplate
from flytekit.core.python_function_task import PythonFunctionTask, PythonInstanceTask
from flytekit.core.reference import get_reference_entity
from flytekit.core.reference_entity import LaunchPlanReference, TaskReference, WorkflowReference
from flytekit.core.resources import Resources
from flytekit.core.schedule import CronSchedule, FixedRate
from flytekit.core.task import Secret, reference_task, task
from flytekit.core.type_engine import BatchSize
from flytekit.core.workflow import ImperativeWorkflow as Workflow
from flytekit.core.workflow import WorkflowFailurePolicy, reference_workflow, workflow
from flytekit.deck import Deck
from flytekit.image_spec import ImageSpec
from flytekit.loggers import LOGGING_RICH_FMT_ENV_VAR, logger
from flytekit.models.common import Annotations, AuthRole, Labels
from flytekit.models.core.execution import WorkflowExecutionPhase
from flytekit.models.core.types import BlobType
from flytekit.models.documentation import Description, Documentation, SourceCode
from flytekit.models.literals import Blob, BlobMetadata, Literal, Scalar
from flytekit.models.types import LiteralType
from flytekit.sensor.sensor_engine import SensorEngine
from flytekit.types import directory, file, iterator
from flytekit.types.structured.structured_dataset import (
    StructuredDataset,
    StructuredDatasetFormat,
    StructuredDatasetTransformerEngine,
    StructuredDatasetType,
)


def current_context() -> ExecutionParameters:
    """
    Use this method to get a handle of specific parameters available in a flyte task.

    Usage

    .. code-block:: python

        flytekit.current_context().logging.info(...)

    Available params are documented in :py:class:`flytekit.core.context_manager.ExecutionParams`.
    There are some special params, that should be available
    """
    return FlyteContextManager.current_context().execution_state.user_space_params


def new_context() -> Generator[FlyteContext, None, None]:
    return FlyteContextManager.with_context(FlyteContextManager.current_context().new_builder())


def load_implicit_plugins():
    """
    This method allows loading all plugins that have the entrypoint specification. This uses the plugin loading
    behavior as explained `here <>`_.

    This is an opt in system and plugins that have an implicit loading requirement should add the implicit loading
    entrypoint specification to their setup.py. The following example shows how we can autoload a module called fsspec
    (whose init files contains the necessary plugin registration step)

    .. code-block::

        # note the group is always ``flytekit.plugins``
        setup(
        ...
        entry_points={'flytekit.plugins': 'fsspec=flytekitplugins.fsspec'},
        ...
        )

    This works as long as the fsspec module has

    .. code-block::

       # For data persistence plugins
       DataPersistencePlugins.register_plugin(f"{k}://", FSSpecPersistence, force=True)
       # OR for type plugins
       TypeEngine.register(PanderaTransformer())
       # etc

    """
    discovered_plugins = entry_points(group="flytekit.plugins")
    for p in discovered_plugins:
        p.load()


# Load all implicit plugins
load_implicit_plugins()

# Pretty-print exception messages
if os.environ.get(LOGGING_RICH_FMT_ENV_VAR) != "0":
    traceback.install(width=None, extra_lines=0)
