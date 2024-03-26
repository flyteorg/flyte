// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: flyteidl/core/identifier.proto

package core

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Indicates a resource type within Flyte.
type ResourceType int32

const (
	ResourceType_UNSPECIFIED ResourceType = 0
	ResourceType_TASK        ResourceType = 1
	ResourceType_WORKFLOW    ResourceType = 2
	ResourceType_LAUNCH_PLAN ResourceType = 3
	// A dataset represents an entity modeled in Flyte DataCatalog. A Dataset is also a versioned entity and can be a compilation of multiple individual objects.
	// Eventually all Catalog objects should be modeled similar to Flyte Objects. The Dataset entities makes it possible for the UI  and CLI to act on the objects
	// in a similar manner to other Flyte objects
	ResourceType_DATASET ResourceType = 4
)

// Enum value maps for ResourceType.
var (
	ResourceType_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "TASK",
		2: "WORKFLOW",
		3: "LAUNCH_PLAN",
		4: "DATASET",
	}
	ResourceType_value = map[string]int32{
		"UNSPECIFIED": 0,
		"TASK":        1,
		"WORKFLOW":    2,
		"LAUNCH_PLAN": 3,
		"DATASET":     4,
	}
)

func (x ResourceType) Enum() *ResourceType {
	p := new(ResourceType)
	*p = x
	return p
}

func (x ResourceType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ResourceType) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_identifier_proto_enumTypes[0].Descriptor()
}

func (ResourceType) Type() protoreflect.EnumType {
	return &file_flyteidl_core_identifier_proto_enumTypes[0]
}

func (x ResourceType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ResourceType.Descriptor instead.
func (ResourceType) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_identifier_proto_rawDescGZIP(), []int{0}
}

// Encapsulation of fields that uniquely identifies a Flyte resource.
type Identifier struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Identifies the specific type of resource that this identifier corresponds to.
	ResourceType ResourceType `protobuf:"varint,1,opt,name=resource_type,json=resourceType,proto3,enum=flyteidl.core.ResourceType" json:"resource_type,omitempty"`
	// Name of the project the resource belongs to.
	Project string `protobuf:"bytes,2,opt,name=project,proto3" json:"project,omitempty"`
	// Name of the domain the resource belongs to.
	// A domain can be considered as a subset within a specific project.
	Domain string `protobuf:"bytes,3,opt,name=domain,proto3" json:"domain,omitempty"`
	// User provided value for the resource.
	Name string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	// Specific version of the resource.
	Version string `protobuf:"bytes,5,opt,name=version,proto3" json:"version,omitempty"`
	// Optional, org key applied to the resource.
	Org string `protobuf:"bytes,6,opt,name=org,proto3" json:"org,omitempty"`
}

func (x *Identifier) Reset() {
	*x = Identifier{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_identifier_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Identifier) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Identifier) ProtoMessage() {}

func (x *Identifier) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_identifier_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Identifier.ProtoReflect.Descriptor instead.
func (*Identifier) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_identifier_proto_rawDescGZIP(), []int{0}
}

func (x *Identifier) GetResourceType() ResourceType {
	if x != nil {
		return x.ResourceType
	}
	return ResourceType_UNSPECIFIED
}

func (x *Identifier) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *Identifier) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *Identifier) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Identifier) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Identifier) GetOrg() string {
	if x != nil {
		return x.Org
	}
	return ""
}

// Encapsulation of fields that uniquely identifies a Flyte workflow execution
type WorkflowExecutionIdentifier struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the project the resource belongs to.
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// Name of the domain the resource belongs to.
	// A domain can be considered as a subset within a specific project.
	Domain string `protobuf:"bytes,2,opt,name=domain,proto3" json:"domain,omitempty"`
	// User or system provided value for the resource.
	Name string `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	// Optional, org key applied to the resource.
	Org string `protobuf:"bytes,5,opt,name=org,proto3" json:"org,omitempty"`
}

func (x *WorkflowExecutionIdentifier) Reset() {
	*x = WorkflowExecutionIdentifier{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_identifier_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowExecutionIdentifier) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowExecutionIdentifier) ProtoMessage() {}

func (x *WorkflowExecutionIdentifier) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_identifier_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowExecutionIdentifier.ProtoReflect.Descriptor instead.
func (*WorkflowExecutionIdentifier) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_identifier_proto_rawDescGZIP(), []int{1}
}

func (x *WorkflowExecutionIdentifier) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *WorkflowExecutionIdentifier) GetDomain() string {
	if x != nil {
		return x.Domain
	}
	return ""
}

func (x *WorkflowExecutionIdentifier) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *WorkflowExecutionIdentifier) GetOrg() string {
	if x != nil {
		return x.Org
	}
	return ""
}

// Encapsulation of fields that identify a Flyte node execution entity.
type NodeExecutionIdentifier struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId      string                       `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	ExecutionId *WorkflowExecutionIdentifier `protobuf:"bytes,2,opt,name=execution_id,json=executionId,proto3" json:"execution_id,omitempty"`
}

