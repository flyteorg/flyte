from __future__ import annotations

from typing import Dict, Type

from flytekit.core.launch_plan import ReferenceLaunchPlan
from flytekit.core.task import ReferenceTask
from flytekit.core.workflow import ReferenceWorkflow
from flytekit.exceptions.user import FlyteValidationException
from flytekit.models.core import identifier as _identifier_model


def get_reference_entity(
    resource_type: int,
    project: str,
    domain: str,
    name: str,
    version: str,
    inputs: Dict[str, Type],
    outputs: Dict[str, Type],
):
    """
    See the documentation for :py:class:`flytekit.reference_task` and :py:class:`flytekit.reference_workflow` as well.

    This function is the general form of the two aforementioned functions. It's better for programmatic usage, as
    the interface is passed in as arguments instead of analyzed from type annotations.

    .. literalinclude:: ../../../tests/flytekit/unit/core/test_references.py
       :start-after: # docs_ref_start
       :end-before: # docs_ref_end
       :language: python
       :dedent: 4

    :param resource_type: This is the type of entity it is. Must be one of
      :py:class:`flytekit.models.core.identifier.ResourceType`
    :param project: The project the entity you're looking for has been registered in.
    :param domain: The domain the entity you're looking for has been registered in.
    :param name: The name of the registered entity
    :param version: The version the entity you're looking for has been registered with.
    :param inputs: An ordered dictionary of input names as strings to their Python types.
    :param outputs: An ordered dictionary of output names as strings to their Python types.
    :return:
    """
    if resource_type == _identifier_model.ResourceType.TASK:
        return ReferenceTask(project, domain, name, version, inputs, outputs)
    elif resource_type == _identifier_model.ResourceType.WORKFLOW:
        return ReferenceWorkflow(project, domain, name, version, inputs, outputs)
    elif resource_type == _identifier_model.ResourceType.LAUNCH_PLAN:
        return ReferenceLaunchPlan(project, domain, name, version, inputs, outputs)
    else:
        raise FlyteValidationException("Resource type must be one of task, workflow, or launch plan")
