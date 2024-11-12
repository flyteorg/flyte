// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        (unknown)
// source: flyteidl/core/execution.proto

package core

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type WorkflowExecution_Phase int32

const (
	WorkflowExecution_UNDEFINED  WorkflowExecution_Phase = 0
	WorkflowExecution_QUEUED     WorkflowExecution_Phase = 1
	WorkflowExecution_RUNNING    WorkflowExecution_Phase = 2
	WorkflowExecution_SUCCEEDING WorkflowExecution_Phase = 3
	WorkflowExecution_SUCCEEDED  WorkflowExecution_Phase = 4
	WorkflowExecution_FAILING    WorkflowExecution_Phase = 5
	WorkflowExecution_FAILED     WorkflowExecution_Phase = 6
	WorkflowExecution_ABORTED    WorkflowExecution_Phase = 7
	WorkflowExecution_TIMED_OUT  WorkflowExecution_Phase = 8
	WorkflowExecution_ABORTING   WorkflowExecution_Phase = 9
)

// Enum value maps for WorkflowExecution_Phase.
var (
	WorkflowExecution_Phase_name = map[int32]string{
		0: "UNDEFINED",
		1: "QUEUED",
		2: "RUNNING",
		3: "SUCCEEDING",
		4: "SUCCEEDED",
		5: "FAILING",
		6: "FAILED",
		7: "ABORTED",
		8: "TIMED_OUT",
		9: "ABORTING",
	}
	WorkflowExecution_Phase_value = map[string]int32{
		"UNDEFINED":  0,
		"QUEUED":     1,
		"RUNNING":    2,
		"SUCCEEDING": 3,
		"SUCCEEDED":  4,
		"FAILING":    5,
		"FAILED":     6,
		"ABORTED":    7,
		"TIMED_OUT":  8,
		"ABORTING":   9,
	}
)

func (x WorkflowExecution_Phase) Enum() *WorkflowExecution_Phase {
	p := new(WorkflowExecution_Phase)
	*p = x
	return p
}

func (x WorkflowExecution_Phase) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (WorkflowExecution_Phase) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_execution_proto_enumTypes[0].Descriptor()
}

func (WorkflowExecution_Phase) Type() protoreflect.EnumType {
	return &file_flyteidl_core_execution_proto_enumTypes[0]
}

func (x WorkflowExecution_Phase) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use WorkflowExecution_Phase.Descriptor instead.
func (WorkflowExecution_Phase) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{0, 0}
}

type NodeExecution_Phase int32

const (
	NodeExecution_UNDEFINED       NodeExecution_Phase = 0
	NodeExecution_QUEUED          NodeExecution_Phase = 1
	NodeExecution_RUNNING         NodeExecution_Phase = 2
	NodeExecution_SUCCEEDED       NodeExecution_Phase = 3
	NodeExecution_FAILING         NodeExecution_Phase = 4
	NodeExecution_FAILED          NodeExecution_Phase = 5
	NodeExecution_ABORTED         NodeExecution_Phase = 6
	NodeExecution_SKIPPED         NodeExecution_Phase = 7
	NodeExecution_TIMED_OUT       NodeExecution_Phase = 8
	NodeExecution_DYNAMIC_RUNNING NodeExecution_Phase = 9
	NodeExecution_RECOVERED       NodeExecution_Phase = 10
)

// Enum value maps for NodeExecution_Phase.
var (
	NodeExecution_Phase_name = map[int32]string{
		0:  "UNDEFINED",
		1:  "QUEUED",
		2:  "RUNNING",
		3:  "SUCCEEDED",
		4:  "FAILING",
		5:  "FAILED",
		6:  "ABORTED",
		7:  "SKIPPED",
		8:  "TIMED_OUT",
		9:  "DYNAMIC_RUNNING",
		10: "RECOVERED",
	}
	NodeExecution_Phase_value = map[string]int32{
		"UNDEFINED":       0,
		"QUEUED":          1,
		"RUNNING":         2,
		"SUCCEEDED":       3,
		"FAILING":         4,
		"FAILED":          5,
		"ABORTED":         6,
		"SKIPPED":         7,
		"TIMED_OUT":       8,
		"DYNAMIC_RUNNING": 9,
		"RECOVERED":       10,
	}
)

