from flyteidl.admin import common_pb2 as _common

from flytekit.models import common as _common_models


class NamedEntityState(object):
    ACTIVE = _common.NAMED_ENTITY_ACTIVE
    ARCHIVED = _common.NAMED_ENTITY_ARCHIVED

    @classmethod
    def enum_to_string(cls, val):
        """
        :param int val:
        :rtype: Text
        """
        if val == cls.ACTIVE:
            return "ACTIVE"
        elif val == cls.ARCHIVED:
            return "ARCHIVED"
        else:
            return "<UNKNOWN>"


class NamedEntityIdentifier(_common_models.FlyteIdlEntity):
    def __init__(self, project, domain, name):
        """
        :param Text project:
        :param Text domain:
        :param Text name:
        """
        self._project = project
        self._domain = domain
        self._name = name

    @property
    def project(self):
        """
        :rtype: Text
        """
        return self._project

    @property
    def domain(self):
        """
        :rtype: Text
        """
        return self._domain

    @property
    def name(self):
        """
        :rtype: Text
        """
        return self._name

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.NamedEntityIdentifier
        """
        return _common.NamedEntityIdentifier(
            project=self.project,
            domain=self.domain,
            name=self.name,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.common_pb2.NamedEntityIdentifier p:
        :rtype: Identifier
        """
        return cls(
            project=p.project,
            domain=p.domain,
            name=p.name,
        )


class NamedEntityMetadata(_common_models.FlyteIdlEntity):
    def __init__(self, description, state):
        """

        :param Text description:
        :param int state: enum value from NamedEntityState
        """
        self._description = description
        self._state = state

    @property
    def description(self):
        """
        :rtype: Text
        """
        return self._description

    @property
    def state(self):
        """
        enum value from NamedEntityState
        :rtype: int
        """
        return self._state

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.NamedEntityMetadata
        """
        return _common.NamedEntityMetadata(
            description=self.description,
            state=self.state,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.core.common_pb2.NamedEntityMetadata p:
        :rtype: Identifier
        """
        return cls(
            description=p.description,
            state=p.state,
        )
