// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: flyteidl/service/watch.proto

package service

import (
	watch "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_flyteidl_service_watch_proto protoreflect.FileDescriptor

var file_flyteidl_service_watch_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2f, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x1a, 0x1f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x77, 0x61, 0x74, 0x63, 0x68,
	0x2f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x32, 0x9b, 0x01, 0x0a, 0x0c, 0x57, 0x61, 0x74, 0x63, 0x68, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x8a, 0x01, 0x0a, 0x1b, 0x57, 0x61, 0x74, 0x63, 0x68, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x73, 0x12, 0x32, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x77, 0x61,
	0x74, 0x63, 0x68, 0x2e, 0x57, 0x61, 0x74, 0x63, 0x68, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69,
	0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x33, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64,
	0x6c, 0x2e, 0x77, 0x61, 0x74, 0x63, 0x68, 0x2e, 0x57, 0x61, 0x74, 0x63, 0x68, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42,
	0xc2, 0x01, 0x0a, 0x14, 0x63, 0x6f, 0x6d, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c,
	0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x0a, 0x57, 0x61, 0x74, 0x63, 0x68, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x6f, 0x72, 0x67, 0x2f, 0x66, 0x6c, 0x79, 0x74,
	0x65, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x62, 0x2d, 0x67, 0x6f, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0xa2, 0x02, 0x03, 0x46, 0x53, 0x58, 0xaa, 0x02, 0x10, 0x46, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xca, 0x02,
	0x10, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0xe2, 0x02, 0x1c, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x11, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x3a, 0x3a, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_flyteidl_service_watch_proto_goTypes = []interface{}{
	(*watch.WatchExecutionStatusUpdatesRequest)(nil),  // 0: flyteidl.watch.WatchExecutionStatusUpdatesRequest
	(*watch.WatchExecutionStatusUpdatesResponse)(nil), // 1: flyteidl.watch.WatchExecutionStatusUpdatesResponse
}
var file_flyteidl_service_watch_proto_depIdxs = []int32{
	0, // 0: flyteidl.service.WatchService.WatchExecutionStatusUpdates:input_type -> flyteidl.watch.WatchExecutionStatusUpdatesRequest
	1, // 1: flyteidl.service.WatchService.WatchExecutionStatusUpdates:output_type -> flyteidl.watch.WatchExecutionStatusUpdatesResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_flyteidl_service_watch_proto_init() }
func file_flyteidl_service_watch_proto_init() {
	if File_flyteidl_service_watch_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_flyteidl_service_watch_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_flyteidl_service_watch_proto_goTypes,
		DependencyIndexes: file_flyteidl_service_watch_proto_depIdxs,
	}.Build()
	File_flyteidl_service_watch_proto = out.File
	file_flyteidl_service_watch_proto_rawDesc = nil
	file_flyteidl_service_watch_proto_goTypes = nil
	file_flyteidl_service_watch_proto_depIdxs = nil
}
