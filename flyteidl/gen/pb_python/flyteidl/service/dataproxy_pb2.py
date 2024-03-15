# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/service/dataproxy.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.api import annotations_pb2 as google_dot_api_dot_annotations__pb2
from protoc_gen_openapiv2.options import annotations_pb2 as protoc__gen__openapiv2_dot_options_dot_annotations__pb2
from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from flyteidl.core import identifier_pb2 as flyteidl_dot_core_dot_identifier__pb2
from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n flyteidl/service/dataproxy.proto\x12\x10\x66lyteidl.service\x1a\x1cgoogle/api/annotations.proto\x1a.protoc-gen-openapiv2/options/annotations.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1e\x66lyteidl/core/identifier.proto\x1a\x1c\x66lyteidl/core/literals.proto\"\xaa\x02\n\x1c\x43reateUploadLocationResponse\x12\x1d\n\nsigned_url\x18\x01 \x01(\tR\tsignedUrl\x12\x1d\n\nnative_url\x18\x02 \x01(\tR\tnativeUrl\x12\x39\n\nexpires_at\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\texpiresAt\x12U\n\x07headers\x18\x04 \x03(\x0b\x32;.flyteidl.service.CreateUploadLocationResponse.HeadersEntryR\x07headers\x1a:\n\x0cHeadersEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"\xb6\x02\n\x1b\x43reateUploadLocationRequest\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x16\n\x06\x64omain\x18\x02 \x01(\tR\x06\x64omain\x12\x1a\n\x08\x66ilename\x18\x03 \x01(\tR\x08\x66ilename\x12\x38\n\nexpires_in\x18\x04 \x01(\x0b\x32\x19.google.protobuf.DurationR\texpiresIn\x12\x1f\n\x0b\x63ontent_md5\x18\x05 \x01(\x0cR\ncontentMd5\x12#\n\rfilename_root\x18\x06 \x01(\tR\x0c\x66ilenameRoot\x12\x37\n\x18\x61\x64\x64_content_md5_metadata\x18\x07 \x01(\x08R\x15\x61\x64\x64\x43ontentMd5Metadata\x12\x10\n\x03org\x18\x08 \x01(\tR\x03org\"|\n\x1d\x43reateDownloadLocationRequest\x12\x1d\n\nnative_url\x18\x01 \x01(\tR\tnativeUrl\x12\x38\n\nexpires_in\x18\x02 \x01(\x0b\x32\x19.google.protobuf.DurationR\texpiresIn:\x02\x18\x01\"~\n\x1e\x43reateDownloadLocationResponse\x12\x1d\n\nsigned_url\x18\x01 \x01(\tR\tsignedUrl\x12\x39\n\nexpires_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\texpiresAt:\x02\x18\x01\"\xfa\x01\n\x19\x43reateDownloadLinkRequest\x12\x43\n\rartifact_type\x18\x01 \x01(\x0e\x32\x1e.flyteidl.service.ArtifactTypeR\x0c\x61rtifactType\x12\x38\n\nexpires_in\x18\x02 \x01(\x0b\x32\x19.google.protobuf.DurationR\texpiresIn\x12T\n\x11node_execution_id\x18\x03 \x01(\x0b\x32&.flyteidl.core.NodeExecutionIdentifierH\x00R\x0fnodeExecutionIdB\x08\n\x06source\"\xc7\x01\n\x1a\x43reateDownloadLinkResponse\x12!\n\nsigned_url\x18\x01 \x03(\tB\x02\x18\x01R\tsignedUrl\x12=\n\nexpires_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\x02\x18\x01R\texpiresAt\x12G\n\x0fpre_signed_urls\x18\x03 \x01(\x0b\x32\x1f.flyteidl.service.PreSignedURLsR\rpreSignedUrls\"i\n\rPreSignedURLs\x12\x1d\n\nsigned_url\x18\x01 \x03(\tR\tsignedUrl\x12\x39\n\nexpires_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\texpiresAt\"-\n\x0eGetDataRequest\x12\x1b\n\tflyte_url\x18\x01 \x01(\tR\x08\x66lyteUrl\"\xd6\x01\n\x0fGetDataResponse\x12<\n\x0bliteral_map\x18\x01 \x01(\x0b\x32\x19.flyteidl.core.LiteralMapH\x00R\nliteralMap\x12I\n\x0fpre_signed_urls\x18\x02 \x01(\x0b\x32\x1f.flyteidl.service.PreSignedURLsH\x00R\rpreSignedUrls\x12\x32\n\x07literal\x18\x03 \x01(\x0b\x32\x16.flyteidl.core.LiteralH\x00R\x07literalB\x06\n\x04\x64\x61ta*C\n\x0c\x41rtifactType\x12\x1b\n\x17\x41RTIFACT_TYPE_UNDEFINED\x10\x00\x12\x16\n\x12\x41RTIFACT_TYPE_DECK\x10\x01\x32\xb4\x07\n\x10\x44\x61taProxyService\x12\xa0\x02\n\x14\x43reateUploadLocation\x12-.flyteidl.service.CreateUploadLocationRequest\x1a..flyteidl.service.CreateUploadLocationResponse\"\xa8\x01\x92\x41M\x1aKCreates a write-only http location that is accessible for tasks at runtime.\x82\xd3\xe4\x93\x02R:\x01*Z-:\x01*\"(/api/v1/org/dataproxy/artifact_urn/{org}\"\x1e/api/v1/dataproxy/artifact_urn\x12\xa9\x02\n\x16\x43reateDownloadLocation\x12/.flyteidl.service.CreateDownloadLocationRequest\x1a\x30.flyteidl.service.CreateDownloadLocationResponse\"\xab\x01\x88\x02\x01\x92\x41\x7f\x1a}Deprecated: Please use CreateDownloadLink instead. Creates a read-only http location that is accessible for tasks at runtime.\x82\xd3\xe4\x93\x02 \x12\x1e/api/v1/dataproxy/artifact_urn\x12\xea\x01\n\x12\x43reateDownloadLink\x12+.flyteidl.service.CreateDownloadLinkRequest\x1a,.flyteidl.service.CreateDownloadLinkResponse\"y\x92\x41L\x1aJCreates a read-only http location that is accessible for tasks at runtime.\x82\xd3\xe4\x93\x02$:\x01*\"\x1f/api/v1/dataproxy/artifact_link\x12\x64\n\x07GetData\x12 .flyteidl.service.GetDataRequest\x1a!.flyteidl.service.GetDataResponse\"\x14\x82\xd3\xe4\x93\x02\x0e\x12\x0c/api/v1/dataB\xc6\x01\n\x14\x63om.flyteidl.serviceB\x0e\x44\x61taproxyProtoP\x01Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service\xa2\x02\x03\x46SX\xaa\x02\x10\x46lyteidl.Service\xca\x02\x10\x46lyteidl\\Service\xe2\x02\x1c\x46lyteidl\\Service\\GPBMetadata\xea\x02\x11\x46lyteidl::Serviceb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.service.dataproxy_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\024com.flyteidl.serviceB\016DataproxyProtoP\001Z=github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service\242\002\003FSX\252\002\020Flyteidl.Service\312\002\020Flyteidl\\Service\342\002\034Flyteidl\\Service\\GPBMetadata\352\002\021Flyteidl::Service'
  _CREATEUPLOADLOCATIONRESPONSE_HEADERSENTRY._options = None
  _CREATEUPLOADLOCATIONRESPONSE_HEADERSENTRY._serialized_options = b'8\001'
  _CREATEDOWNLOADLOCATIONREQUEST._options = None
  _CREATEDOWNLOADLOCATIONREQUEST._serialized_options = b'\030\001'
  _CREATEDOWNLOADLOCATIONRESPONSE._options = None
  _CREATEDOWNLOADLOCATIONRESPONSE._serialized_options = b'\030\001'
  _CREATEDOWNLOADLINKRESPONSE.fields_by_name['signed_url']._options = None
  _CREATEDOWNLOADLINKRESPONSE.fields_by_name['signed_url']._serialized_options = b'\030\001'
  _CREATEDOWNLOADLINKRESPONSE.fields_by_name['expires_at']._options = None
  _CREATEDOWNLOADLINKRESPONSE.fields_by_name['expires_at']._serialized_options = b'\030\001'
  _DATAPROXYSERVICE.methods_by_name['CreateUploadLocation']._options = None
  _DATAPROXYSERVICE.methods_by_name['CreateUploadLocation']._serialized_options = b'\222AM\032KCreates a write-only http location that is accessible for tasks at runtime.\202\323\344\223\002R:\001*Z-:\001*\"(/api/v1/org/dataproxy/artifact_urn/{org}\"\036/api/v1/dataproxy/artifact_urn'
  _DATAPROXYSERVICE.methods_by_name['CreateDownloadLocation']._options = None
  _DATAPROXYSERVICE.methods_by_name['CreateDownloadLocation']._serialized_options = b'\210\002\001\222A\177\032}Deprecated: Please use CreateDownloadLink instead. Creates a read-only http location that is accessible for tasks at runtime.\202\323\344\223\002 \022\036/api/v1/dataproxy/artifact_urn'
  _DATAPROXYSERVICE.methods_by_name['CreateDownloadLink']._options = None
  _DATAPROXYSERVICE.methods_by_name['CreateDownloadLink']._serialized_options = b'\222AL\032JCreates a read-only http location that is accessible for tasks at runtime.\202\323\344\223\002$:\001*\"\037/api/v1/dataproxy/artifact_link'
  _DATAPROXYSERVICE.methods_by_name['GetData']._options = None
  _DATAPROXYSERVICE.methods_by_name['GetData']._serialized_options = b'\202\323\344\223\002\016\022\014/api/v1/data'
  _globals['_ARTIFACTTYPE']._serialized_start=1953
  _globals['_ARTIFACTTYPE']._serialized_end=2020
  _globals['_CREATEUPLOADLOCATIONRESPONSE']._serialized_start=260
  _globals['_CREATEUPLOADLOCATIONRESPONSE']._serialized_end=558
  _globals['_CREATEUPLOADLOCATIONRESPONSE_HEADERSENTRY']._serialized_start=500
  _globals['_CREATEUPLOADLOCATIONRESPONSE_HEADERSENTRY']._serialized_end=558
  _globals['_CREATEUPLOADLOCATIONREQUEST']._serialized_start=561
  _globals['_CREATEUPLOADLOCATIONREQUEST']._serialized_end=871
  _globals['_CREATEDOWNLOADLOCATIONREQUEST']._serialized_start=873
  _globals['_CREATEDOWNLOADLOCATIONREQUEST']._serialized_end=997
  _globals['_CREATEDOWNLOADLOCATIONRESPONSE']._serialized_start=999
  _globals['_CREATEDOWNLOADLOCATIONRESPONSE']._serialized_end=1125
  _globals['_CREATEDOWNLOADLINKREQUEST']._serialized_start=1128
  _globals['_CREATEDOWNLOADLINKREQUEST']._serialized_end=1378
  _globals['_CREATEDOWNLOADLINKRESPONSE']._serialized_start=1381
  _globals['_CREATEDOWNLOADLINKRESPONSE']._serialized_end=1580
  _globals['_PRESIGNEDURLS']._serialized_start=1582
  _globals['_PRESIGNEDURLS']._serialized_end=1687
  _globals['_GETDATAREQUEST']._serialized_start=1689
  _globals['_GETDATAREQUEST']._serialized_end=1734
  _globals['_GETDATARESPONSE']._serialized_start=1737
  _globals['_GETDATARESPONSE']._serialized_end=1951
  _globals['_DATAPROXYSERVICE']._serialized_start=2023
  _globals['_DATAPROXYSERVICE']._serialized_end=2971
# @@protoc_insertion_point(module_scope)
