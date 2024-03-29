// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: flyteidl/admin/event.proto

package admin

import (
	event "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
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

// Indicates that a sent event was not used to update execution state due to
// the referenced execution already being terminated (and therefore ineligible
// for further state transitions).
type EventErrorAlreadyInTerminalState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// +required
	CurrentPhase string `protobuf:"bytes,1,opt,name=current_phase,json=currentPhase,proto3" json:"current_phase,omitempty"`
}

func (x *EventErrorAlreadyInTerminalState) Reset() {
	*x = EventErrorAlreadyInTerminalState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventErrorAlreadyInTerminalState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventErrorAlreadyInTerminalState) ProtoMessage() {}

func (x *EventErrorAlreadyInTerminalState) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventErrorAlreadyInTerminalState.ProtoReflect.Descriptor instead.
func (*EventErrorAlreadyInTerminalState) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{0}
}

func (x *EventErrorAlreadyInTerminalState) GetCurrentPhase() string {
	if x != nil {
		return x.CurrentPhase
	}
	return ""
}

// Indicates an event was rejected because it came from a different cluster than
// is on record as running the execution.
type EventErrorIncompatibleCluster struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The cluster which has been recorded as processing the execution.
	// +required
	Cluster string `protobuf:"bytes,1,opt,name=cluster,proto3" json:"cluster,omitempty"`
}

func (x *EventErrorIncompatibleCluster) Reset() {
	*x = EventErrorIncompatibleCluster{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventErrorIncompatibleCluster) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventErrorIncompatibleCluster) ProtoMessage() {}

func (x *EventErrorIncompatibleCluster) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventErrorIncompatibleCluster.ProtoReflect.Descriptor instead.
func (*EventErrorIncompatibleCluster) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{1}
}

func (x *EventErrorIncompatibleCluster) GetCluster() string {
	if x != nil {
		return x.Cluster
	}
	return ""
}

// Indicates why a sent event was not used to update execution.
type EventFailureReason struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// +required
	//
	// Types that are assignable to Reason:
	//
	//	*EventFailureReason_AlreadyInTerminalState
	//	*EventFailureReason_IncompatibleCluster
	Reason isEventFailureReason_Reason `protobuf_oneof:"reason"`
}

func (x *EventFailureReason) Reset() {
	*x = EventFailureReason{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventFailureReason) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventFailureReason) ProtoMessage() {}

func (x *EventFailureReason) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventFailureReason.ProtoReflect.Descriptor instead.
func (*EventFailureReason) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{2}
}

func (m *EventFailureReason) GetReason() isEventFailureReason_Reason {
	if m != nil {
		return m.Reason
	}
	return nil
}

func (x *EventFailureReason) GetAlreadyInTerminalState() *EventErrorAlreadyInTerminalState {
	if x, ok := x.GetReason().(*EventFailureReason_AlreadyInTerminalState); ok {
		return x.AlreadyInTerminalState
	}
	return nil
}

func (x *EventFailureReason) GetIncompatibleCluster() *EventErrorIncompatibleCluster {
	if x, ok := x.GetReason().(*EventFailureReason_IncompatibleCluster); ok {
		return x.IncompatibleCluster
	}
	return nil
}

type isEventFailureReason_Reason interface {
	isEventFailureReason_Reason()
}

type EventFailureReason_AlreadyInTerminalState struct {
	AlreadyInTerminalState *EventErrorAlreadyInTerminalState `protobuf:"bytes,1,opt,name=already_in_terminal_state,json=alreadyInTerminalState,proto3,oneof"`
}

type EventFailureReason_IncompatibleCluster struct {
	IncompatibleCluster *EventErrorIncompatibleCluster `protobuf:"bytes,2,opt,name=incompatible_cluster,json=incompatibleCluster,proto3,oneof"`
}

func (*EventFailureReason_AlreadyInTerminalState) isEventFailureReason_Reason() {}

func (*EventFailureReason_IncompatibleCluster) isEventFailureReason_Reason() {}

// Request to send a notification that a workflow execution event has occurred.
type WorkflowExecutionEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique ID for this request that can be traced between services
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// Details about the event that occurred.
	Event *event.WorkflowExecutionEvent `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *WorkflowExecutionEventRequest) Reset() {
	*x = WorkflowExecutionEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowExecutionEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowExecutionEventRequest) ProtoMessage() {}

func (x *WorkflowExecutionEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowExecutionEventRequest.ProtoReflect.Descriptor instead.
func (*WorkflowExecutionEventRequest) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{3}
}

func (x *WorkflowExecutionEventRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *WorkflowExecutionEventRequest) GetEvent() *event.WorkflowExecutionEvent {
	if x != nil {
		return x.Event
	}
	return nil
}

type WorkflowExecutionEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WorkflowExecutionEventResponse) Reset() {
	*x = WorkflowExecutionEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowExecutionEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowExecutionEventResponse) ProtoMessage() {}

func (x *WorkflowExecutionEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowExecutionEventResponse.ProtoReflect.Descriptor instead.
func (*WorkflowExecutionEventResponse) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{4}
}

// Request to send a notification that a node execution event has occurred.
type NodeExecutionEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique ID for this request that can be traced between services
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// Details about the event that occurred.
	Event *event.NodeExecutionEvent `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *NodeExecutionEventRequest) Reset() {
	*x = NodeExecutionEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeExecutionEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeExecutionEventRequest) ProtoMessage() {}

