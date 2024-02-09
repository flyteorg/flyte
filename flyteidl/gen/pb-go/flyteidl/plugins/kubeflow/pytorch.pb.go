// Code generated by protoc-gen-go. DO NOT EDIT.
// source: flyteidl/plugins/kubeflow/pytorch.proto

package plugins

import (
	fmt "fmt"
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Custom proto for torch elastic config for distributed training using
// https://github.com/kubeflow/training-operator/blob/master/pkg/apis/kubeflow.org/v1/pytorch_types.go
type ElasticConfig struct {
	RdzvBackend          string   `protobuf:"bytes,1,opt,name=rdzv_backend,json=rdzvBackend,proto3" json:"rdzv_backend,omitempty"`
	MinReplicas          int32    `protobuf:"varint,2,opt,name=min_replicas,json=minReplicas,proto3" json:"min_replicas,omitempty"`
	MaxReplicas          int32    `protobuf:"varint,3,opt,name=max_replicas,json=maxReplicas,proto3" json:"max_replicas,omitempty"`
	NprocPerNode         int32    `protobuf:"varint,4,opt,name=nproc_per_node,json=nprocPerNode,proto3" json:"nproc_per_node,omitempty"`
	MaxRestarts          int32    `protobuf:"varint,5,opt,name=max_restarts,json=maxRestarts,proto3" json:"max_restarts,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ElasticConfig) Reset()         { *m = ElasticConfig{} }
func (m *ElasticConfig) String() string { return proto.CompactTextString(m) }
func (*ElasticConfig) ProtoMessage()    {}
func (*ElasticConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_37e97bee6e09d707, []int{0}
}

func (m *ElasticConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ElasticConfig.Unmarshal(m, b)
}
func (m *ElasticConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ElasticConfig.Marshal(b, m, deterministic)
}
func (m *ElasticConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ElasticConfig.Merge(m, src)
}
func (m *ElasticConfig) XXX_Size() int {
	return xxx_messageInfo_ElasticConfig.Size(m)
}
func (m *ElasticConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ElasticConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ElasticConfig proto.InternalMessageInfo

func (m *ElasticConfig) GetRdzvBackend() string {
	if m != nil {
		return m.RdzvBackend
	}
	return ""
}

func (m *ElasticConfig) GetMinReplicas() int32 {
	if m != nil {
		return m.MinReplicas
	}
	return 0
}

func (m *ElasticConfig) GetMaxReplicas() int32 {
	if m != nil {
		return m.MaxReplicas
	}
	return 0
}

func (m *ElasticConfig) GetNprocPerNode() int32 {
	if m != nil {
		return m.NprocPerNode
	}
	return 0
}

func (m *ElasticConfig) GetMaxRestarts() int32 {
	if m != nil {
		return m.MaxRestarts
	}
	return 0
}

// Proto for plugin that enables distributed training using https://github.com/kubeflow/pytorch-operator
type DistributedPyTorchTrainingTask struct {
	// Worker replicas spec
	WorkerReplicas *DistributedPyTorchTrainingReplicaSpec `protobuf:"bytes,1,opt,name=worker_replicas,json=workerReplicas,proto3" json:"worker_replicas,omitempty"`
	// Master replicas spec, master replicas can only have 1 replica
	MasterReplicas *DistributedPyTorchTrainingReplicaSpec `protobuf:"bytes,2,opt,name=master_replicas,json=masterReplicas,proto3" json:"master_replicas,omitempty"`
	// RunPolicy encapsulates various runtime policies of the distributed training
	// job, for example how to clean up resources and how long the job can stay
	// active.
	RunPolicy *RunPolicy `protobuf:"bytes,3,opt,name=run_policy,json=runPolicy,proto3" json:"run_policy,omitempty"`
	// config for an elastic pytorch job
	ElasticConfig        *ElasticConfig `protobuf:"bytes,4,opt,name=elastic_config,json=elasticConfig,proto3" json:"elastic_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *DistributedPyTorchTrainingTask) Reset()         { *m = DistributedPyTorchTrainingTask{} }
func (m *DistributedPyTorchTrainingTask) String() string { return proto.CompactTextString(m) }
func (*DistributedPyTorchTrainingTask) ProtoMessage()    {}
func (*DistributedPyTorchTrainingTask) Descriptor() ([]byte, []int) {
	return fileDescriptor_37e97bee6e09d707, []int{1}
}

func (m *DistributedPyTorchTrainingTask) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DistributedPyTorchTrainingTask.Unmarshal(m, b)
}
func (m *DistributedPyTorchTrainingTask) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DistributedPyTorchTrainingTask.Marshal(b, m, deterministic)
}
func (m *DistributedPyTorchTrainingTask) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DistributedPyTorchTrainingTask.Merge(m, src)
}
func (m *DistributedPyTorchTrainingTask) XXX_Size() int {
	return xxx_messageInfo_DistributedPyTorchTrainingTask.Size(m)
}
func (m *DistributedPyTorchTrainingTask) XXX_DiscardUnknown() {
	xxx_messageInfo_DistributedPyTorchTrainingTask.DiscardUnknown(m)
}

