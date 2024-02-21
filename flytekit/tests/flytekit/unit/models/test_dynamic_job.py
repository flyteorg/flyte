from datetime import timedelta as _timedelta
from itertools import product

import pytest
from google.protobuf import text_format

from flytekit.models import array_job as _array_job
from flytekit.models import dynamic_job as _dynamic_job
from flytekit.models import literals as _literals
from flytekit.models import task as _task
from flytekit.models.core import identifier as _identifier
from flytekit.models.core import workflow as _workflow
from tests.flytekit.common import parameterizers

LIST_OF_DYNAMIC_TASKS = [
    _task.TaskTemplate(
        _identifier.Identifier(_identifier.ResourceType.TASK, "p", "d", "n", "v"),
        "python",
        task_metadata,
        interfaces,
        _array_job.ArrayJob(2, 2, 2).to_dict(),
        container=_task.Container(
            "my_image",
            ["this", "is", "a", "cmd"],
            ["this", "is", "an", "arg"],
            resources,
            {"a": "b"},
            {"d": "e"},
        ),
    )
    for task_metadata, interfaces, resources in product(
        parameterizers.LIST_OF_TASK_METADATA,
        parameterizers.LIST_OF_INTERFACES,
        parameterizers.LIST_OF_RESOURCES,
    )
]


@pytest.mark.parametrize("task", LIST_OF_DYNAMIC_TASKS)
def test_future_task_document(task):
    rs = _literals.RetryStrategy(0)
    nm = _workflow.NodeMetadata("node-name", _timedelta(minutes=10), rs)
    n = _workflow.Node(
        id="id",
        metadata=nm,
        inputs=[],
        upstream_node_ids=[],
        output_aliases=[],
        task_node=_workflow.TaskNode(task.id),
    )
    n.to_flyte_idl()
    doc = _dynamic_job.DynamicJobSpec(
        tasks=[task],
        nodes=[n],
        min_successes=1,
        outputs=[_literals.Binding("var", _literals.BindingData())],
        subworkflows=[],
    )
    assert text_format.MessageToString(doc.to_flyte_idl()) == text_format.MessageToString(
        _dynamic_job.DynamicJobSpec.from_flyte_idl(doc.to_flyte_idl()).to_flyte_idl()
    )