func (x NodeExecution_Phase) Enum() *NodeExecution_Phase {
	p := new(NodeExecution_Phase)
	*p = x
	return p
}

func (x NodeExecution_Phase) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NodeExecution_Phase) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_execution_proto_enumTypes[1].Descriptor()
}

func (NodeExecution_Phase) Type() protoreflect.EnumType {
	return &file_flyteidl_core_execution_proto_enumTypes[1]
}

func (x NodeExecution_Phase) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NodeExecution_Phase.Descriptor instead.
func (NodeExecution_Phase) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{1, 0}
}

type TaskExecution_Phase int32

const (
	TaskExecution_UNDEFINED TaskExecution_Phase = 0
	TaskExecution_QUEUED    TaskExecution_Phase = 1
	TaskExecution_RUNNING   TaskExecution_Phase = 2
	TaskExecution_SUCCEEDED TaskExecution_Phase = 3
	TaskExecution_ABORTED   TaskExecution_Phase = 4
	TaskExecution_FAILED    TaskExecution_Phase = 5
	// To indicate cases where task is initializing, like: ErrImagePull, ContainerCreating, PodInitializing
	TaskExecution_INITIALIZING TaskExecution_Phase = 6
	// To address cases, where underlying resource is not available: Backoff error, Resource quota exceeded
	TaskExecution_WAITING_FOR_RESOURCES TaskExecution_Phase = 7
)

// Enum value maps for TaskExecution_Phase.
var (
	TaskExecution_Phase_name = map[int32]string{
		0: "UNDEFINED",
		1: "QUEUED",
		2: "RUNNING",
		3: "SUCCEEDED",
		4: "ABORTED",
		5: "FAILED",
		6: "INITIALIZING",
		7: "WAITING_FOR_RESOURCES",
	}
	TaskExecution_Phase_value = map[string]int32{
		"UNDEFINED":             0,
		"QUEUED":                1,
		"RUNNING":               2,
		"SUCCEEDED":             3,
		"ABORTED":               4,
		"FAILED":                5,
		"INITIALIZING":          6,
		"WAITING_FOR_RESOURCES": 7,
	}
)

func (x TaskExecution_Phase) Enum() *TaskExecution_Phase {
	p := new(TaskExecution_Phase)
	*p = x
	return p
}

func (x TaskExecution_Phase) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskExecution_Phase) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_execution_proto_enumTypes[2].Descriptor()
}

func (TaskExecution_Phase) Type() protoreflect.EnumType {
	return &file_flyteidl_core_execution_proto_enumTypes[2]
}

func (x TaskExecution_Phase) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskExecution_Phase.Descriptor instead.
func (TaskExecution_Phase) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{2, 0}
}

// Error type: System or User
type ExecutionError_ErrorKind int32

const (
	ExecutionError_UNKNOWN ExecutionError_ErrorKind = 0
	ExecutionError_USER    ExecutionError_ErrorKind = 1
	ExecutionError_SYSTEM  ExecutionError_ErrorKind = 2
)

// Enum value maps for ExecutionError_ErrorKind.
var (
	ExecutionError_ErrorKind_name = map[int32]string{
		0: "UNKNOWN",
		1: "USER",
		2: "SYSTEM",
	}
	ExecutionError_ErrorKind_value = map[string]int32{
		"UNKNOWN": 0,
		"USER":    1,
		"SYSTEM":  2,
	}
)

func (x ExecutionError_ErrorKind) Enum() *ExecutionError_ErrorKind {
	p := new(ExecutionError_ErrorKind)
	*p = x
	return p
}

func (x ExecutionError_ErrorKind) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ExecutionError_ErrorKind) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_execution_proto_enumTypes[3].Descriptor()
}

func (ExecutionError_ErrorKind) Type() protoreflect.EnumType {
	return &file_flyteidl_core_execution_proto_enumTypes[3]
}

