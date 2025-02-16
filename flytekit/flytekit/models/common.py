import abc as _abc
import json as _json
import re
from typing import Dict

from flyteidl.admin import common_pb2 as _common_pb2
from flyteidl.core import literals_pb2 as _literals_pb2
from google.protobuf import json_format as _json_format
from google.protobuf import struct_pb2 as _struct


class FlyteABCMeta(_abc.ABCMeta):
    def __instancecheck__(cls, instance):
        if cls in type(instance).__mro__:
            return True
        return super(FlyteABCMeta, cls).__instancecheck__(instance)


class FlyteType(FlyteABCMeta):
    def __repr__(cls):
        return cls.short_class_string()

    def __str__(cls):
        return cls.verbose_class_string()

    def short_class_string(cls):
        """
        :rtype: Text
        """
        return super(FlyteType, cls).__repr__()

    def verbose_class_string(cls):
        """
        :rtype: Text
        """
        return cls.short_class_string()

    @_abc.abstractmethod
    def from_flyte_idl(cls, idl_object):
        pass


class FlyteIdlEntity(object, metaclass=FlyteType):
    def __eq__(self, other):
        return isinstance(other, FlyteIdlEntity) and other.to_flyte_idl() == self.to_flyte_idl()

    def __ne__(self, other):
        return not (self == other)

    def __repr__(self):
        return self.short_string()

    def __str__(self):
        return self.verbose_string()

    def __hash__(self):
        return hash(self.to_flyte_idl().SerializeToString(deterministic=True))

    def short_string(self):
        """
        :rtype: Text
        """
        literal_str = re.sub(r"\s+", " ", str(self.to_flyte_idl())).strip()
        return f"<FlyteLiteral {literal_str}>"

    def verbose_string(self):
        """
        :rtype: Text
        """
        return self.short_string()

    def serialize_to_string(self) -> str:
        return self.to_flyte_idl().SerializeToString()

    @property
    def is_empty(self):
        return len(self.to_flyte_idl().SerializeToString()) == 0

    @_abc.abstractmethod
    def to_flyte_idl(self):
        pass


class FlyteCustomIdlEntity(FlyteIdlEntity):
    @classmethod
    def from_flyte_idl(cls, idl_object):
        """

        :param _struct.Struct idl_object:
        :return: FlyteCustomIdlEntity
        """
        return cls.from_dict(idl_dict=_json_format.MessageToDict(idl_object))

    def to_flyte_idl(self):
        return _json_format.Parse(_json.dumps(self.to_dict()), _struct.Struct())

    @_abc.abstractmethod
    def from_dict(self, idl_dict):
        pass

    @_abc.abstractmethod
    def to_dict(self):
        """
        Converts self to a dictionary.
        :rtype: dict[Text, T]
        """
        pass


class NamedEntityIdentifier(FlyteIdlEntity):
    def __init__(self, project, domain, name=None):
        """
        :param Text project: The name of the project in which this entity lives.
        :param Text domain: The name of the domain within the project.
        :param Text name: [Optional] The name of the entity within the namespace of the project and domain.
        """
        self._project = project
        self._domain = domain
        self._name = name

    @property
    def project(self):
        """
        The name of the project in which this entity lives.
        :rtype: Text
        """
        return self._project

    @property
    def domain(self):
        """
        The name of the domain within the project.
        :rtype: Text
        """
        return self._domain

    @property
    def name(self):
        """
        The name of the entity within the namespace of the project and domain.
        :rtype: Text
        """
        return self._name

    def to_flyte_idl(self):
        """
        Stores object to a Flyte-IDL defined protobuf.
        :rtype: flyteidl.admin.common_pb2.NamedEntityIdentifier
        """

        # We use the kwarg constructor of the protobuf and setting name=None is equivalent to not setting it at all
        return _common_pb2.NamedEntityIdentifier(project=self.project, domain=self.domain, name=self.name)

    @classmethod
    def from_flyte_idl(cls, idl_object):
        """
        :param flyteidl.admin.common_pb2.NamedEntityIdentifier idl_object:
        :rtype: NamedEntityIdentifier
        """
        return cls(idl_object.project, idl_object.domain, idl_object.name)


