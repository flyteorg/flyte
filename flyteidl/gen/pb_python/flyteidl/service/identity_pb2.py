# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/service/identity.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.api import annotations_pb2 as google_dot_api_dot_annotations__pb2
from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1f\x66lyteidl/service/identity.proto\x12\x10\x66lyteidl.service\x1a\x1cgoogle/api/annotations.proto\x1a\x1cgoogle/protobuf/struct.proto\"\x11\n\x0fUserInfoRequest\"\xa5\x02\n\x10UserInfoResponse\x12\x18\n\x07subject\x18\x01 \x01(\tR\x07subject\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12-\n\x12preferred_username\x18\x03 \x01(\tR\x11preferredUsername\x12\x1d\n\ngiven_name\x18\x04 \x01(\tR\tgivenName\x12\x1f\n\x0b\x66\x61mily_name\x18\x05 \x01(\tR\nfamilyName\x12\x14\n\x05\x65mail\x18\x06 \x01(\tR\x05\x65mail\x12\x18\n\x07picture\x18\x07 \x01(\tR\x07picture\x12\x44\n\x11\x61\x64\x64itional_claims\x18\x08 \x01(\x0b\x32\x17.google.protobuf.StructR\x10\x61\x64\x64itionalClaims2q\n\x0fIdentityService\x12^\n\x08UserInfo\x12!.flyteidl.service.UserInfoRequest\x1a\".flyteidl.service.UserInfoResponse\"\x0b\x82\xd3\xe4\x93\x02\x05\x12\x03/meB\xc5\x01\n\x14\x63om.flyteidl.serviceB\rIdentityProtoP\x01Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service\xa2\x02\x03\x46SX\xaa\x02\x10\x46lyteidl.Service\xca\x02\x10\x46lyteidl\\Service\xe2\x02\x1c\x46lyteidl\\Service\\GPBMetadata\xea\x02\x11\x46lyteidl::Serviceb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.service.identity_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\024com.flyteidl.serviceB\rIdentityProtoP\001Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service\242\002\003FSX\252\002\020Flyteidl.Service\312\002\020Flyteidl\\Service\342\002\034Flyteidl\\Service\\GPBMetadata\352\002\021Flyteidl::Service'
  _IDENTITYSERVICE.methods_by_name['UserInfo']._options = None
  _IDENTITYSERVICE.methods_by_name['UserInfo']._serialized_options = b'\202\323\344\223\002\005\022\003/me'
  _globals['_USERINFOREQUEST']._serialized_start=113
  _globals['_USERINFOREQUEST']._serialized_end=130
  _globals['_USERINFORESPONSE']._serialized_start=133
  _globals['_USERINFORESPONSE']._serialized_end=426
  _globals['_IDENTITYSERVICE']._serialized_start=428
  _globals['_IDENTITYSERVICE']._serialized_end=541
# @@protoc_insertion_point(module_scope)