var xxx_messageInfo_DistributedPyTorchTrainingTask proto.InternalMessageInfo

func (m *DistributedPyTorchTrainingTask) GetWorkerReplicas() *DistributedPyTorchTrainingReplicaSpec {
	if m != nil {
		return m.WorkerReplicas
	}
	return nil
}

func (m *DistributedPyTorchTrainingTask) GetMasterReplicas() *DistributedPyTorchTrainingReplicaSpec {
	if m != nil {
		return m.MasterReplicas
	}
	return nil
}

func (m *DistributedPyTorchTrainingTask) GetRunPolicy() *RunPolicy {
	if m != nil {
		return m.RunPolicy
	}
	return nil
}

func (m *DistributedPyTorchTrainingTask) GetElasticConfig() *ElasticConfig {
	if m != nil {
		return m.ElasticConfig
	}
	return nil
}

type DistributedPyTorchTrainingReplicaSpec struct {
	// Number of replicas
	Replicas int32 `protobuf:"varint,1,opt,name=replicas,proto3" json:"replicas,omitempty"`
	// Image used for the replica group
	Image string `protobuf:"bytes,2,opt,name=image,proto3" json:"image,omitempty"`
	// Resources required for the replica group
	Resources *core.Resources `protobuf:"bytes,3,opt,name=resources,proto3" json:"resources,omitempty"`
	// RestartPolicy determines whether pods will be restarted when they exit
	RestartPolicy RestartPolicy `protobuf:"varint,4,opt,name=restart_policy,json=restartPolicy,proto3,enum=flyteidl.plugins.kubeflow.RestartPolicy" json:"restart_policy,omitempty"`
	// Node selectors for the replica group
	NodeSelectors        map[string]string `protobuf:"bytes,5,rep,name=node_selectors,json=nodeSelectors,proto3" json:"node_selectors,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *DistributedPyTorchTrainingReplicaSpec) Reset()         { *m = DistributedPyTorchTrainingReplicaSpec{} }
func (m *DistributedPyTorchTrainingReplicaSpec) String() string { return proto.CompactTextString(m) }
func (*DistributedPyTorchTrainingReplicaSpec) ProtoMessage()    {}
func (*DistributedPyTorchTrainingReplicaSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_37e97bee6e09d707, []int{2}
}

func (m *DistributedPyTorchTrainingReplicaSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DistributedPyTorchTrainingReplicaSpec.Unmarshal(m, b)
}
func (m *DistributedPyTorchTrainingReplicaSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DistributedPyTorchTrainingReplicaSpec.Marshal(b, m, deterministic)
}
func (m *DistributedPyTorchTrainingReplicaSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DistributedPyTorchTrainingReplicaSpec.Merge(m, src)
}
func (m *DistributedPyTorchTrainingReplicaSpec) XXX_Size() int {
	return xxx_messageInfo_DistributedPyTorchTrainingReplicaSpec.Size(m)
}
func (m *DistributedPyTorchTrainingReplicaSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_DistributedPyTorchTrainingReplicaSpec.DiscardUnknown(m)
}

var xxx_messageInfo_DistributedPyTorchTrainingReplicaSpec proto.InternalMessageInfo

func (m *DistributedPyTorchTrainingReplicaSpec) GetReplicas() int32 {
	if m != nil {
		return m.Replicas
	}
	return 0
}

func (m *DistributedPyTorchTrainingReplicaSpec) GetImage() string {
	if m != nil {
		return m.Image
	}
	return ""
}

func (m *DistributedPyTorchTrainingReplicaSpec) GetResources() *core.Resources {
	if m != nil {
		return m.Resources
	}
	return nil
}

func (m *DistributedPyTorchTrainingReplicaSpec) GetRestartPolicy() RestartPolicy {
	if m != nil {
		return m.RestartPolicy
	}
	return RestartPolicy_RESTART_POLICY_NEVER
}

func (m *DistributedPyTorchTrainingReplicaSpec) GetNodeSelectors() map[string]string {
	if m != nil {
		return m.NodeSelectors
	}
	return nil
}

func init() {
	proto.RegisterType((*ElasticConfig)(nil), "flyteidl.plugins.kubeflow.ElasticConfig")
	proto.RegisterType((*DistributedPyTorchTrainingTask)(nil), "flyteidl.plugins.kubeflow.DistributedPyTorchTrainingTask")
	proto.RegisterType((*DistributedPyTorchTrainingReplicaSpec)(nil), "flyteidl.plugins.kubeflow.DistributedPyTorchTrainingReplicaSpec")
	proto.RegisterMapType((map[string]string)(nil), "flyteidl.plugins.kubeflow.DistributedPyTorchTrainingReplicaSpec.NodeSelectorsEntry")
}

func init() {
	proto.RegisterFile("flyteidl/plugins/kubeflow/pytorch.proto", fileDescriptor_37e97bee6e09d707)
}

var fileDescriptor_37e97bee6e09d707 = []byte{
	// 534 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0xdd, 0x6e, 0xd3, 0x30,
	0x18, 0x55, 0x5b, 0x8a, 0xa8, 0xb3, 0x76, 0x28, 0xe2, 0xa2, 0xf4, 0x02, 0x8d, 0x6a, 0x40, 0x6f,
	0x48, 0xa4, 0x22, 0x21, 0x84, 0x84, 0x98, 0x36, 0x76, 0x0b, 0x95, 0xdb, 0x2b, 0x6e, 0x22, 0xc7,
	0x71, 0x33, 0x2b, 0x89, 0x1d, 0x7d, 0x76, 0xb6, 0x65, 0xcf, 0xc0, 0x43, 0xf0, 0x2a, 0xbc, 0x19,
	0xb2, 0xf3, 0xd3, 0x0c, 0xd4, 0x0a, 0x89, 0xdd, 0xf9, 0xfb, 0x72, 0x72, 0x3e, 0x9f, 0xe3, 0x63,
	0xa3, 0x37, 0xdb, 0xb4, 0xd4, 0x8c, 0x47, 0xa9, 0x9f, 0xa7, 0x45, 0xcc, 0x85, 0xf2, 0x93, 0x22,
	0x64, 0xdb, 0x54, 0xde, 0xf8, 0x79, 0xa9, 0x25, 0xd0, 0x2b, 0x2f, 0x07, 0xa9, 0xa5, 0xfb, 0xbc,
	0x01, 0x7a, 0x35, 0xd0, 0x6b, 0x80, 0xb3, 0xf6, 0x93, 0x4f, 0x25, 0x30, 0x5f, 0x13, 0x95, 0xa8,
	0xea, 0xaf, 0xd9, 0xeb, 0xfd, 0xf4, 0x54, 0x66, 0x99, 0x14, 0x15, 0x6e, 0xfe, 0xab, 0x87, 0xc6,
	0x97, 0x29, 0x51, 0x9a, 0xd3, 0x0b, 0x29, 0xb6, 0x3c, 0x76, 0x5f, 0xa2, 0x23, 0x88, 0xee, 0xae,
	0x83, 0x90, 0xd0, 0x84, 0x89, 0x68, 0xda, 0x3b, 0xe9, 0x2d, 0x46, 0xd8, 0x31, 0xbd, 0xf3, 0xaa,
	0x65, 0x20, 0x19, 0x17, 0x01, 0xb0, 0x3c, 0xe5, 0x94, 0xa8, 0x69, 0xff, 0xa4, 0xb7, 0x18, 0x62,
	0x27, 0xe3, 0x02, 0xd7, 0x2d, 0x0b, 0x21, 0xb7, 0x3b, 0xc8, 0xa0, 0x86, 0x90, 0xdb, 0x16, 0x72,
	0x8a, 0x26, 0x22, 0x07, 0x49, 0x83, 0x9c, 0x41, 0x20, 0x64, 0xc4, 0xa6, 0x8f, 0x2c, 0xe8, 0xc8,
	0x76, 0x57, 0x0c, 0xbe, 0xca, 0x88, 0xed, 0x88, 0x94, 0x26, 0xa0, 0xd5, 0x74, 0xd8, 0x21, 0xaa,
	0x5a, 0xf3, 0x1f, 0x03, 0xf4, 0xe2, 0x0b, 0x57, 0x1a, 0x78, 0x58, 0x68, 0x16, 0xad, 0xca, 0x8d,
	0xb1, 0x6f, 0x03, 0x84, 0x0b, 0x2e, 0xe2, 0x0d, 0x51, 0x89, 0xcb, 0xd1, 0xf1, 0x8d, 0x84, 0x84,
	0xc1, 0x6e, 0x47, 0x46, 0x97, 0xb3, 0x3c, 0xf3, 0xf6, 0xda, 0xeb, 0xed, 0xe7, 0xac, 0x35, 0xac,
	0x73, 0x46, 0xf1, 0xa4, 0x22, 0x6e, 0x65, 0x71, 0x74, 0x9c, 0x11, 0xa5, 0xbb, 0xa3, 0xfa, 0x0f,
	0x35, 0xaa, 0x22, 0x6e, 0x47, 0x5d, 0x20, 0x04, 0x85, 0x08, 0x72, 0x99, 0x72, 0x5a, 0x5a, 0x8b,
	0x9d, 0xe5, 0xe9, 0x81, 0x29, 0xb8, 0x10, 0x2b, 0x8b, 0xc5, 0x23, 0x68, 0x96, 0xee, 0x37, 0x34,
	0x61, 0x55, 0x00, 0x02, 0x6a, 0x13, 0x60, 0x8f, 0xc1, 0x59, 0x2e, 0x0e, 0x10, 0xdd, 0x4b, 0x0c,
	0x1e, 0xb3, 0x6e, 0x39, 0xff, 0x39, 0x40, 0xaf, 0xfe, 0x49, 0x8f, 0x3b, 0x43, 0x4f, 0xee, 0x1d,
	0xc7, 0x10, 0xb7, 0xb5, 0xfb, 0x0c, 0x0d, 0x79, 0x46, 0x62, 0x66, 0xcd, 0x1b, 0xe1, 0xaa, 0x70,
	0xdf, 0xa3, 0x11, 0x30, 0x25, 0x0b, 0xa0, 0x4c, 0xd5, 0x82, 0xa7, 0xbb, 0x7d, 0x9a, 0x5b, 0xe0,
	0xe1, 0xe6, 0x3b, 0xde, 0x41, 0x8d, 0xc8, 0x3a, 0x41, 0x8d, 0x5b, 0x46, 0xe4, 0xe4, 0xa0, 0xc8,
	0x3a, 0x5f, 0xb5, 0x63, 0x63, 0xe8, 0x96, 0xee, 0x1d, 0x9a, 0x98, 0xc8, 0x06, 0x8a, 0xa5, 0x8c,
	0x6a, 0x09, 0x26, 0x98, 0x83, 0x85, 0xb3, 0x5c, 0xff, 0xef, 0x21, 0x7b, 0x26, 0xf5, 0xeb, 0x86,
	0xf5, 0x52, 0x68, 0x28, 0xf1, 0x58, 0x74, 0x7b, 0xb3, 0x33, 0xe4, 0xfe, 0x0d, 0x72, 0x9f, 0xa2,
	0x41, 0xc2, 0xca, 0xfa, 0xba, 0x9a, 0xa5, 0xb1, 0xf0, 0x9a, 0xa4, 0x45, 0x6b, 0xa1, 0x2d, 0x3e,
	0xf6, 0x3f, 0xf4, 0xce, 0x3f, 0x7f, 0xff, 0x14, 0x73, 0x7d, 0x55, 0x84, 0x1e, 0x95, 0x99, 0x6f,
	0x77, 0x2c, 0x21, 0xae, 0x16, 0x7e, 0xfb, 0x72, 0xc4, 0x4c, 0xf8, 0x79, 0xf8, 0x36, 0x96, 0xfe,
	0x9f, 0x8f, 0x49, 0xf8, 0xd8, 0xbe, 0x1e, 0xef, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0xf1, 0x94,
	0xc7, 0x0f, 0xc6, 0x04, 0x00, 0x00,
}