class EmailNotification(FlyteIdlEntity):
    def __init__(self, recipients_email):
        """
        :param list[Text] recipients_email:
        """
        self._recipients_email = recipients_email

    @property
    def recipients_email(self):
        """
        :rtype: list[Text]
        """
        return self._recipients_email

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.EmailNotification
        """
        return _common_pb2.EmailNotification(recipients_email=self.recipients_email)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.common_pb2.EmailNotification pb2_object:
        :rtype: EmailNotification
        """
        return cls(pb2_object.recipients_email)


class SlackNotification(FlyteIdlEntity):
    def __init__(self, recipients_email):
        """
        :param list[Text] recipients_email:
        """
        self._recipients_email = recipients_email

    @property
    def recipients_email(self):
        """
        :rtype: list[Text]
        """
        return self._recipients_email

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.SlackNotification
        """
        return _common_pb2.SlackNotification(recipients_email=self.recipients_email)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.common_pb2.SlackNotification pb2_object:
        :rtype: EmailNotification
        """
        return cls(pb2_object.recipients_email)


class PagerDutyNotification(FlyteIdlEntity):
    def __init__(self, recipients_email):
        """
        :param list[Text] recipients_email:
        """
        self._recipients_email = recipients_email

    @property
    def recipients_email(self):
        """
        :rtype: list[Text]
        """
        return self._recipients_email

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.PagerDutyNotification
        """
        return _common_pb2.PagerDutyNotification(recipients_email=self.recipients_email)

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.common_pb2.PagerDutyNotification pb2_object:
        :rtype: EmailNotification
        """
        return cls(pb2_object.recipients_email)


class Notification(FlyteIdlEntity):
    def __init__(
        self,
        phases,
        email: EmailNotification = None,
        pager_duty: PagerDutyNotification = None,
        slack: SlackNotification = None,
    ):
        """
        Represents a structure for notifications based on execution status.

        :param list[int] phases: A list of phases to which users can associate the notifications.
        :param EmailNotification email: [Optional] Specify this for an email notification.
        :param PagerDutyNotification email: [Optional] Specify this for a PagerDuty notification.
        :param SlackNotification email: [Optional] Specify this for a Slack notification.
        """
        self._phases = phases
        self._email = email
        self._pager_duty = pager_duty
        self._slack = slack

    @property
    def phases(self):
        """
        A list of phases to which users can associate the notifications.
        :rtype: list[int]
        """
        return self._phases

    @property
    def email(self):
        """
        :rtype: EmailNotification
        """
        return self._email

    @property
    def pager_duty(self):
        """
        :rtype: PagerDutyNotification
        """
        return self._pager_duty

    @property
    def slack(self):
        """
        :rtype: SlackNotification
        """
        return self._slack

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.Notification
        """
        return _common_pb2.Notification(
            phases=self.phases,
            email=self.email.to_flyte_idl() if self.email else None,
            pager_duty=self.pager_duty.to_flyte_idl() if self.pager_duty else None,
            slack=self.slack.to_flyte_idl() if self.slack else None,
        )

    @classmethod
    def from_flyte_idl(cls, p):
        """
        :param flyteidl.admin.common_pb2.Notification p:
        :rtype: Notification
        """
        return cls(
            p.phases,
            email=EmailNotification.from_flyte_idl(p.email) if p.HasField("email") else None,
            pager_duty=PagerDutyNotification.from_flyte_idl(p.pager_duty) if p.HasField("pager_duty") else None,
            slack=SlackNotification.from_flyte_idl(p.slack) if p.HasField("slack") else None,
        )


class Labels(FlyteIdlEntity):
    def __init__(self, values):
        """
        Label values to be applied to a workflow execution resource.

        :param dict[Text, Text] values:
        """
        self._values = values

    @property
    def values(self):
        return self._values

    def to_flyte_idl(self):
        """
        :rtype: dict[Text, Text]
        """
        return _common_pb2.Labels(values={k: v for k, v in self.values.items()})

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.common_pb2.Labels pb2_object:
        :rtype: Labels
        """
        return cls({k: v for k, v in pb2_object.values.items()})