func (x ExecutionError_ErrorKind) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ExecutionError_ErrorKind.Descriptor instead.
func (ExecutionError_ErrorKind) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{3, 0}
}

type TaskLog_MessageFormat int32

const (
	TaskLog_UNKNOWN TaskLog_MessageFormat = 0
	TaskLog_CSV     TaskLog_MessageFormat = 1
	TaskLog_JSON    TaskLog_MessageFormat = 2
)

// Enum value maps for TaskLog_MessageFormat.
var (
	TaskLog_MessageFormat_name = map[int32]string{
		0: "UNKNOWN",
		1: "CSV",
		2: "JSON",
	}
	TaskLog_MessageFormat_value = map[string]int32{
		"UNKNOWN": 0,
		"CSV":     1,
		"JSON":    2,
	}
)

func (x TaskLog_MessageFormat) Enum() *TaskLog_MessageFormat {
	p := new(TaskLog_MessageFormat)
	*p = x
	return p
}

func (x TaskLog_MessageFormat) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskLog_MessageFormat) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_execution_proto_enumTypes[4].Descriptor()
}

func (TaskLog_MessageFormat) Type() protoreflect.EnumType {
	return &file_flyteidl_core_execution_proto_enumTypes[4]
}

func (x TaskLog_MessageFormat) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskLog_MessageFormat.Descriptor instead.
func (TaskLog_MessageFormat) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{4, 0}
}

type QualityOfService_Tier int32

const (
	// Default: no quality of service specified.
	QualityOfService_UNDEFINED QualityOfService_Tier = 0
	QualityOfService_HIGH      QualityOfService_Tier = 1
	QualityOfService_MEDIUM    QualityOfService_Tier = 2
	QualityOfService_LOW       QualityOfService_Tier = 3
)

// Enum value maps for QualityOfService_Tier.
var (
	QualityOfService_Tier_name = map[int32]string{
		0: "UNDEFINED",
		1: "HIGH",
		2: "MEDIUM",
		3: "LOW",
	}
	QualityOfService_Tier_value = map[string]int32{
		"UNDEFINED": 0,
		"HIGH":      1,
		"MEDIUM":    2,
		"LOW":       3,
	}
)

func (x QualityOfService_Tier) Enum() *QualityOfService_Tier {
	p := new(QualityOfService_Tier)
	*p = x
	return p
}

func (x QualityOfService_Tier) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QualityOfService_Tier) Descriptor() protoreflect.EnumDescriptor {
	return file_flyteidl_core_execution_proto_enumTypes[5].Descriptor()
}

func (QualityOfService_Tier) Type() protoreflect.EnumType {
	return &file_flyteidl_core_execution_proto_enumTypes[5]
}

func (x QualityOfService_Tier) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QualityOfService_Tier.Descriptor instead.
func (QualityOfService_Tier) EnumDescriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{6, 0}
}

// Indicates various phases of Workflow Execution
type WorkflowExecution struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *WorkflowExecution) Reset() {
	*x = WorkflowExecution{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkflowExecution) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowExecution) ProtoMessage() {}

func (x *WorkflowExecution) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowExecution.ProtoReflect.Descriptor instead.
func (*WorkflowExecution) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{0}
}

// Indicates various phases of Node Execution that only include the time spent to run the nodes/workflows
type NodeExecution struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NodeExecution) Reset() {
	*x = NodeExecution{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeExecution) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeExecution) ProtoMessage() {}

func (x *NodeExecution) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeExecution.ProtoReflect.Descriptor instead.
func (*NodeExecution) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{1}
}

// Phases that task plugins can go through. Not all phases may be applicable to a specific plugin task,
// but this is the cumulative list that customers may want to know about for their task.
type TaskExecution struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TaskExecution) Reset() {
	*x = TaskExecution{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskExecution) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskExecution) ProtoMessage() {}

func (x *TaskExecution) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskExecution.ProtoReflect.Descriptor instead.
func (*TaskExecution) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{2}
}

