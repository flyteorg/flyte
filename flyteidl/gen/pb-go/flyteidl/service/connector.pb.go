// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: flyteidl/service/connector.proto

package service

import (
	connector "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/connector"
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

var File_flyteidl_service_connector_proto protoreflect.FileDescriptor

var file_flyteidl_service_connector_proto_rawDesc = []byte{
	0x0a, 0x20, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x10, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x22, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xb1, 0x01, 0x0a, 0x14, 0x53, 0x79, 0x6e, 0x63, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x98, 0x01, 0x0a, 0x0f, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x53,
	0x79, 0x6e, 0x63, 0x12, 0x2a, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65,
	0x54, 0x61, 0x73, 0x6b, 0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2b, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b,
	0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x28, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x22, 0x3a, 0x01, 0x2a, 0x22, 0x1d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31,
	0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2f,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x28, 0x01, 0x30, 0x01, 0x32, 0xa3, 0x07, 0x0a, 0x15, 0x41,
	0x73, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x7e, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x61,
	0x73, 0x6b, 0x12, 0x25, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x66, 0x6c, 0x79, 0x74,
	0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x21, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1b, 0x3a, 0x01, 0x2a, 0x22, 0x16, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f,
	0x74, 0x61, 0x73, 0x6b, 0x12, 0xaf, 0x01, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b,
	0x12, 0x22, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73,
	0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x5b, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x55, 0x12, 0x53, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2f, 0x7b, 0x74, 0x61, 0x73, 0x6b, 0x5f,
	0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x2f, 0x7b,
	0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x7d, 0x2f, 0x7b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x5f, 0x6d, 0x65, 0x74, 0x61, 0x7d, 0x12, 0xc3, 0x01, 0x0a, 0x0a, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x25, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x66,
	0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x66, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x60, 0x2a, 0x5e, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f,
	0x74, 0x61, 0x73, 0x6b, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f,
	0x7b, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x6e,
	0x61, 0x6d, 0x65, 0x7d, 0x2f, 0x7b, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67,
	0x6f, 0x72, 0x79, 0x2e, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x7d, 0x2f, 0x7b, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x7d, 0x12, 0xcc, 0x01, 0x0a,
	0x0e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12,
	0x29, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x66, 0x6c, 0x79,
	0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e,
	0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x63, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x5d, 0x12, 0x5b,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2f, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2f, 0x7b,
	0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x6e, 0x61,
	0x6d, 0x65, 0x7d, 0x2f, 0x7b, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f,
	0x72, 0x79, 0x2e, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x7d, 0x2f, 0x7b, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x7d, 0x12, 0xc2, 0x01, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x73, 0x12, 0x26, 0x2e, 0x66, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b,
	0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x60, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x5a, 0x12, 0x58, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2f, 0x6c, 0x6f, 0x67,
	0x73, 0x2f, 0x7b, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x2e, 0x6e, 0x61, 0x6d, 0x65, 0x7d, 0x2f, 0x7b, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x63, 0x61, 0x74,
	0x65, 0x67, 0x6f, 0x72, 0x79, 0x2e, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x7d, 0x2f, 0x7b,
	0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x7d, 0x30, 0x01,
	0x32, 0xa6, 0x02, 0x0a, 0x18, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x83, 0x01,
	0x0a, 0x0c, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x27,
	0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69,
	0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x47, 0x65, 0x74,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x20, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1a, 0x12, 0x18, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2f, 0x7b, 0x6e, 0x61,
	0x6d, 0x65, 0x7d, 0x12, 0x83, 0x01, 0x0a, 0x0e, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x12, 0x29, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64,
	0x6c, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x2a, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1a, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x14, 0x12, 0x12, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x42, 0xc6, 0x01, 0x0a, 0x14, 0x63, 0x6f,
	0x6d, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x42, 0x0e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x6f, 0x72, 0x67, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x2f,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x62, 0x2d,
	0x67, 0x6f, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0xa2, 0x02, 0x03, 0x46, 0x53, 0x58, 0xaa, 0x02, 0x10, 0x46, 0x6c, 0x79, 0x74,
	0x65, 0x69, 0x64, 0x6c, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xca, 0x02, 0x10, 0x46,
	0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0xe2,
	0x02, 0x1c, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x11, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x3a, 0x3a, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_flyteidl_service_connector_proto_goTypes = []interface{}{
	(*connector.ExecuteTaskSyncRequest)(nil),  // 0: flyteidl.connector.ExecuteTaskSyncRequest
	(*connector.CreateTaskRequest)(nil),       // 1: flyteidl.connector.CreateTaskRequest
	(*connector.GetTaskRequest)(nil),          // 2: flyteidl.connector.GetTaskRequest
	(*connector.DeleteTaskRequest)(nil),       // 3: flyteidl.connector.DeleteTaskRequest
	(*connector.GetTaskMetricsRequest)(nil),   // 4: flyteidl.connector.GetTaskMetricsRequest
	(*connector.GetTaskLogsRequest)(nil),      // 5: flyteidl.connector.GetTaskLogsRequest
	(*connector.GetConnectorRequest)(nil),     // 6: flyteidl.connector.GetConnectorRequest
	(*connector.ListConnectorsRequest)(nil),   // 7: flyteidl.connector.ListConnectorsRequest
	(*connector.ExecuteTaskSyncResponse)(nil), // 8: flyteidl.connector.ExecuteTaskSyncResponse
	(*connector.CreateTaskResponse)(nil),      // 9: flyteidl.connector.CreateTaskResponse
	(*connector.GetTaskResponse)(nil),         // 10: flyteidl.connector.GetTaskResponse
	(*connector.DeleteTaskResponse)(nil),      // 11: flyteidl.connector.DeleteTaskResponse
	(*connector.GetTaskMetricsResponse)(nil),  // 12: flyteidl.connector.GetTaskMetricsResponse
	(*connector.GetTaskLogsResponse)(nil),     // 13: flyteidl.connector.GetTaskLogsResponse
	(*connector.GetConnectorResponse)(nil),    // 14: flyteidl.connector.GetConnectorResponse
	(*connector.ListConnectorsResponse)(nil),  // 15: flyteidl.connector.ListConnectorsResponse
}
var file_flyteidl_service_connector_proto_depIdxs = []int32{
	0,  // 0: flyteidl.service.SyncConnectorService.ExecuteTaskSync:input_type -> flyteidl.connector.ExecuteTaskSyncRequest
	1,  // 1: flyteidl.service.AsyncConnectorService.CreateTask:input_type -> flyteidl.connector.CreateTaskRequest
	2,  // 2: flyteidl.service.AsyncConnectorService.GetTask:input_type -> flyteidl.connector.GetTaskRequest
	3,  // 3: flyteidl.service.AsyncConnectorService.DeleteTask:input_type -> flyteidl.connector.DeleteTaskRequest
	4,  // 4: flyteidl.service.AsyncConnectorService.GetTaskMetrics:input_type -> flyteidl.connector.GetTaskMetricsRequest
	5,  // 5: flyteidl.service.AsyncConnectorService.GetTaskLogs:input_type -> flyteidl.connector.GetTaskLogsRequest
	6,  // 6: flyteidl.service.ConnectorMetadataService.GetConnector:input_type -> flyteidl.connector.GetConnectorRequest
	7,  // 7: flyteidl.service.ConnectorMetadataService.ListConnectors:input_type -> flyteidl.connector.ListConnectorsRequest
	8,  // 8: flyteidl.service.SyncConnectorService.ExecuteTaskSync:output_type -> flyteidl.connector.ExecuteTaskSyncResponse
	9,  // 9: flyteidl.service.AsyncConnectorService.CreateTask:output_type -> flyteidl.connector.CreateTaskResponse
	10, // 10: flyteidl.service.AsyncConnectorService.GetTask:output_type -> flyteidl.connector.GetTaskResponse
	11, // 11: flyteidl.service.AsyncConnectorService.DeleteTask:output_type -> flyteidl.connector.DeleteTaskResponse
	12, // 12: flyteidl.service.AsyncConnectorService.GetTaskMetrics:output_type -> flyteidl.connector.GetTaskMetricsResponse
	13, // 13: flyteidl.service.AsyncConnectorService.GetTaskLogs:output_type -> flyteidl.connector.GetTaskLogsResponse
	14, // 14: flyteidl.service.ConnectorMetadataService.GetConnector:output_type -> flyteidl.connector.GetConnectorResponse
	15, // 15: flyteidl.service.ConnectorMetadataService.ListConnectors:output_type -> flyteidl.connector.ListConnectorsResponse
	8,  // [8:16] is the sub-list for method output_type
	0,  // [0:8] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_flyteidl_service_connector_proto_init() }
func file_flyteidl_service_connector_proto_init() {
	if File_flyteidl_service_connector_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_flyteidl_service_connector_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   3,
		},
		GoTypes:           file_flyteidl_service_connector_proto_goTypes,
		DependencyIndexes: file_flyteidl_service_connector_proto_depIdxs,
	}.Build()
	File_flyteidl_service_connector_proto = out.File
	file_flyteidl_service_connector_proto_rawDesc = nil
	file_flyteidl_service_connector_proto_goTypes = nil
	file_flyteidl_service_connector_proto_depIdxs = nil
}