class Annotations(FlyteIdlEntity):
    def __init__(self, values):
        """
        Annotation values to be applied to a workflow execution resource.

        :param dict[Text, Text] values:
        """
        self._values = values

    @property
    def values(self):
        return self._values

    def to_flyte_idl(self):
        """
        :rtype: _common_pb2.Annotations
        """
        return _common_pb2.Annotations(values={k: v for k, v in self.values.items()})

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.common_pb2.Annotations pb2_object:
        :rtype: Annotations
        """
        return cls({k: v for k, v in pb2_object.values.items()})


class UrlBlob(FlyteIdlEntity):
    def __init__(self, url, bytes):
        """
        :param Text url:
        :param int bytes:
        """
        self._url = url
        self._bytes = bytes

    @property
    def url(self):
        """
        :rtype: Text
        """
        return self._url

    @property
    def bytes(self):
        """
        :rtype: int
        """
        return self._bytes

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.UrlBlob
        """
        return _common_pb2.UrlBlob(url=self.url, bytes=self.bytes)

    @classmethod
    def from_flyte_idl(cls, pb):
        """
        :param flyteidl.admin.common_pb2.UrlBlob pb:
        :rtype: UrlBlob
        """
        return cls(url=pb.url, bytes=pb.bytes)


class AuthRole(FlyteIdlEntity):
    def __init__(self, assumable_iam_role=None, kubernetes_service_account=None):
        """Auth configuration for IAM or K8s service account.

        Either one or both of the assumable IAM role and/or the K8s service account can be set.

        :param Text assumable_iam_role: IAM identity with set permissions policies.
        :param Text kubernetes_service_account: Provides an identity for workflow execution resources.
          Flyte deployment administrators are responsible for handling permissions as they
          relate to the service account.
        """
        self._assumable_iam_role = assumable_iam_role
        self._kubernetes_service_account = kubernetes_service_account

    @property
    def assumable_iam_role(self):
        """
        The IAM role to execute the workflow with
        :rtype: Text
        """
        return self._assumable_iam_role

    @property
    def kubernetes_service_account(self):
        """
        The kubernetes service account to execute the workflow with
        :rtype: Text
        """
        return self._kubernetes_service_account

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.launch_plan_pb2.Auth
        """
        return _common_pb2.AuthRole(
            assumable_iam_role=self.assumable_iam_role if self.assumable_iam_role else None,
            kubernetes_service_account=self.kubernetes_service_account if self.kubernetes_service_account else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object):
        """
        :param flyteidl.admin.launch_plan_pb2.Auth pb2_object:
        :rtype: Auth
        """
        return cls(
            assumable_iam_role=pb2_object.assumable_iam_role,
            kubernetes_service_account=pb2_object.kubernetes_service_account,
        )


class RawOutputDataConfig(FlyteIdlEntity):
    def __init__(self, output_location_prefix):
        """
        :param Text output_location_prefix: Location of offloaded data for things like S3, etc.
        """
        self._output_location_prefix = output_location_prefix

    @property
    def output_location_prefix(self):
        return self._output_location_prefix

    def to_flyte_idl(self):
        """
        :rtype: flyteidl.admin.common_pb2.Auth
        """
        return _common_pb2.RawOutputDataConfig(output_location_prefix=self.output_location_prefix)

    @classmethod
    def from_flyte_idl(cls, pb2):
        return cls(output_location_prefix=pb2.output_location_prefix)


class Envs(FlyteIdlEntity):
    def __init__(self, envs: Dict[str, str]):
        self._envs = envs

    @property
    def envs(self) -> Dict[str, str]:
        return self._envs

    def to_flyte_idl(self) -> _common_pb2.Envs:
        return _common_pb2.Envs(values=[_literals_pb2.KeyValuePair(key=k, value=v) for k, v in self.envs.items()])

    @classmethod
    def from_flyte_idl(cls, pb2: _common_pb2.Envs) -> _common_pb2.Envs:
        return cls(envs={kv.key: kv.value for kv in pb2.values})