// Represents the error message from the execution.
type ExecutionError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Error code indicates a grouping of a type of error.
	// More Info: <Link>
	Code string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	// Detailed description of the error - including stack trace.
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	// Full error contents accessible via a URI
	ErrorUri string                   `protobuf:"bytes,3,opt,name=error_uri,json=errorUri,proto3" json:"error_uri,omitempty"`
	Kind     ExecutionError_ErrorKind `protobuf:"varint,4,opt,name=kind,proto3,enum=flyteidl.core.ExecutionError_ErrorKind" json:"kind,omitempty"`
	// Timestamp of the error
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Worker that generated the error
	Worker string `protobuf:"bytes,6,opt,name=worker,proto3" json:"worker,omitempty"`
}

func (x *ExecutionError) Reset() {
	*x = ExecutionError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecutionError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecutionError) ProtoMessage() {}

func (x *ExecutionError) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecutionError.ProtoReflect.Descriptor instead.
func (*ExecutionError) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{3}
}

func (x *ExecutionError) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *ExecutionError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ExecutionError) GetErrorUri() string {
	if x != nil {
		return x.ErrorUri
	}
	return ""
}

func (x *ExecutionError) GetKind() ExecutionError_ErrorKind {
	if x != nil {
		return x.Kind
	}
	return ExecutionError_UNKNOWN
}

func (x *ExecutionError) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *ExecutionError) GetWorker() string {
	if x != nil {
		return x.Worker
	}
	return ""
}

