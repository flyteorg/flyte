// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: flyteidl/service/agent.proto

package service

import (
	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

var File_flyteidl_service_agent_proto protoreflect.FileDescriptor

var file_flyteidl_service_agent_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xcc, 0x03, 0x0a, 0x11, 0x41,
	0x73, 0x79, 0x6e, 0x63, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x55, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x21,
	0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x22, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4c, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x54, 0x61,
	0x73, 0x6b, 0x12, 0x1e, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x55, 0x0a, 0x0a, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54,
	0x61, 0x73, 0x6b, 0x12, 0x21, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64,
	0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x61, 0x0a, 0x0e,
	0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x25,
	0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c,
	0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x58, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x73, 0x12, 0x22,
	0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x23, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x32, 0xf0, 0x01, 0x0a, 0x14, 0x41, 0x67,
	0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x6b, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x12, 0x1f,
	0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e,
	0x47, 0x65, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x20, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x47, 0x65, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x12, 0x14, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x12,
	0x6b, 0x0a, 0x0a, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x21, 0x2e,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x22, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x16, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x10, 0x12, 0x0e, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x73, 0x42, 0xc2, 0x01, 0x0a,
	0x14, 0x63, 0x6f, 0x6d, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x0a, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x6f, 0x72, 0x67, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x2f, 0x66,
	0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x62, 0x2d, 0x67,
	0x6f, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0xa2, 0x02, 0x03, 0x46, 0x53, 0x58, 0xaa, 0x02, 0x10, 0x46, 0x6c, 0x79, 0x74, 0x65,
	0x69, 0x64, 0x6c, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xca, 0x02, 0x10, 0x46, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xe2, 0x02,
	0x1c, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x11,
	0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x3a, 0x3a, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_flyteidl_service_agent_proto_goTypes = []interface{}{
	(*admin.CreateTaskRequest)(nil),      // 0: flyteidl.admin.CreateTaskRequest
	(*admin.GetTaskRequest)(nil),         // 1: flyteidl.admin.GetTaskRequest
	(*admin.DeleteTaskRequest)(nil),      // 2: flyteidl.admin.DeleteTaskRequest
	(*admin.GetTaskMetricsRequest)(nil),  // 3: flyteidl.admin.GetTaskMetricsRequest
	(*admin.GetTaskLogsRequest)(nil),     // 4: flyteidl.admin.GetTaskLogsRequest
	(*admin.GetAgentRequest)(nil),        // 5: flyteidl.admin.GetAgentRequest
	(*admin.ListAgentsRequest)(nil),      // 6: flyteidl.admin.ListAgentsRequest
	(*admin.CreateTaskResponse)(nil),     // 7: flyteidl.admin.CreateTaskResponse
	(*admin.GetTaskResponse)(nil),        // 8: flyteidl.admin.GetTaskResponse
	(*admin.DeleteTaskResponse)(nil),     // 9: flyteidl.admin.DeleteTaskResponse
	(*admin.GetTaskMetricsResponse)(nil), // 10: flyteidl.admin.GetTaskMetricsResponse
	(*admin.GetTaskLogsResponse)(nil),    // 11: flyteidl.admin.GetTaskLogsResponse
	(*admin.GetAgentResponse)(nil),       // 12: flyteidl.admin.GetAgentResponse
	(*admin.ListAgentsResponse)(nil),     // 13: flyteidl.admin.ListAgentsResponse
}
var file_flyteidl_service_agent_proto_depIdxs = []int32{
	0,  // 0: flyteidl.service.AsyncAgentService.CreateTask:input_type -> flyteidl.admin.CreateTaskRequest
	1,  // 1: flyteidl.service.AsyncAgentService.GetTask:input_type -> flyteidl.admin.GetTaskRequest
	2,  // 2: flyteidl.service.AsyncAgentService.DeleteTask:input_type -> flyteidl.admin.DeleteTaskRequest
	3,  // 3: flyteidl.service.AsyncAgentService.GetTaskMetrics:input_type -> flyteidl.admin.GetTaskMetricsRequest
	4,  // 4: flyteidl.service.AsyncAgentService.GetTaskLogs:input_type -> flyteidl.admin.GetTaskLogsRequest
	5,  // 5: flyteidl.service.AgentMetadataService.GetAgent:input_type -> flyteidl.admin.GetAgentRequest
	6,  // 6: flyteidl.service.AgentMetadataService.ListAgents:input_type -> flyteidl.admin.ListAgentsRequest
	7,  // 7: flyteidl.service.AsyncAgentService.CreateTask:output_type -> flyteidl.admin.CreateTaskResponse
	8,  // 8: flyteidl.service.AsyncAgentService.GetTask:output_type -> flyteidl.admin.GetTaskResponse
	9,  // 9: flyteidl.service.AsyncAgentService.DeleteTask:output_type -> flyteidl.admin.DeleteTaskResponse
	10, // 10: flyteidl.service.AsyncAgentService.GetTaskMetrics:output_type -> flyteidl.admin.GetTaskMetricsResponse
	11, // 11: flyteidl.service.AsyncAgentService.GetTaskLogs:output_type -> flyteidl.admin.GetTaskLogsResponse
	12, // 12: flyteidl.service.AgentMetadataService.GetAgent:output_type -> flyteidl.admin.GetAgentResponse
	13, // 13: flyteidl.service.AgentMetadataService.ListAgents:output_type -> flyteidl.admin.ListAgentsResponse
	7,  // [7:14] is the sub-list for method output_type
	0,  // [0:7] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_flyteidl_service_agent_proto_init() }
func file_flyteidl_service_agent_proto_init() {
	if File_flyteidl_service_agent_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_flyteidl_service_agent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_flyteidl_service_agent_proto_goTypes,
		DependencyIndexes: file_flyteidl_service_agent_proto_depIdxs,
	}.Build()
	File_flyteidl_service_agent_proto = out.File
	file_flyteidl_service_agent_proto_rawDesc = nil
	file_flyteidl_service_agent_proto_goTypes = nil
	file_flyteidl_service_agent_proto_depIdxs = nil
}
