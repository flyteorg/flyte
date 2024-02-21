from flyteidl.core import catalog_pb2

from flytekit.models import common as _common_models
from flytekit.models.core import identifier as _identifier


class CatalogArtifactTag(_common_models.FlyteIdlEntity):
    def __init__(self, artifact_id: str, name: str):
        self._artifact_id = artifact_id
        self._name = name

    @property
    def artifact_id(self) -> str:
        return self._artifact_id

    @property
    def name(self) -> str:
        return self._name

    def to_flyte_idl(self) -> catalog_pb2.CatalogArtifactTag:
        return catalog_pb2.CatalogArtifactTag(artifact_id=self.artifact_id, name=self.name)

    @classmethod
    def from_flyte_idl(cls, p: catalog_pb2.CatalogArtifactTag) -> "CatalogArtifactTag":
        return cls(
            artifact_id=p.artifact_id,
            name=p.name,
        )


class CatalogMetadata(_common_models.FlyteIdlEntity):
    def __init__(
        self,
        dataset_id: _identifier.Identifier,
        artifact_tag: CatalogArtifactTag,
        source_task_execution: _identifier.TaskExecutionIdentifier,
    ):
        self._dataset_id = dataset_id
        self._artifact_tag = artifact_tag
        self._source_task_execution = source_task_execution

    @property
    def dataset_id(self) -> _identifier.Identifier:
        return self._dataset_id

    @property
    def artifact_tag(self) -> CatalogArtifactTag:
        return self._artifact_tag

    @property
    def source_task_execution(self) -> _identifier.TaskExecutionIdentifier:
        return self._source_task_execution

    @property
    def source_execution(self) -> _identifier.TaskExecutionIdentifier:
        """
        This is a one of but for now there's only one thing in the one of
        """
        return self._source_task_execution

    def to_flyte_idl(self) -> catalog_pb2.CatalogMetadata:
        return catalog_pb2.CatalogMetadata(
            dataset_id=self.dataset_id.to_flyte_idl(),
            artifact_tag=self.artifact_tag.to_flyte_idl(),
            source_task_execution=self.source_task_execution.to_flyte_idl(),
        )

    @classmethod
    def from_flyte_idl(cls, pb: catalog_pb2.CatalogMetadata) -> "CatalogMetadata":
        return cls(
            dataset_id=_identifier.Identifier.from_flyte_idl(pb.dataset_id),
            artifact_tag=CatalogArtifactTag.from_flyte_idl(pb.artifact_tag),
            # Add HasField check if more things are ever added to the one of
            source_task_execution=_identifier.TaskExecutionIdentifier.from_flyte_idl(pb.source_task_execution),
        )
