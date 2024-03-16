from flytekit.models.core import catalog, identifier


def test_catalog_artifact_tag():
    obj = catalog.CatalogArtifactTag("my-artifact-id", "some name")
    assert obj.artifact_id == "my-artifact-id"
    assert obj.name == "some name"

    obj2 = catalog.CatalogArtifactTag.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
    assert obj2.artifact_id == "my-artifact-id"
    assert obj2.name == "some name"


def test_catalog_metadata():
    task_id = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "name", "version")
    wf_exec_id = identifier.WorkflowExecutionIdentifier("project", "domain", "name")
    node_exec_id = identifier.NodeExecutionIdentifier(
        "node_id",
        wf_exec_id,
    )
    te_id = identifier.TaskExecutionIdentifier(task_id, node_exec_id, 3)
    ds_id = identifier.Identifier(identifier.ResourceType.TASK, "project", "domain", "t1", "abcdef")
    tag = catalog.CatalogArtifactTag("my-artifact-id", "some name")
    obj = catalog.CatalogMetadata(dataset_id=ds_id, artifact_tag=tag, source_task_execution=te_id)
    assert obj.dataset_id is ds_id
    assert obj.source_execution is te_id
    assert obj.source_task_execution is te_id
    assert obj.artifact_tag is tag

    obj2 = catalog.CatalogMetadata.from_flyte_idl(obj.to_flyte_idl())
    assert obj == obj2
