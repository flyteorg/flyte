# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: flyteidl/datacatalog/datacatalog.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from flyteidl.core import literals_pb2 as flyteidl_dot_core_dot_literals__pb2
from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n&flyteidl/datacatalog/datacatalog.proto\x12\x0b\x64\x61tacatalog\x1a\x1c\x66lyteidl/core/literals.proto\x1a\x1egoogle/protobuf/duration.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"F\n\x14\x43reateDatasetRequest\x12.\n\x07\x64\x61taset\x18\x01 \x01(\x0b\x32\x14.datacatalog.DatasetR\x07\x64\x61taset\"\x17\n\x15\x43reateDatasetResponse\"E\n\x11GetDatasetRequest\x12\x30\n\x07\x64\x61taset\x18\x01 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x07\x64\x61taset\"D\n\x12GetDatasetResponse\x12.\n\x07\x64\x61taset\x18\x01 \x01(\x0b\x32\x14.datacatalog.DatasetR\x07\x64\x61taset\"\x96\x01\n\x12GetArtifactRequest\x12\x30\n\x07\x64\x61taset\x18\x01 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x07\x64\x61taset\x12!\n\x0b\x61rtifact_id\x18\x02 \x01(\tH\x00R\nartifactId\x12\x1b\n\x08tag_name\x18\x03 \x01(\tH\x00R\x07tagNameB\x0e\n\x0cquery_handle\"H\n\x13GetArtifactResponse\x12\x31\n\x08\x61rtifact\x18\x01 \x01(\x0b\x32\x15.datacatalog.ArtifactR\x08\x61rtifact\"J\n\x15\x43reateArtifactRequest\x12\x31\n\x08\x61rtifact\x18\x01 \x01(\x0b\x32\x15.datacatalog.ArtifactR\x08\x61rtifact\"\x18\n\x16\x43reateArtifactResponse\"3\n\rAddTagRequest\x12\"\n\x03tag\x18\x01 \x01(\x0b\x32\x10.datacatalog.TagR\x03tag\"\x10\n\x0e\x41\x64\x64TagResponse\"\xbf\x01\n\x14ListArtifactsRequest\x12\x30\n\x07\x64\x61taset\x18\x01 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x07\x64\x61taset\x12\x35\n\x06\x66ilter\x18\x02 \x01(\x0b\x32\x1d.datacatalog.FilterExpressionR\x06\x66ilter\x12>\n\npagination\x18\x03 \x01(\x0b\x32\x1e.datacatalog.PaginationOptionsR\npagination\"k\n\x15ListArtifactsResponse\x12\x33\n\tartifacts\x18\x01 \x03(\x0b\x32\x15.datacatalog.ArtifactR\tartifacts\x12\x1d\n\nnext_token\x18\x02 \x01(\tR\tnextToken\"\x8c\x01\n\x13ListDatasetsRequest\x12\x35\n\x06\x66ilter\x18\x01 \x01(\x0b\x32\x1d.datacatalog.FilterExpressionR\x06\x66ilter\x12>\n\npagination\x18\x02 \x01(\x0b\x32\x1e.datacatalog.PaginationOptionsR\npagination\"g\n\x14ListDatasetsResponse\x12\x30\n\x08\x64\x61tasets\x18\x01 \x03(\x0b\x32\x14.datacatalog.DatasetR\x08\x64\x61tasets\x12\x1d\n\nnext_token\x18\x02 \x01(\tR\tnextToken\"\xfb\x01\n\x15UpdateArtifactRequest\x12\x30\n\x07\x64\x61taset\x18\x01 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x07\x64\x61taset\x12!\n\x0b\x61rtifact_id\x18\x02 \x01(\tH\x00R\nartifactId\x12\x1b\n\x08tag_name\x18\x03 \x01(\tH\x00R\x07tagName\x12-\n\x04\x64\x61ta\x18\x04 \x03(\x0b\x32\x19.datacatalog.ArtifactDataR\x04\x64\x61ta\x12\x31\n\x08metadata\x18\x05 \x01(\x0b\x32\x15.datacatalog.MetadataR\x08metadataB\x0e\n\x0cquery_handle\"9\n\x16UpdateArtifactResponse\x12\x1f\n\x0b\x61rtifact_id\x18\x01 \x01(\tR\nartifactId\"a\n\rReservationID\x12\x35\n\ndataset_id\x18\x01 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\tdatasetId\x12\x19\n\x08tag_name\x18\x02 \x01(\tR\x07tagName\"\xc7\x01\n\x1dGetOrExtendReservationRequest\x12\x41\n\x0ereservation_id\x18\x01 \x01(\x0b\x32\x1a.datacatalog.ReservationIDR\rreservationId\x12\x19\n\x08owner_id\x18\x02 \x01(\tR\x07ownerId\x12H\n\x12heartbeat_interval\x18\x03 \x01(\x0b\x32\x19.google.protobuf.DurationR\x11heartbeatInterval\"\xa3\x02\n\x0bReservation\x12\x41\n\x0ereservation_id\x18\x01 \x01(\x0b\x32\x1a.datacatalog.ReservationIDR\rreservationId\x12\x19\n\x08owner_id\x18\x02 \x01(\tR\x07ownerId\x12H\n\x12heartbeat_interval\x18\x03 \x01(\x0b\x32\x19.google.protobuf.DurationR\x11heartbeatInterval\x12\x39\n\nexpires_at\x18\x04 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\texpiresAt\x12\x31\n\x08metadata\x18\x06 \x01(\x0b\x32\x15.datacatalog.MetadataR\x08metadata\"\\\n\x1eGetOrExtendReservationResponse\x12:\n\x0breservation\x18\x01 \x01(\x0b\x32\x18.datacatalog.ReservationR\x0breservation\"y\n\x19ReleaseReservationRequest\x12\x41\n\x0ereservation_id\x18\x01 \x01(\x0b\x32\x1a.datacatalog.ReservationIDR\rreservationId\x12\x19\n\x08owner_id\x18\x02 \x01(\tR\x07ownerId\"\x1c\n\x1aReleaseReservationResponse\"\x8a\x01\n\x07\x44\x61taset\x12&\n\x02id\x18\x01 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x02id\x12\x31\n\x08metadata\x18\x02 \x01(\x0b\x32\x15.datacatalog.MetadataR\x08metadata\x12$\n\rpartitionKeys\x18\x03 \x03(\tR\rpartitionKeys\"3\n\tPartition\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value\"\x91\x01\n\tDatasetID\x12\x18\n\x07project\x18\x01 \x01(\tR\x07project\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x16\n\x06\x64omain\x18\x03 \x01(\tR\x06\x64omain\x12\x18\n\x07version\x18\x04 \x01(\tR\x07version\x12\x12\n\x04UUID\x18\x05 \x01(\tR\x04UUID\x12\x10\n\x03org\x18\x06 \x01(\tR\x03org\"\xc7\x02\n\x08\x41rtifact\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\x12\x30\n\x07\x64\x61taset\x18\x02 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x07\x64\x61taset\x12-\n\x04\x64\x61ta\x18\x03 \x03(\x0b\x32\x19.datacatalog.ArtifactDataR\x04\x64\x61ta\x12\x31\n\x08metadata\x18\x04 \x01(\x0b\x32\x15.datacatalog.MetadataR\x08metadata\x12\x36\n\npartitions\x18\x05 \x03(\x0b\x32\x16.datacatalog.PartitionR\npartitions\x12$\n\x04tags\x18\x06 \x03(\x0b\x32\x10.datacatalog.TagR\x04tags\x12\x39\n\ncreated_at\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\"P\n\x0c\x41rtifactData\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12,\n\x05value\x18\x02 \x01(\x0b\x32\x16.flyteidl.core.LiteralR\x05value\"l\n\x03Tag\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1f\n\x0b\x61rtifact_id\x18\x02 \x01(\tR\nartifactId\x12\x30\n\x07\x64\x61taset\x18\x03 \x01(\x0b\x32\x16.datacatalog.DatasetIDR\x07\x64\x61taset\"\x81\x01\n\x08Metadata\x12:\n\x07key_map\x18\x01 \x03(\x0b\x32!.datacatalog.Metadata.KeyMapEntryR\x06keyMap\x1a\x39\n\x0bKeyMapEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value:\x02\x38\x01\"O\n\x10\x46ilterExpression\x12;\n\x07\x66ilters\x18\x01 \x03(\x0b\x32!.datacatalog.SinglePropertyFilterR\x07\x66ilters\"\xce\x03\n\x14SinglePropertyFilter\x12?\n\ntag_filter\x18\x01 \x01(\x0b\x32\x1e.datacatalog.TagPropertyFilterH\x00R\ttagFilter\x12Q\n\x10partition_filter\x18\x02 \x01(\x0b\x32$.datacatalog.PartitionPropertyFilterH\x00R\x0fpartitionFilter\x12N\n\x0f\x61rtifact_filter\x18\x03 \x01(\x0b\x32#.datacatalog.ArtifactPropertyFilterH\x00R\x0e\x61rtifactFilter\x12K\n\x0e\x64\x61taset_filter\x18\x04 \x01(\x0b\x32\".datacatalog.DatasetPropertyFilterH\x00R\rdatasetFilter\x12P\n\x08operator\x18\n \x01(\x0e\x32\x34.datacatalog.SinglePropertyFilter.ComparisonOperatorR\x08operator\" \n\x12\x43omparisonOperator\x12\n\n\x06\x45QUALS\x10\x00\x42\x11\n\x0fproperty_filter\"G\n\x16\x41rtifactPropertyFilter\x12!\n\x0b\x61rtifact_id\x18\x01 \x01(\tH\x00R\nartifactIdB\n\n\x08property\"<\n\x11TagPropertyFilter\x12\x1b\n\x08tag_name\x18\x01 \x01(\tH\x00R\x07tagNameB\n\n\x08property\"[\n\x17PartitionPropertyFilter\x12\x34\n\x07key_val\x18\x01 \x01(\x0b\x32\x19.datacatalog.KeyValuePairH\x00R\x06keyValB\n\n\x08property\"6\n\x0cKeyValuePair\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value\"\x9f\x01\n\x15\x44\x61tasetPropertyFilter\x12\x1a\n\x07project\x18\x01 \x01(\tH\x00R\x07project\x12\x14\n\x04name\x18\x02 \x01(\tH\x00R\x04name\x12\x18\n\x06\x64omain\x18\x03 \x01(\tH\x00R\x06\x64omain\x12\x1a\n\x07version\x18\x04 \x01(\tH\x00R\x07version\x12\x12\n\x03org\x18\x05 \x01(\tH\x00R\x03orgB\n\n\x08property\"\x93\x02\n\x11PaginationOptions\x12\x14\n\x05limit\x18\x01 \x01(\rR\x05limit\x12\x14\n\x05token\x18\x02 \x01(\tR\x05token\x12@\n\x07sortKey\x18\x03 \x01(\x0e\x32&.datacatalog.PaginationOptions.SortKeyR\x07sortKey\x12\x46\n\tsortOrder\x18\x04 \x01(\x0e\x32(.datacatalog.PaginationOptions.SortOrderR\tsortOrder\"*\n\tSortOrder\x12\x0e\n\nDESCENDING\x10\x00\x12\r\n\tASCENDING\x10\x01\"\x1c\n\x07SortKey\x12\x11\n\rCREATION_TIME\x10\x00\x32\xa0\t\n\x0b\x44\x61taCatalog\x12V\n\rCreateDataset\x12!.datacatalog.CreateDatasetRequest\x1a\".datacatalog.CreateDatasetResponse\x12M\n\nGetDataset\x12\x1e.datacatalog.GetDatasetRequest\x1a\x1f.datacatalog.GetDatasetResponse\x12Y\n\x0e\x43reateArtifact\x12\".datacatalog.CreateArtifactRequest\x1a#.datacatalog.CreateArtifactResponse\x12P\n\x0bGetArtifact\x12\x1f.datacatalog.GetArtifactRequest\x1a .datacatalog.GetArtifactResponse\x12_\n\x14\x43reateFutureArtifact\x12\".datacatalog.CreateArtifactRequest\x1a#.datacatalog.CreateArtifactResponse\x12V\n\x11GetFutureArtifact\x12\x1f.datacatalog.GetArtifactRequest\x1a .datacatalog.GetArtifactResponse\x12_\n\x14UpdateFutureArtifact\x12\".datacatalog.UpdateArtifactRequest\x1a#.datacatalog.UpdateArtifactResponse\x12\x41\n\x06\x41\x64\x64Tag\x12\x1a.datacatalog.AddTagRequest\x1a\x1b.datacatalog.AddTagResponse\x12V\n\rListArtifacts\x12!.datacatalog.ListArtifactsRequest\x1a\".datacatalog.ListArtifactsResponse\x12S\n\x0cListDatasets\x12 .datacatalog.ListDatasetsRequest\x1a!.datacatalog.ListDatasetsResponse\x12Y\n\x0eUpdateArtifact\x12\".datacatalog.UpdateArtifactRequest\x1a#.datacatalog.UpdateArtifactResponse\x12q\n\x16GetOrExtendReservation\x12*.datacatalog.GetOrExtendReservationRequest\x1a+.datacatalog.GetOrExtendReservationResponse\x12\x65\n\x12ReleaseReservation\x12&.datacatalog.ReleaseReservationRequest\x1a\'.datacatalog.ReleaseReservationResponseB\xb2\x01\n\x0f\x63om.datacatalogB\x10\x44\x61tacatalogProtoP\x01ZAgithub.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog\xa2\x02\x03\x44XX\xaa\x02\x0b\x44\x61tacatalog\xca\x02\x0b\x44\x61tacatalog\xe2\x02\x17\x44\x61tacatalog\\GPBMetadata\xea\x02\x0b\x44\x61tacatalogb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'flyteidl.datacatalog.datacatalog_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\017com.datacatalogB\020DatacatalogProtoP\001ZAgithub.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog\242\002\003DXX\252\002\013Datacatalog\312\002\013Datacatalog\342\002\027Datacatalog\\GPBMetadata\352\002\013Datacatalog'
  _METADATA_KEYMAPENTRY._options = None
  _METADATA_KEYMAPENTRY._serialized_options = b'8\001'
  _globals['_CREATEDATASETREQUEST']._serialized_start=150
  _globals['_CREATEDATASETREQUEST']._serialized_end=220
  _globals['_CREATEDATASETRESPONSE']._serialized_start=222
  _globals['_CREATEDATASETRESPONSE']._serialized_end=245
  _globals['_GETDATASETREQUEST']._serialized_start=247
  _globals['_GETDATASETREQUEST']._serialized_end=316
  _globals['_GETDATASETRESPONSE']._serialized_start=318
  _globals['_GETDATASETRESPONSE']._serialized_end=386
  _globals['_GETARTIFACTREQUEST']._serialized_start=389
  _globals['_GETARTIFACTREQUEST']._serialized_end=539
  _globals['_GETARTIFACTRESPONSE']._serialized_start=541
  _globals['_GETARTIFACTRESPONSE']._serialized_end=613
  _globals['_CREATEARTIFACTREQUEST']._serialized_start=615
  _globals['_CREATEARTIFACTREQUEST']._serialized_end=689
  _globals['_CREATEARTIFACTRESPONSE']._serialized_start=691
  _globals['_CREATEARTIFACTRESPONSE']._serialized_end=715
  _globals['_ADDTAGREQUEST']._serialized_start=717
  _globals['_ADDTAGREQUEST']._serialized_end=768
  _globals['_ADDTAGRESPONSE']._serialized_start=770
  _globals['_ADDTAGRESPONSE']._serialized_end=786
  _globals['_LISTARTIFACTSREQUEST']._serialized_start=789
  _globals['_LISTARTIFACTSREQUEST']._serialized_end=980
  _globals['_LISTARTIFACTSRESPONSE']._serialized_start=982
  _globals['_LISTARTIFACTSRESPONSE']._serialized_end=1089
  _globals['_LISTDATASETSREQUEST']._serialized_start=1092
  _globals['_LISTDATASETSREQUEST']._serialized_end=1232
  _globals['_LISTDATASETSRESPONSE']._serialized_start=1234
  _globals['_LISTDATASETSRESPONSE']._serialized_end=1337
  _globals['_UPDATEARTIFACTREQUEST']._serialized_start=1340
  _globals['_UPDATEARTIFACTREQUEST']._serialized_end=1591
  _globals['_UPDATEARTIFACTRESPONSE']._serialized_start=1593
  _globals['_UPDATEARTIFACTRESPONSE']._serialized_end=1650
  _globals['_RESERVATIONID']._serialized_start=1652
  _globals['_RESERVATIONID']._serialized_end=1749
  _globals['_GETOREXTENDRESERVATIONREQUEST']._serialized_start=1752
  _globals['_GETOREXTENDRESERVATIONREQUEST']._serialized_end=1951
  _globals['_RESERVATION']._serialized_start=1954
  _globals['_RESERVATION']._serialized_end=2245
  _globals['_GETOREXTENDRESERVATIONRESPONSE']._serialized_start=2247
  _globals['_GETOREXTENDRESERVATIONRESPONSE']._serialized_end=2339
  _globals['_RELEASERESERVATIONREQUEST']._serialized_start=2341
  _globals['_RELEASERESERVATIONREQUEST']._serialized_end=2462
  _globals['_RELEASERESERVATIONRESPONSE']._serialized_start=2464
  _globals['_RELEASERESERVATIONRESPONSE']._serialized_end=2492
  _globals['_DATASET']._serialized_start=2495
  _globals['_DATASET']._serialized_end=2633
  _globals['_PARTITION']._serialized_start=2635
  _globals['_PARTITION']._serialized_end=2686
  _globals['_DATASETID']._serialized_start=2689
  _globals['_DATASETID']._serialized_end=2834
  _globals['_ARTIFACT']._serialized_start=2837
  _globals['_ARTIFACT']._serialized_end=3164
  _globals['_ARTIFACTDATA']._serialized_start=3166
  _globals['_ARTIFACTDATA']._serialized_end=3246
  _globals['_TAG']._serialized_start=3248
  _globals['_TAG']._serialized_end=3356
  _globals['_METADATA']._serialized_start=3359
  _globals['_METADATA']._serialized_end=3488
  _globals['_METADATA_KEYMAPENTRY']._serialized_start=3431
  _globals['_METADATA_KEYMAPENTRY']._serialized_end=3488
  _globals['_FILTEREXPRESSION']._serialized_start=3490
  _globals['_FILTEREXPRESSION']._serialized_end=3569
  _globals['_SINGLEPROPERTYFILTER']._serialized_start=3572
  _globals['_SINGLEPROPERTYFILTER']._serialized_end=4034
  _globals['_SINGLEPROPERTYFILTER_COMPARISONOPERATOR']._serialized_start=3983
  _globals['_SINGLEPROPERTYFILTER_COMPARISONOPERATOR']._serialized_end=4015
  _globals['_ARTIFACTPROPERTYFILTER']._serialized_start=4036
  _globals['_ARTIFACTPROPERTYFILTER']._serialized_end=4107
  _globals['_TAGPROPERTYFILTER']._serialized_start=4109
  _globals['_TAGPROPERTYFILTER']._serialized_end=4169
  _globals['_PARTITIONPROPERTYFILTER']._serialized_start=4171
  _globals['_PARTITIONPROPERTYFILTER']._serialized_end=4262
  _globals['_KEYVALUEPAIR']._serialized_start=4264
  _globals['_KEYVALUEPAIR']._serialized_end=4318
  _globals['_DATASETPROPERTYFILTER']._serialized_start=4321
  _globals['_DATASETPROPERTYFILTER']._serialized_end=4480
  _globals['_PAGINATIONOPTIONS']._serialized_start=4483
  _globals['_PAGINATIONOPTIONS']._serialized_end=4758
  _globals['_PAGINATIONOPTIONS_SORTORDER']._serialized_start=4686
  _globals['_PAGINATIONOPTIONS_SORTORDER']._serialized_end=4728
  _globals['_PAGINATIONOPTIONS_SORTKEY']._serialized_start=4730
  _globals['_PAGINATIONOPTIONS_SORTKEY']._serialized_end=4758
  _globals['_DATACATALOG']._serialized_start=4761
  _globals['_DATACATALOG']._serialized_end=5945
# @@protoc_insertion_point(module_scope)