func (x *NodeExecutionIdentifier) Reset() {
	*x = NodeExecutionIdentifier{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_identifier_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeExecutionIdentifier) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeExecutionIdentifier) ProtoMessage() {}

func (x *NodeExecutionIdentifier) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_identifier_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeExecutionIdentifier.ProtoReflect.Descriptor instead.
func (*NodeExecutionIdentifier) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_identifier_proto_rawDescGZIP(), []int{2}
}

func (x *NodeExecutionIdentifier) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *NodeExecutionIdentifier) GetExecutionId() *WorkflowExecutionIdentifier {
	if x != nil {
		return x.ExecutionId
	}
	return nil
}

// Encapsulation of fields that identify a Flyte task execution entity.
type TaskExecutionIdentifier struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskId          *Identifier              `protobuf:"bytes,1,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	NodeExecutionId *NodeExecutionIdentifier `protobuf:"bytes,2,opt,name=node_execution_id,json=nodeExecutionId,proto3" json:"node_execution_id,omitempty"`
	RetryAttempt    uint32                   `protobuf:"varint,3,opt,name=retry_attempt,json=retryAttempt,proto3" json:"retry_attempt,omitempty"`
}

func (x *TaskExecutionIdentifier) Reset() {
	*x = TaskExecutionIdentifier{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_identifier_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskExecutionIdentifier) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskExecutionIdentifier) ProtoMessage() {}

func (x *TaskExecutionIdentifier) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_identifier_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskExecutionIdentifier.ProtoReflect.Descriptor instead.
func (*TaskExecutionIdentifier) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_identifier_proto_rawDescGZIP(), []int{3}
}

func (x *TaskExecutionIdentifier) GetTaskId() *Identifier {
	if x != nil {
		return x.TaskId
	}
	return nil
}

func (x *TaskExecutionIdentifier) GetNodeExecutionId() *NodeExecutionIdentifier {
	if x != nil {
		return x.NodeExecutionId
	}
	return nil
}

func (x *TaskExecutionIdentifier) GetRetryAttempt() uint32 {
	if x != nil {
		return x.RetryAttempt
	}
	return 0
}

// Encapsulation of fields the uniquely identify a signal.
type SignalIdentifier struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique identifier for a signal.
	SignalId string `protobuf:"bytes,1,opt,name=signal_id,json=signalId,proto3" json:"signal_id,omitempty"`
	// Identifies the Flyte workflow execution this signal belongs to.
	ExecutionId *WorkflowExecutionIdentifier `protobuf:"bytes,2,opt,name=execution_id,json=executionId,proto3" json:"execution_id,omitempty"`
}

func (x *SignalIdentifier) Reset() {
	*x = SignalIdentifier{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_identifier_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignalIdentifier) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalIdentifier) ProtoMessage() {}

func (x *SignalIdentifier) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_identifier_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalIdentifier.ProtoReflect.Descriptor instead.
func (*SignalIdentifier) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_identifier_proto_rawDescGZIP(), []int{4}
}

func (x *SignalIdentifier) GetSignalId() string {
	if x != nil {
		return x.SignalId
	}
	return ""
}

func (x *SignalIdentifier) GetExecutionId() *WorkflowExecutionIdentifier {
	if x != nil {
		return x.ExecutionId
	}
	return nil
}

var File_flyteidl_core_identifier_proto protoreflect.FileDescriptor

var file_flyteidl_core_identifier_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0d, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x22,
	0xc0, 0x01, 0x0a, 0x0a, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x40,
	0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x6f,
	0x6d, 0x61, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x6f, 0x6d, 0x61,
	0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x10, 0x0a, 0x03, 0x6f, 0x72, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6f,
	0x72, 0x67, 0x22, 0x75, 0x0a, 0x1b, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78,
	0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65,
	0x72, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x64,
	0x6f, 0x6d, 0x61, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x6f, 0x6d,
	0x61, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6f, 0x72, 0x67, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6f, 0x72, 0x67, 0x22, 0x81, 0x01, 0x0a, 0x17, 0x4e, 0x6f,
	0x64, 0x65, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x4d,
	0x0a, 0x0c, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e,
	0x63, 0x6f, 0x72, 0x65, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x52, 0x0b, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22, 0xc6, 0x01,
	0x0a, 0x17, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x32, 0x0a, 0x07, 0x74, 0x61, 0x73,
	0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x66, 0x6c, 0x79,
	0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x66, 0x69, 0x65, 0x72, 0x52, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x52, 0x0a,
	0x11, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65,
	0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x45, 0x78, 0x65,
	0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x52, 0x0f, 0x6e, 0x6f, 0x64, 0x65, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x12, 0x23, 0x0a, 0x0d, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x61, 0x74, 0x74, 0x65, 0x6d,
	0x70, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41,
	0x74, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x22, 0x7e, 0x0a, 0x10, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73,
	0x69, 0x67, 0x6e, 0x61, 0x6c, 0x49, 0x64, 0x12, 0x4d, 0x0a, 0x0c, 0x65, 0x78, 0x65, 0x63, 0x75,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e,
	0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x57, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x52, 0x0b, 0x65, 0x78, 0x65, 0x63, 0x75,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x2a, 0x55, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x54, 0x41, 0x53, 0x4b, 0x10,
	0x01, 0x12, 0x0c, 0x0a, 0x08, 0x57, 0x4f, 0x52, 0x4b, 0x46, 0x4c, 0x4f, 0x57, 0x10, 0x02, 0x12,
	0x0f, 0x0a, 0x0b, 0x4c, 0x41, 0x55, 0x4e, 0x43, 0x48, 0x5f, 0x50, 0x4c, 0x41, 0x4e, 0x10, 0x03,
	0x12, 0x0b, 0x0a, 0x07, 0x44, 0x41, 0x54, 0x41, 0x53, 0x45, 0x54, 0x10, 0x04, 0x42, 0xb5, 0x01,
	0x0a, 0x11, 0x63, 0x6f, 0x6d, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63,
	0x6f, 0x72, 0x65, 0x42, 0x0f, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x6f, 0x72, 0x67, 0x2f, 0x66, 0x6c, 0x79, 0x74,
	0x65, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70,
	0x62, 0x2d, 0x67, 0x6f, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x63, 0x6f,
	0x72, 0x65, 0xa2, 0x02, 0x03, 0x46, 0x43, 0x58, 0xaa, 0x02, 0x0d, 0x46, 0x6c, 0x79, 0x74, 0x65,
	0x69, 0x64, 0x6c, 0x2e, 0x43, 0x6f, 0x72, 0x65, 0xca, 0x02, 0x0d, 0x46, 0x6c, 0x79, 0x74, 0x65,
	0x69, 0x64, 0x6c, 0x5c, 0x43, 0x6f, 0x72, 0x65, 0xe2, 0x02, 0x19, 0x46, 0x6c, 0x79, 0x74, 0x65,
	0x69, 0x64, 0x6c, 0x5c, 0x43, 0x6f, 0x72, 0x65, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0e, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x3a,
	0x3a, 0x43, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_flyteidl_core_identifier_proto_rawDescOnce sync.Once
	file_flyteidl_core_identifier_proto_rawDescData = file_flyteidl_core_identifier_proto_rawDesc
)

func file_flyteidl_core_identifier_proto_rawDescGZIP() []byte {
	file_flyteidl_core_identifier_proto_rawDescOnce.Do(func() {
		file_flyteidl_core_identifier_proto_rawDescData = protoimpl.X.CompressGZIP(file_flyteidl_core_identifier_proto_rawDescData)
	})
	return file_flyteidl_core_identifier_proto_rawDescData
}

var file_flyteidl_core_identifier_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_flyteidl_core_identifier_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_flyteidl_core_identifier_proto_goTypes = []interface{}{
	(ResourceType)(0),                   // 0: flyteidl.core.ResourceType
	(*Identifier)(nil),                  // 1: flyteidl.core.Identifier
	(*WorkflowExecutionIdentifier)(nil), // 2: flyteidl.core.WorkflowExecutionIdentifier
	(*NodeExecutionIdentifier)(nil),     // 3: flyteidl.core.NodeExecutionIdentifier
	(*TaskExecutionIdentifier)(nil),     // 4: flyteidl.core.TaskExecutionIdentifier
	(*SignalIdentifier)(nil),            // 5: flyteidl.core.SignalIdentifier
}
var file_flyteidl_core_identifier_proto_depIdxs = []int32{
	0, // 0: flyteidl.core.Identifier.resource_type:type_name -> flyteidl.core.ResourceType
	2, // 1: flyteidl.core.NodeExecutionIdentifier.execution_id:type_name -> flyteidl.core.WorkflowExecutionIdentifier
	1, // 2: flyteidl.core.TaskExecutionIdentifier.task_id:type_name -> flyteidl.core.Identifier
	3, // 3: flyteidl.core.TaskExecutionIdentifier.node_execution_id:type_name -> flyteidl.core.NodeExecutionIdentifier
	2, // 4: flyteidl.core.SignalIdentifier.execution_id:type_name -> flyteidl.core.WorkflowExecutionIdentifier
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_flyteidl_core_identifier_proto_init() }
func file_flyteidl_core_identifier_proto_init() {
	if File_flyteidl_core_identifier_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_flyteidl_core_identifier_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Identifier); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_flyteidl_core_identifier_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowExecutionIdentifier); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_flyteidl_core_identifier_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeExecutionIdentifier); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_flyteidl_core_identifier_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskExecutionIdentifier); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_flyteidl_core_identifier_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignalIdentifier); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_flyteidl_core_identifier_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_flyteidl_core_identifier_proto_goTypes,
		DependencyIndexes: file_flyteidl_core_identifier_proto_depIdxs,
		EnumInfos:         file_flyteidl_core_identifier_proto_enumTypes,
		MessageInfos:      file_flyteidl_core_identifier_proto_msgTypes,
	}.Build()
	File_flyteidl_core_identifier_proto = out.File
	file_flyteidl_core_identifier_proto_rawDesc = nil
	file_flyteidl_core_identifier_proto_goTypes = nil
	file_flyteidl_core_identifier_proto_depIdxs = nil
}
