# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/admin/project_attributes.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.admin import matchable_resource_pb2 as flyteidl_dot_admin_dot_matchable__resource__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\'flyteidl/admin/project_attributes.proto\x12\x0e\x66lyteidl.admin\x1a\'flyteidl/admin/matchable_resource.proto\"\x82\x01\n\x11ProjectAttributes\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12S\n\x13matching_attributes\x18\x02 \x01(\x0b\x32\".flyteidl.admin.MatchingAttributesR\x12matchingAttributes\"c\n\x1eProjectAttributesUpdateRequest\x12\x41\n\nattributes\x18\x01 \x01(\x0b\x32!.flyteidl.admin.ProjectAttributesR\nattributes\"!\n\x1fProjectAttributesUpdateResponse\"\x7f\n\x1bProjectAttributesGetRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x46\n\rresource_type\x18\x02 \x01(\x0e\x32!.flyteidl.admin.MatchableResourceR\x0cresourceType\"a\n\x1cProjectAttributesGetResponse\x12\x41\n\nattributes\x18\x01 \x01(\x0b\x32!.flyteidl.admin.ProjectAttributesR\nattributes\"\x82\x01\n\x1eProjectAttributesDeleteRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x46\n\rresource_type\x18\x02 \x01(\x0e\x32!.flyteidl.admin.MatchableResourceR\x0cresourceType\"!\n\x1fProjectAttributesDeleteResponseB\xc2\x01\n\x12\x63om.flyteidl.adminB\x16ProjectAttributesProtoP\x01Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\xa2\x02\x03\x46\x41X\xaa\x02\x0e\x46lyteidl.Admin\xca\x02\x0e\x46lyteidl\\Admin\xe2\x02\x1a\x46lyteidl\\Admin\\GPBMetadata\xea\x02\x0f\x46lyteidl::Adminb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.admin.project_attributes_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\022com.flyteidl.adminB\026ProjectAttributesProtoP\001Z;github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin\242\002\003FAX\252\002\016Flyteidl.Admin\312\002\016Flyteidl\\Admin\342\002\032Flyteidl\\Admin\\GPBMetadata\352\002\017Flyteidl::Admin'
  _globals['_PROJECTATTRIBUTES']._serialized_start=101
  _globals['_PROJECTATTRIBUTES']._serialized_end=231
  _globals['_PROJECTATTRIBUTESUPDATEREQUEST']._serialized_start=233
  _globals['_PROJECTATTRIBUTESUPDATEREQUEST']._serialized_end=332
  _globals['_PROJECTATTRIBUTESUPDATERESPONSE']._serialized_start=334
  _globals['_PROJECTATTRIBUTESUPDATERESPONSE']._serialized_end=367
  _globals['_PROJECTATTRIBUTESGETREQUEST']._serialized_start=369
  _globals['_PROJECTATTRIBUTESGETREQUEST']._serialized_end=496
  _globals['_PROJECTATTRIBUTESGETRESPONSE']._serialized_start=498
  _globals['_PROJECTATTRIBUTESGETRESPONSE']._serialized_end=595
  _globals['_PROJECTATTRIBUTESDELETEREQUEST']._serialized_start=598
  _globals['_PROJECTATTRIBUTESDELETEREQUEST']._serialized_end=728
  _globals['_PROJECTATTRIBUTESDELETERESPONSE']._serialized_start=730
  _globals['_PROJECTATTRIBUTESDELETERESPONSE']._serialized_end=763
# @@protoc_insertion_point(module_scope)
