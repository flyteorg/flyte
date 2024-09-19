from flytekit.models import node_execution as node_execution_models
from flytekit.models.core import catalog, identifier
from tests.flytekit.unit.common_tests.test_workflow_promote import get_compiled_workflow_closure


def test_metadata():
    md = node_execution_models.NodeExecutionMetaData(retry_group="0", is_parent_node=True, spec_node_id="n0")
    md2 = node_execution_models.NodeExecutionMetaData.from_flyte_idl(md.to_flyte_idl())
    assert md == md2


def test_workflow_node_metadata():
    wf_exec_id = identifier.WorkflowExecutionIdentifier("project", "domain", "name")

    obj = node_execution_models.WorkflowNodeMetadata(execution_id=wf_exec_id)
    assert obj.execution_id is wf_exec_id

    obj2 = node_execution_models.WorkflowNodeMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2


def test_task_node_metadata():
    task_id = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version")
    wf_exec_id = identifier.WorkflowExecutionIdentifier("project", "domain", "name")
    node_exec_id = identifier.NodeExecutionIdentifier(
        "node_id",
        wf_exec_id,
    )
    te_id = identifier.TaskExecutionIdentifier(task_id, node_exec_id, 3)
    ds_id = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "t1", "abcdef")
    tag = catalog.CatalogArtifactTag("my-artifact-id", "some name")
    catalog_metadata = catalog.CatalogMetadata(dataset_id=ds_id, artifact_tag=tag, source_task_execution=te_id)

    obj = node_execution_models.TaskNodeMetadata(cache_status=0, catalog_key=catalog_metadata)
    assert obj.cache_status == 0
    assert obj.catalog_key == catalog_metadata

    obj2 = node_execution_models.TaskNodeMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj


def test_dynamic_wf_node_metadata():
    wf_id = identifier.Identifier(identifier.ResourceType.WORKFLOW, "project", "domain", "name", "version")
    cwc = get_compiled_workflow_closure()

    obj = node_execution_models.DynamicWorkflowNodeMetadata(id=wf_id, compiled_workflow=cwc)
    assert obj.id == wf_id
    assert obj.compiled_workflow == cwc

    obj2 = node_execution_models.DynamicWorkflowNodeMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj2 == obj