// Log information for the task that is specific to a log sink
// When our log story is flushed out, we may have more metadata here like log link expiry
type TaskLog struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uri              string                `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty"`
	Name             string                `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	MessageFormat    TaskLog_MessageFormat `protobuf:"varint,3,opt,name=message_format,json=messageFormat,proto3,enum=flyteidl.core.TaskLog_MessageFormat" json:"message_format,omitempty"`
	Ttl              *durationpb.Duration  `protobuf:"bytes,4,opt,name=ttl,proto3" json:"ttl,omitempty"`
	ShowWhilePending bool                  `protobuf:"varint,5,opt,name=ShowWhilePending,proto3" json:"ShowWhilePending,omitempty"`
	HideOnceFinished bool                  `protobuf:"varint,6,opt,name=HideOnceFinished,proto3" json:"HideOnceFinished,omitempty"`
}

func (x *TaskLog) Reset() {
	*x = TaskLog{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskLog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskLog) ProtoMessage() {}

func (x *TaskLog) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskLog.ProtoReflect.Descriptor instead.
func (*TaskLog) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{4}
}

func (x *TaskLog) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *TaskLog) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TaskLog) GetMessageFormat() TaskLog_MessageFormat {
	if x != nil {
		return x.MessageFormat
	}
	return TaskLog_UNKNOWN
}

func (x *TaskLog) GetTtl() *durationpb.Duration {
	if x != nil {
		return x.Ttl
	}
	return nil
}

func (x *TaskLog) GetShowWhilePending() bool {
	if x != nil {
		return x.ShowWhilePending
	}
	return false
}

func (x *TaskLog) GetHideOnceFinished() bool {
	if x != nil {
		return x.HideOnceFinished
	}
	return false
}

// Represents customized execution run-time attributes.
type QualityOfServiceSpec struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Indicates how much queueing delay an execution can tolerate.
	QueueingBudget *durationpb.Duration `protobuf:"bytes,1,opt,name=queueing_budget,json=queueingBudget,proto3" json:"queueing_budget,omitempty"`
}

func (x *QualityOfServiceSpec) Reset() {
	*x = QualityOfServiceSpec{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QualityOfServiceSpec) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QualityOfServiceSpec) ProtoMessage() {}

func (x *QualityOfServiceSpec) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QualityOfServiceSpec.ProtoReflect.Descriptor instead.
func (*QualityOfServiceSpec) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{5}
}

func (x *QualityOfServiceSpec) GetQueueingBudget() *durationpb.Duration {
	if x != nil {
		return x.QueueingBudget
	}
	return nil
}

// Indicates the priority of an execution.
type QualityOfService struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Designation:
	//
	//	*QualityOfService_Tier_
	//	*QualityOfService_Spec
	Designation isQualityOfService_Designation `protobuf_oneof:"designation"`
}

func (x *QualityOfService) Reset() {
	*x = QualityOfService{}
	if protoimpl.UnsafeEnabled {
		mi := &file_flyteidl_core_execution_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QualityOfService) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QualityOfService) ProtoMessage() {}

func (x *QualityOfService) ProtoReflect() protoreflect.Message {
	mi := &file_flyteidl_core_execution_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QualityOfService.ProtoReflect.Descriptor instead.
func (*QualityOfService) Descriptor() ([]byte, []int) {
	return file_flyteidl_core_execution_proto_rawDescGZIP(), []int{6}
}

func (m *QualityOfService) GetDesignation() isQualityOfService_Designation {
	if m != nil {
		return m.Designation
	}
	return nil
}

func (x *QualityOfService) GetTier() QualityOfService_Tier {
	if x, ok := x.GetDesignation().(*QualityOfService_Tier_); ok {
		return x.Tier
	}
	return QualityOfService_UNDEFINED
}

func (x *QualityOfService) GetSpec() *QualityOfServiceSpec {
	if x, ok := x.GetDesignation().(*QualityOfService_Spec); ok {
		return x.Spec
	}
	return nil
}

type isQualityOfService_Designation interface {
	isQualityOfService_Designation()
}

type QualityOfService_Tier_ struct {
	Tier QualityOfService_Tier `protobuf:"varint,1,opt,name=tier,proto3,enum=flyteidl.core.QualityOfService_Tier,oneof"`
}

type QualityOfService_Spec struct {
	Spec *QualityOfServiceSpec `protobuf:"bytes,2,opt,name=spec,proto3,oneof"`
}

func (*QualityOfService_Tier_) isQualityOfService_Designation() {}

func (*QualityOfService_Spec) isQualityOfService_Designation() {}

var File_flyteidl_core_execution_proto protoreflect.FileDescriptor

var file_flyteidl_core_execution_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0d, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x1a, 0x1e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xa7, 0x01, 0x0a, 0x11, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x91, 0x01, 0x0a, 0x05, 0x50, 0x68, 0x61, 0x73, 0x65, 0x12,
	0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0a,
	0x0a, 0x06, 0x51, 0x55, 0x45, 0x55, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55,
	0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x55, 0x43, 0x43, 0x45,
	0x45, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x03, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x55, 0x43, 0x43, 0x45,
	0x45, 0x44, 0x45, 0x44, 0x10, 0x04, 0x12, 0x0b, 0x0a, 0x07, 0x46, 0x41, 0x49, 0x4c, 0x49, 0x4e,
	0x47, 0x10, 0x05, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x06, 0x12,
	0x0b, 0x0a, 0x07, 0x41, 0x42, 0x4f, 0x52, 0x54, 0x45, 0x44, 0x10, 0x07, 0x12, 0x0d, 0x0a, 0x09,
	0x54, 0x49, 0x4d, 0x45, 0x44, 0x5f, 0x4f, 0x55, 0x54, 0x10, 0x08, 0x12, 0x0c, 0x0a, 0x08, 0x41,
	0x42, 0x4f, 0x52, 0x54, 0x49, 0x4e, 0x47, 0x10, 0x09, 0x22, 0xb6, 0x01, 0x0a, 0x0d, 0x4e, 0x6f,
	0x64, 0x65, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xa4, 0x01, 0x0a, 0x05,
	0x50, 0x68, 0x61, 0x73, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e,
	0x45, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x51, 0x55, 0x45, 0x55, 0x45, 0x44, 0x10, 0x01,
	0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x0d, 0x0a,
	0x09, 0x53, 0x55, 0x43, 0x43, 0x45, 0x45, 0x44, 0x45, 0x44, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07,
	0x46, 0x41, 0x49, 0x4c, 0x49, 0x4e, 0x47, 0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49,
	0x4c, 0x45, 0x44, 0x10, 0x05, 0x12, 0x0b, 0x0a, 0x07, 0x41, 0x42, 0x4f, 0x52, 0x54, 0x45, 0x44,
	0x10, 0x06, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x4b, 0x49, 0x50, 0x50, 0x45, 0x44, 0x10, 0x07, 0x12,
	0x0d, 0x0a, 0x09, 0x54, 0x49, 0x4d, 0x45, 0x44, 0x5f, 0x4f, 0x55, 0x54, 0x10, 0x08, 0x12, 0x13,
	0x0a, 0x0f, 0x44, 0x59, 0x4e, 0x41, 0x4d, 0x49, 0x43, 0x5f, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e,
	0x47, 0x10, 0x09, 0x12, 0x0d, 0x0a, 0x09, 0x52, 0x45, 0x43, 0x4f, 0x56, 0x45, 0x52, 0x45, 0x44,
	0x10, 0x0a, 0x22, 0x96, 0x01, 0x0a, 0x0d, 0x54, 0x61, 0x73, 0x6b, 0x45, 0x78, 0x65, 0x63, 0x75,
	0x74, 0x69, 0x6f, 0x6e, 0x22, 0x84, 0x01, 0x0a, 0x05, 0x50, 0x68, 0x61, 0x73, 0x65, 0x12, 0x0d,
	0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a,
	0x06, 0x51, 0x55, 0x45, 0x55, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55, 0x4e,
	0x4e, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x55, 0x43, 0x43, 0x45, 0x45,
	0x44, 0x45, 0x44, 0x10, 0x03, 0x12, 0x0b, 0x0a, 0x07, 0x41, 0x42, 0x4f, 0x52, 0x54, 0x45, 0x44,
	0x10, 0x04, 0x12, 0x0a, 0x0a, 0x06, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x05, 0x12, 0x10,
	0x0a, 0x0c, 0x49, 0x4e, 0x49, 0x54, 0x49, 0x41, 0x4c, 0x49, 0x5a, 0x49, 0x4e, 0x47, 0x10, 0x06,
	0x12, 0x19, 0x0a, 0x15, 0x57, 0x41, 0x49, 0x54, 0x49, 0x4e, 0x47, 0x5f, 0x46, 0x4f, 0x52, 0x5f,
	0x52, 0x45, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x53, 0x10, 0x07, 0x22, 0x9a, 0x02, 0x0a, 0x0e,
	0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x6f,
	0x64, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1b, 0x0a, 0x09,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x75, 0x72, 0x69, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x55, 0x72, 0x69, 0x12, 0x3b, 0x0a, 0x04, 0x6b, 0x69, 0x6e,
	0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69,
	0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f,
	0x6e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x4b, 0x69, 0x6e, 0x64,
	0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x16, 0x0a, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x22, 0x2e, 0x0a, 0x09, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x4b, 0x69, 0x6e, 0x64, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e,
	0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06,
	0x53, 0x59, 0x53, 0x54, 0x45, 0x4d, 0x10, 0x02, 0x22, 0xb2, 0x02, 0x0a, 0x07, 0x54, 0x61, 0x73,
	0x6b, 0x4c, 0x6f, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x75, 0x72, 0x69, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x4b, 0x0a, 0x0e, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x5f, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x24, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f,
	0x72, 0x65, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x4c, 0x6f, 0x67, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x52, 0x0d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x2b, 0x0a, 0x03, 0x74, 0x74, 0x6c, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x03, 0x74, 0x74, 0x6c, 0x12, 0x2a, 0x0a, 0x10, 0x53, 0x68, 0x6f, 0x77, 0x57, 0x68, 0x69, 0x6c,
	0x65, 0x50, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10,
	0x53, 0x68, 0x6f, 0x77, 0x57, 0x68, 0x69, 0x6c, 0x65, 0x50, 0x65, 0x6e, 0x64, 0x69, 0x6e, 0x67,
	0x12, 0x2a, 0x0a, 0x10, 0x48, 0x69, 0x64, 0x65, 0x4f, 0x6e, 0x63, 0x65, 0x46, 0x69, 0x6e, 0x69,
	0x73, 0x68, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10, 0x48, 0x69, 0x64, 0x65,
	0x4f, 0x6e, 0x63, 0x65, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x65, 0x64, 0x22, 0x2f, 0x0a, 0x0d,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x0b, 0x0a,
	0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x43, 0x53,
	0x56, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x4a, 0x53, 0x4f, 0x4e, 0x10, 0x02, 0x22, 0x5a, 0x0a,
	0x14, 0x51, 0x75, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x4f, 0x66, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x53, 0x70, 0x65, 0x63, 0x12, 0x42, 0x0a, 0x0f, 0x71, 0x75, 0x65, 0x75, 0x65, 0x69, 0x6e,
	0x67, 0x5f, 0x62, 0x75, 0x64, 0x67, 0x65, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0e, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x69, 0x6e, 0x67, 0x42, 0x75, 0x64, 0x67, 0x65, 0x74, 0x22, 0xce, 0x01, 0x0a, 0x10, 0x51, 0x75,
	0x61, 0x6c, 0x69, 0x74, 0x79, 0x4f, 0x66, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3a,
	0x0a, 0x04, 0x74, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x66,
	0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x51, 0x75, 0x61,
	0x6c, 0x69, 0x74, 0x79, 0x4f, 0x66, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x54, 0x69,
	0x65, 0x72, 0x48, 0x00, 0x52, 0x04, 0x74, 0x69, 0x65, 0x72, 0x12, 0x39, 0x0a, 0x04, 0x73, 0x70,
	0x65, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65,
	0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x51, 0x75, 0x61, 0x6c, 0x69, 0x74, 0x79,
	0x4f, 0x66, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x70, 0x65, 0x63, 0x48, 0x00, 0x52,
	0x04, 0x73, 0x70, 0x65, 0x63, 0x22, 0x34, 0x0a, 0x04, 0x54, 0x69, 0x65, 0x72, 0x12, 0x0d, 0x0a,
	0x09, 0x55, 0x4e, 0x44, 0x45, 0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04,
	0x48, 0x49, 0x47, 0x48, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x4d, 0x45, 0x44, 0x49, 0x55, 0x4d,
	0x10, 0x02, 0x12, 0x07, 0x0a, 0x03, 0x4c, 0x4f, 0x57, 0x10, 0x03, 0x42, 0x0d, 0x0a, 0x0b, 0x64,
	0x65, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0xb4, 0x01, 0x0a, 0x11, 0x63,
	0x6f, 0x6d, 0x2e, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e, 0x63, 0x6f, 0x72, 0x65,
	0x42, 0x0e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66,
	0x6c, 0x79, 0x74, 0x65, 0x6f, 0x72, 0x67, 0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x2f, 0x66, 0x6c,
	0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x62, 0x2d, 0x67, 0x6f,
	0x2f, 0x66, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0xa2, 0x02,
	0x03, 0x46, 0x43, 0x58, 0xaa, 0x02, 0x0d, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x2e,
	0x43, 0x6f, 0x72, 0x65, 0xca, 0x02, 0x0d, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c,
	0x43, 0x6f, 0x72, 0x65, 0xe2, 0x02, 0x19, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x5c,
	0x43, 0x6f, 0x72, 0x65, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x0e, 0x46, 0x6c, 0x79, 0x74, 0x65, 0x69, 0x64, 0x6c, 0x3a, 0x3a, 0x43, 0x6f, 0x72,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_flyteidl_core_execution_proto_rawDescOnce sync.Once
	file_flyteidl_core_execution_proto_rawDescData = file_flyteidl_core_execution_proto_rawDesc
)

func file_flyteidl_core_execution_proto_rawDescGZIP() []byte {
	file_flyteidl_core_execution_proto_rawDescOnce.Do(func() {
		file_flyteidl_core_execution_proto_rawDescData = protoimpl.X.CompressGZIP(file_flyteidl_core_execution_proto_rawDescData)
	})
	return file_flyteidl_core_execution_proto_rawDescData
}

var file_flyteidl_core_execution_proto_enumTypes = make([]protoimpl.EnumInfo, 6)
var file_flyteidl_core_execution_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_flyteidl_core_execution_proto_goTypes = []interface{}{
	(WorkflowExecution_Phase)(0),  // 0: flyteidl.core.WorkflowExecution.Phase
	(NodeExecution_Phase)(0),      // 1: flyteidl.core.NodeExecution.Phase
	(TaskExecution_Phase)(0),      // 2: flyteidl.core.TaskExecution.Phase
	(ExecutionError_ErrorKind)(0), // 3: flyteidl.core.ExecutionError.ErrorKind
	(TaskLog_MessageFormat)(0),    // 4: flyteidl.core.TaskLog.MessageFormat
	(QualityOfService_Tier)(0),    // 5: flyteidl.core.QualityOfService.Tier
	(*WorkflowExecution)(nil),     // 6: flyteidl.core.WorkflowExecution
	(*NodeExecution)(nil),         // 7: flyteidl.core.NodeExecution
	(*TaskExecution)(nil),         // 8: flyteidl.core.TaskExecution
	(*ExecutionError)(nil),        // 9: flyteidl.core.ExecutionError
	(*TaskLog)(nil),               // 10: flyteidl.core.TaskLog
	(*QualityOfServiceSpec)(nil),  // 11: flyteidl.core.QualityOfServiceSpec
	(*QualityOfService)(nil),      // 12: flyteidl.core.QualityOfService
	(*timestamppb.Timestamp)(nil), // 13: google.protobuf.Timestamp
	(*durationpb.Duration)(nil),   // 14: google.protobuf.Duration
}
var file_flyteidl_core_execution_proto_depIdxs = []int32{
	3,  // 0: flyteidl.core.ExecutionError.kind:type_name -> flyteidl.core.ExecutionError.ErrorKind
	13, // 1: flyteidl.core.ExecutionError.timestamp:type_name -> google.protobuf.Timestamp
	4,  // 2: flyteidl.core.TaskLog.message_format:type_name -> flyteidl.core.TaskLog.MessageFormat
	14, // 3: flyteidl.core.TaskLog.ttl:type_name -> google.protobuf.Duration
	14, // 4: flyteidl.core.QualityOfServiceSpec.queueing_budget:type_name -> google.protobuf.Duration
	5,  // 5: flyteidl.core.QualityOfService.tier:type_name -> flyteidl.core.QualityOfService.Tier
	11, // 6: flyteidl.core.QualityOfService.spec:type_name -> flyteidl.core.QualityOfServiceSpec
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_flyteidl_core_execution_proto_init() }
func file_flyteidl_core_execution_proto_init() {
	if File_flyteidl_core_execution_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_flyteidl_core_execution_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkflowExecution); i {
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
		file_flyteidl_core_execution_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeExecution); i {
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
		file_flyteidl_core_execution_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskExecution); i {
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
		file_flyteidl_core_execution_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecutionError); i {
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
		file_flyteidl_core_execution_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskLog); i {
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
		file_flyteidl_core_execution_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QualityOfServiceSpec); i {
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
		file_flyteidl_core_execution_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QualityOfService); i {
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
	file_flyteidl_core_execution_proto_msgTypes[6].OneofWrappers = []interface{}{
		(*QualityOfService_Tier_)(nil),
		(*QualityOfService_Spec)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_flyteidl_core_execution_proto_rawDesc,
			NumEnums:      6,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_flyteidl_core_execution_proto_goTypes,
		DependencyIndexes: file_flyteidl_core_execution_proto_depIdxs,
		EnumInfos:         file_flyteidl_core_execution_proto_enumTypes,
		MessageInfos:      file_flyteidl_core_execution_proto_msgTypes,
	}.Build()
	File_flyteidl_core_execution_proto = out.File
	file_flyteidl_core_execution_proto_rawDesc = nil
	file_flyteidl_core_execution_proto_goTypes = nil
	file_flyteidl_core_execution_proto_depIdxs = nil
}
