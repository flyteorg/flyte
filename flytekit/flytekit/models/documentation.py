from dataclasses import dataclass
from enum import Enum
from typing import Optional

from flyteidl.admin import description_entity_pb2

from flytekit.models import common as _common_models


@dataclass
class Description(_common_models.FlyteIdlEntity):
    """
    Full user description with formatting preserved. This can be rendered
    by clients, such as the console or command line tools with in-tact
    formatting.
    """

    class DescriptionFormat(Enum):
        UNKNOWN = 0
        MARKDOWN = 1
        HTML = 2
        RST = 3

    value: Optional[str] = None
    uri: Optional[str] = None
    icon_link: Optional[str] = None
    format: DescriptionFormat = DescriptionFormat.RST

    def to_flyte_idl(self):
        return description_entity_pb2.Description(
            value=self.value if self.value else None,
            uri=self.uri if self.uri else None,
            format=self.format.value,
            icon_link=self.icon_link,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: description_entity_pb2.Description) -> "Description":
        return cls(
            value=pb2_object.value if pb2_object.value else None,
            uri=pb2_object.uri if pb2_object.uri else None,
            format=Description.DescriptionFormat(pb2_object.format),
            icon_link=pb2_object.icon_link if pb2_object.icon_link else None,
        )


@dataclass
class SourceCode(_common_models.FlyteIdlEntity):
    """
    Link to source code used to define this task or workflow.
    """

    link: Optional[str] = None

    def to_flyte_idl(self):
        return description_entity_pb2.SourceCode(link=self.link)

    @classmethod
    def from_flyte_idl(cls, pb2_object: description_entity_pb2.SourceCode) -> "SourceCode":
        return cls(link=pb2_object.link) if pb2_object.link else None


@dataclass
class Documentation(_common_models.FlyteIdlEntity):
    """
    DescriptionEntity contains detailed description for the task/workflow/launch plan.
    Documentation could provide insight into the algorithms, business use case, etc.
    Args:
        short_description (str): One-liner overview of the entity.
        long_description (Optional[Description]): Full user description with formatting preserved.
        source_code (Optional[SourceCode]): link to source code used to define this entity
    """

    short_description: Optional[str] = None
    long_description: Optional[Description] = None
    source_code: Optional[SourceCode] = None

    def to_flyte_idl(self):
        return description_entity_pb2.DescriptionEntity(
            short_description=self.short_description,
            long_description=self.long_description.to_flyte_idl() if self.long_description else None,
            source_code=self.source_code.to_flyte_idl() if self.source_code else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: description_entity_pb2.DescriptionEntity) -> "Documentation":
        return cls(
            short_description=pb2_object.short_description,
            long_description=Description.from_flyte_idl(pb2_object.long_description)
            if pb2_object.long_description
            else None,
            source_code=SourceCode.from_flyte_idl(pb2_object.source_code) if pb2_object.source_code else None,
        )