func (x *NodeExecutionEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeExecutionEventRequest.ProtoReflect.Descriptor instead.
func (*NodeExecutionEventRequest) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{5}
}

func (x *NodeExecutionEventRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *NodeExecutionEventRequest) GetEvent() *event.NodeExecutionEvent {
	if x != nil {
		return x.Event
	}
	return nil
}

type NodeExecutionEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NodeExecutionEventResponse) Reset() {
	*x = NodeExecutionEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeExecutionEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeExecutionEventResponse) ProtoMessage() {}

func (x *NodeExecutionEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeExecutionEventResponse.ProtoReflect.Descriptor instead.
func (*NodeExecutionEventResponse) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{6}
}

// Request to send a notification that a task execution event has occurred.
type TaskExecutionEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique ID for this request that can be traced between services
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// Details about the event that occurred.
	Event *event.TaskExecutionEvent `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *TaskExecutionEventRequest) Reset() {
	*x = TaskExecutionEventRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskExecutionEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskExecutionEventRequest) ProtoMessage() {}

func (x *TaskExecutionEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskExecutionEventRequest.ProtoReflect.Descriptor instead.
func (*TaskExecutionEventRequest) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{7}
}

func (x *TaskExecutionEventRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *TaskExecutionEventRequest) GetEvent() *event.TaskExecutionEvent {
	if x != nil {
		return x.Event
	}
	return nil
}

type TaskExecutionEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TaskExecutionEventResponse) Reset() {
	*x = TaskExecutionEventResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_admin_event_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskExecutionEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskExecutionEventResponse) ProtoMessage() {}

func (x *TaskExecutionEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_admin_event_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskExecutionEventResponse.ProtoReflect.Descriptor instead.
func (*TaskExecutionEventResponse) Descriptor() ([]byte, []int) {
	return file_flyteidl_admin_event_proto_rawDescGZIP(), []int{8}
}

var File_flyteidl_admin_event_proto protoreflect.FileDescriptor

var file_flyteidl_admin_event_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x66, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x1a, 0x1a, 0x66, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2f, 0x65, 0x76, 0x65,
	0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x47, 0x0a, 0x20, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x41, 0x6c, 0x72, 0x65, 0x61, 0x64, 0x79, 0x49, 0x6e, 0x54,
	0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x23, 0x0a, 0x0d,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x68, 0x61, 0x73, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x50, 0x68, 0x61, 0x73,
	0x65, 0x22, 0x39, 0x0a, 0x1d, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x49,
	0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62, 0x6c, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x22, 0xf1, 0x01, 0x0a,
	0x12, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x46, 0x61, 0x69, 0x6c, 0x75, 0x72, 0x65, 0x52, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x12, 0x6d, 0x0a, 0x19, 0x61, 0x6c, 0x72, 0x65, 0x61, 0x64, 0x79, 0x5f, 0x69,
	0x6e, 0x5f, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64,
	0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x41, 0x6c, 0x72, 0x65, 0x61, 0x64, 0x79, 0x49, 0x6e, 0x54, 0x65, 0x72, 0x6d, 0x69,
	0x6e, 0x61, 0x6c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x48, 0x00, 0x52, 0x16, 0x61, 0x6c, 0x72, 0x65,
	0x61, 0x64, 0x79, 0x49, 0x6e, 0x54, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x62, 0x0a, 0x14, 0x69, 0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62,
	0x6c, 0x65, 0x5f, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x2d, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x49, 0x6e, 0x63, 0x6f,
	0x6d, 0x70, 0x61, 0x74, 0x69, 0x62, 0x6c, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x48,
	0x00, 0x52, 0x13, 0x69, 0x6e, 0x63, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62, 0x6c, 0x65, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x42, 0x08, 0x0a, 0x06, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e,
	0x22, 0x7c, 0x0a, 0x1d, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64,
	0x12, 0x3c, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x26, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69,
	0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x20,
	0x0a, 0x1e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74,
	0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x74, 0x0a, 0x19, 0x4e, 0x6f, 0x64, 0x65, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f,
	0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a,
	0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x38, 0x0a, 0x05,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x66, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x4e, 0x6f, 0x64,
	0x65, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52,
	0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x1c, 0x0a, 0x1a, 0x4e, 0x6f, 0x64, 0x65, 0x45, 0x78,
	0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x74, 0x0a, 0x19, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64,
	0x12, 0x38, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x22, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x2e, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x1c, 0x0a, 0x1a, 0x54, 0x61,
	0x73, 0x6b, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xb6, 0x01, 0x0a, 0x12, 0x63, 0x6f, 0x6d,
	0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x42,
	0x0a, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x6f,
	0x72, 0x67, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64,
	0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x62, 0x2d, 0x67, 0x6f, 0x2f, 0x66, 0x6c, 0x79, 0x74,
	0x65, 0x69, 0x64, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0xa2, 0x02, 0x03, 0x46, 0x41, 0x58,
	0xaa, 0x02, 0x0e, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0xca, 0x02, 0x0e, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x41, 0x64, 0x6d,
	0x69, 0x6e, 0xe2, 0x02, 0x1a, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c, 0x41, 0x64,
	0x6d, 0x69, 0x6e, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x0f, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x3a, 0x3a, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_flyteidl_admin_event_proto_rawDescOnce sync.Once
	file_flyteidl_admin_event_proto_rawDescData = file_flyteidl_admin_event_proto_rawDesc
)

func file_flyteidl_admin_event_proto_rawDescGZIP() []byte {
	file_flyteidl_admin_event_proto_rawDescOnce.Do(func() {
		file_flyteidl_admin_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_flyteidl_admin_event_proto_rawDescData)
	})
	return file_flyteidl_admin_event_proto_rawDescData
}

var file_flyteidl_admin_event_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_flyteidl_admin_event_proto_goTypes = []interface{}{
	(*EventErrorAlreadyInTerminalState)(nil), // 0: flyteidl.admin.EventErrorAlreadyInTerminalState
	(*EventErrorIncompatibleCluster)(nil),    // 1: flyteidl.admin.EventErrorIncompatibleCluster
	(*EventFailureReason)(nil),               // 2: flyteidl.admin.EventFailureReason
	(*WorkflowExecutionEventRequest)(nil),    // 3: flyteidl.admin.WorkflowExecutionEventRequest
	(*WorkflowExecutionEventResponse)(nil),   // 4: flyteidl.admin.WorkflowExecutionEventResponse
	(*NodeExecutionEventRequest)(nil),        // 5: flyteidl.admin.NodeExecutionEventRequest
	(*NodeExecutionEventResponse)(nil),       // 6: flyteidl.admin.NodeExecutionEventResponse
	(*TaskExecutionEventRequest)(nil),        // 7: flyteidl.admin.TaskExecutionEventRequest
	(*TaskExecutionEventResponse)(nil),       // 8: flyteidl.admin.TaskExecutionEventResponse
	(*event.WorkflowExecutionEvent)(nil),     // 9: flyteidl.event.WorkflowExecutionEvent
	(*event.NodeExecutionEvent)(nil),         // 10: flyteidl.event.NodeExecutionEvent
	(*event.TaskExecutionEvent)(nil),         // 11: flyteidl.event.TaskExecutionEvent
}
var file_flyteidl_admin_event_proto_depIdxs = []int32{
	0,  // 0: flyteidl.admin.EventFailureReason.already_in_terminal_state:type_name -> flyteidl.admin.EventErrorAlreadyInTerminalState
	1,  // 1: flyteidl.admin.EventFailureReason.incompatible_cluster:type_name -> flyteidl.admin.EventErrorIncompatibleCluster
	9,  // 2: flyteidl.admin.WorkflowExecutionEventRequest.event:type_name -> flyteidl.event.WorkflowExecutionEvent
	10, // 3: flyteidl.admin.NodeExecutionEventRequest.event:type_name -> flyteidl.event.NodeExecutionEvent
	11, // 4: flyteidl.admin.TaskExecutionEventRequest.event:type_name -> flyteidl.event.TaskExecutionEvent
	5,  // [5:5] is the sub-list for method output_type
	5,  // [5:5] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_flyteidl_admin_event_proto_init() }
func file_flyteidl_admin_event_proto_init() {
	if File_flyteidl_admin_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_flyteidl_admin_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventErrorAlreadyInTerminalState); i {
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
		file_flyteidl_admin_event_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventErrorIncompatibleCluster); i {
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
		file_flyteidl_admin_event_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventFailureReason); i {
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
		file_flyteidl_admin_event_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowExecutionEventRequest); i {
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
		file_flyteidl_admin_event_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowExecutionEventResponse); i {
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
		file_flyteidl_admin_event_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeExecutionEventRequest); i {
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
		file_flyteidl_admin_event_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeExecutionEventResponse); i {
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
		file_flyteidl_admin_event_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskExecutionEventRequest); i {
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
		file_flyteidl_admin_event_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskExecutionEventResponse); i {
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
	file_flyteidl_admin_event_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*EventFailureReason_AlreadyInTerminalState)(nil),
		(*EventFailureReason_IncompatibleCluster)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_flyteidl_admin_event_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_flyteidl_admin_event_proto_goTypes,
		DependencyIndexes: file_flyteidl_admin_event_proto_depIdxs,
		MessageInfos:      file_flyteidl_admin_event_proto_msgTypes,
	}.Build()
	File_flyteidl_admin_event_proto = out.File
	file_flyteidl_admin_event_proto_rawDesc = nil
	file_flyteidl_admin_event_proto_goTypes = nil
	file_flyteidl_admin_event_proto_depIdxs = nil
}
