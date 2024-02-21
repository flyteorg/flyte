from dataclasses import dataclass
from enum import Enum
from typing import List, Optional

from flyteidl.core import security_pb2 as _sec

from flytekit.models import common as _common


@dataclass
class Secret(_common.FlyteIdlEntity):
    """
    See :std:ref:`cookbook:secrets` for usage examples.

    Args:
        group is the Name of the secret. For example in kubernetes secrets is the name of the secret
        key is optional and can be an individual secret identifier within the secret For k8s this is required
        version is the version of the secret. This is an optional field
        mount_requirement provides a hint to the system as to how the secret should be injected
    """

    class MountType(Enum):
        ANY = _sec.Secret.MountType.ANY
        """
        Use this if the secret can be injected as either an environment variable / file and this should be left for the
        platform to decide. This is the most flexible option
        """
        ENV_VAR = _sec.Secret.MountType.ENV_VAR
        """
        Use this if the secret can be injected as an environment variable. Usually works for symmetric keys, passwords etc
        """
        FILE = _sec.Secret.MountType.FILE
        """
        Use this for Secrets that cannot be injected into env-var or need to be available as a file
        Caution: May not be supported in all environments
        """

    group: Optional[str] = None
    key: Optional[str] = None
    group_version: Optional[str] = None
    mount_requirement: MountType = MountType.ANY

    def __post_init__(self):
        from flytekit.configuration.plugin import get_plugin

        if get_plugin().secret_requires_group() and self.group is None:
            raise ValueError("Group is a required parameter")

    def to_flyte_idl(self) -> _sec.Secret:
        return _sec.Secret(
            group=self.group,
            group_version=self.group_version,
            key=self.key,
            mount_requirement=self.mount_requirement.value,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _sec.Secret) -> "Secret":
        return cls(
            group=pb2_object.group,
            group_version=pb2_object.group_version if pb2_object.group_version else None,
            key=pb2_object.key if pb2_object.key else None,
            mount_requirement=Secret.MountType(pb2_object.mount_requirement),
        )


@dataclass
class OAuth2Client(_common.FlyteIdlEntity):
    client_id: str
    client_secret: str

    def to_flyte_idl(self) -> _sec.OAuth2Client:
        return _sec.OAuth2Client(
            client_id=self.client_id,
            client_secret=self.client_secret,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _sec.OAuth2Client) -> "OAuth2Client":
        return cls(
            client_id=pb2_object.client_id,
            client_secret=pb2_object.client_secret,
        )


@dataclass
class Identity(_common.FlyteIdlEntity):
    iam_role: Optional[str] = None
    k8s_service_account: Optional[str] = None
    oauth2_client: Optional[OAuth2Client] = None

    def to_flyte_idl(self) -> _sec.Identity:
        return _sec.Identity(
            iam_role=self.iam_role if self.iam_role else None,
            k8s_service_account=self.k8s_service_account if self.k8s_service_account else None,
            oauth2_client=self.oauth2_client.to_flyte_idl() if self.oauth2_client else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _sec.Identity) -> "Identity":
        return cls(
            iam_role=pb2_object.iam_role if pb2_object.iam_role else None,
            k8s_service_account=pb2_object.k8s_service_account if pb2_object.k8s_service_account else None,
            oauth2_client=OAuth2Client.from_flyte_idl(pb2_object.oauth2_client)
            if pb2_object.oauth2_client and pb2_object.oauth2_client.ByteSize()
            else None,
        )


@dataclass
class OAuth2TokenRequest(_common.FlyteIdlEntity):
    class Type(Enum):
        CLIENT_CREDENTIALS = _sec.OAuth2TokenRequest.Type.CLIENT_CREDENTIALS

    name: str
    client: OAuth2Client
    idp_discovery_endpoint: Optional[str] = None
    token_endpoint: Optional[str] = None
    type_: Type = Type.CLIENT_CREDENTIALS

    def to_flyte_idl(self) -> _sec.OAuth2TokenRequest:
        return _sec.OAuth2TokenRequest(
            name=self.name,
            type=self.type_,
            token_endpoint=self.token_endpoint,
            idp_discovery_endpoint=self.idp_discovery_endpoint,
            client=self.client.to_flyte_idl() if self.client else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _sec.OAuth2TokenRequest) -> "OAuth2TokenRequest":
        return cls(
            name=pb2_object.name,
            idp_discovery_endpoint=pb2_object.idp_discovery_endpoint,
            token_endpoint=pb2_object.token_endpoint,
            type_=pb2_object.type,
            client=OAuth2Client.from_flyte_idl(pb2_object.client) if pb2_object.HasField("client") else None,
        )


@dataclass
class SecurityContext(_common.FlyteIdlEntity):
    """
    This is a higher level wrapper object that for the most part users shouldn't have to worry about. You should
    be able to just use :py:class:`flytekit.Secret` instead.
    """

    run_as: Optional[Identity] = None
    secrets: Optional[List[Secret]] = None
    tokens: Optional[List[OAuth2TokenRequest]] = None

    def __post_init__(self):
        if self.secrets and not isinstance(self.secrets, list):
            self.secrets = [self.secrets]
        if self.tokens and not isinstance(self.tokens, list):
            self.tokens = [self.tokens]

    def to_flyte_idl(self) -> _sec.SecurityContext:
        if self.run_as is None and self.secrets is None and self.tokens is None:
            return None
        return _sec.SecurityContext(
            run_as=self.run_as.to_flyte_idl() if self.run_as else None,
            secrets=[s.to_flyte_idl() for s in self.secrets] if self.secrets else None,
            tokens=[t.to_flyte_idl() for t in self.tokens] if self.tokens else None,
        )

    @classmethod
    def from_flyte_idl(cls, pb2_object: _sec.SecurityContext) -> "SecurityContext":
        return cls(
            run_as=Identity.from_flyte_idl(pb2_object.run_as)
            if pb2_object.run_as and pb2_object.run_as.ByteSize() > 0
            else None,
            secrets=[Secret.from_flyte_idl(s) for s in pb2_object.secrets] if pb2_object.secrets else None,
            tokens=[OAuth2TokenRequest.from_flyte_idl(t) for t in pb2_object.tokens] if pb2_object.tokens else None,
        )
