from flytekit.models.core import identifier


def test_identifier():
    obj = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version")
    assert obj.project == "project"
    assert obj.domain == "domain"
    assert obj.name == "name"
    assert obj.version == "version"
    assert obj.resource_type == identifier.ResourceType.TASK
    assert obj.resource_type_name() == "TASK"

    obj2 = identifier.Identifier.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.project == "project"
    assert obj2.domain == "domain"
    assert obj2.name == "name"
    assert obj2.version == "version"
    assert obj2.resource_type == identifier.ResourceType.TASK
    assert obj2.resource_type_name() == "TASK"


def test_node_execution_identifier():
    wf_exec_id = identifier.WorkflowExecutionIdentifier("project", "domain", "name")
    obj = identifier.NodeExecutionIdentifier("node_id", wf_exec_id)
    assert obj.node_id == "node_id"
    assert obj.execution_id == wf_exec_id

    obj2 = identifier.NodeExecutionIdentifier.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.node_id == "node_id"
    assert obj2.execution_id == wf_exec_id


def test_task_execution_identifier():
    task_id = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version")
    wf_exec_id = identifier.WorkflowExecutionIdentifier("project", "domain", "name")
    node_exec_id = identifier.NodeExecutionIdentifier(
        "node_id",
        wf_exec_id,
    )
    obj = identifier.TaskExecutionIdentifier(task_id, node_exec_id, 3)
    assert obj.retry_attempt == 3
    assert obj.task_id == task_id
    assert obj.node_execution_id == node_exec_id

    obj2 = identifier.TaskExecutionIdentifier.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.retry_attempt == 3
    assert obj2.task_id == task_id
    assert obj2.node_execution_id == node_exec_id


def test_workflow_execution_identifier():
    obj = identifier.WorkflowExecutionIdentifier("project", "domain", "name")
    assert obj.project == "project"
    assert obj.domain == "domain"
    assert obj.name == "name"

    obj2 = identifier.WorkflowExecutionIdentifier.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
    assert obj2.project == "project"
    assert obj2.domain == "domain"
    assert obj2.name == "name"


def test_identifier_emptiness():
    empty_id = identifier.Identifier(identifier.ResourceType.UNSPECIFIED, "", "", "", "")
    not_empty_id = identifier.Identifier(identifier.ResourceType.UNSPECIFIED, "", "", "", "version")
    assert empty_id.is_empty
    assert not not_empty_id.is_empty
    assert not_empty_id.resource_type_name() == "UNSPECIFIED"
