# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/core/security.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1c\x66lyteidl/core/security.proto\x12\rflyteidl.core\"\xe9\x01\n\x06Secret\x12\x14\n\x05group\x18\x01 \x01(\tR\x05group\x12#\n\rgroup_version\x18\x02 \x01(\tR\x0cgroupVersion\x12\x10\n\x03key\x18\x03 \x01(\tR\x03key\x12L\n\x11mount_requirement\x18\x04 \x01(\x0e\x32\x1f.flyteidl.core.Secret.MountTypeR\x10mountRequirement\x12\x17\n\x07\x65nv_var\x18\x05 \x01(\tR\x06\x65nvVar\"+\n\tMountType\x12\x07\n\x03\x41NY\x10\x00\x12\x0b\n\x07\x45NV_VAR\x10\x01\x12\x08\n\x04\x46ILE\x10\x02\"g\n\x0cOAuth2Client\x12\x1b\n\tclient_id\x18\x01 \x01(\tR\x08\x63lientId\x12:\n\rclient_secret\x18\x02 \x01(\x0b\x32\x15.flyteidl.core.SecretR\x0c\x63lientSecret\"\xc6\x01\n\x08Identity\x12\x19\n\x08iam_role\x18\x01 \x01(\tR\x07iamRole\x12.\n\x13k8s_service_account\x18\x02 \x01(\tR\x11k8sServiceAccount\x12@\n\roauth2_client\x18\x03 \x01(\x0b\x32\x1b.flyteidl.core.OAuth2ClientR\x0coauth2Client\x12-\n\x12\x65xecution_identity\x18\x04 \x01(\tR\x11\x65xecutionIdentity\"\x96\x02\n\x12OAuth2TokenRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12:\n\x04type\x18\x02 \x01(\x0e\x32&.flyteidl.core.OAuth2TokenRequest.TypeR\x04type\x12\x33\n\x06\x63lient\x18\x03 \x01(\x0b\x32\x1b.flyteidl.core.OAuth2ClientR\x06\x63lient\x12\x34\n\x16idp_discovery_endpoint\x18\x04 \x01(\tR\x14idpDiscoveryEndpoint\x12%\n\x0etoken_endpoint\x18\x05 \x01(\tR\rtokenEndpoint\"\x1e\n\x04Type\x12\x16\n\x12\x43LIENT_CREDENTIALS\x10\x00\"\xad\x01\n\x0fSecurityContext\x12.\n\x06run_as\x18\x01 \x01(\x0b\x32\x17.flyteidl.core.IdentityR\x05runAs\x12/\n\x07secrets\x18\x02 \x03(\x0b\x32\x15.flyteidl.core.SecretR\x07secrets\x12\x39\n\x06tokens\x18\x03 \x03(\x0b\x32!.flyteidl.core.OAuth2TokenRequestR\x06tokensB\xb3\x01\n\x11\x63om.flyteidl.coreB\rSecurityProtoP\x01Z:github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core\xa2\x02\x03\x46\x43X\xaa\x02\rFlyteidl.Core\xca\x02\rFlyteidl\\Core\xe2\x02\x19\x46lyteidl\\Core\\GPBMetadata\xea\x02\x0e\x46lyteidl::Coreb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.core.security_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\021com.flyteidl.coreB\rSecurityProtoP\001Z:github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core\242\002\003FCX\252\002\rFlyteidl.Core\312\002\rFlyteidl\\Core\342\002\031Flyteidl\\Core\\GPBMetadata\352\002\016Flyteidl::Core'
  _globals['_SECRET']._serialized_start=48
  _globals['_SECRET']._serialized_end=281
  _globals['_SECRET_MOUNTTYPE']._serialized_start=238
  _globals['_SECRET_MOUNTTYPE']._serialized_end=281
  _globals['_OAUTH2CLIENT']._serialized_start=283
  _globals['_OAUTH2CLIENT']._serialized_end=386
  _globals['_IDENTITY']._serialized_start=389
  _globals['_IDENTITY']._serialized_end=587
  _globals['_OAUTH2TOKENREQUEST']._serialized_start=590
  _globals['_OAUTH2TOKENREQUEST']._serialized_end=868
  _globals['_OAUTH2TOKENREQUEST_TYPE']._serialized_start=838
  _globals['_OAUTH2TOKENREQUEST_TYPE']._serialized_end=868
  _globals['_SECURITYCONTEXT']._serialized_start=871
  _globals['_SECURITYCONTEXT']._serialized_end=1044
# @@protoc_insertion_point(module_scope)
