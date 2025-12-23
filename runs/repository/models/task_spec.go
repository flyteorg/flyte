package models

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/flytestdlib/pbhash"
	flyteWorkflow "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

var (
	Marshaller = proto.MarshalOptions{}
)

// TaskSpec is the model for ALL action specs, including normal tasks, traces, conditional actions, etc.
type TaskSpec struct {
	// Base64 encoded digest used as a unique identifier for the task spec
	Digest string `gorm:"primaryKey" db:"digest"`

	// Base fields
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"updated_at"`

	// Marshaled task spec
	Spec []byte `gorm:"not null" db:"spec"`
}

// TableName specifies the table name
func (TaskSpec) TableName() string { return "task_specs" }

func NewTaskSpecModel(ctx context.Context, spec *flyteWorkflow.TaskSpec) (*TaskSpec, error) {
	digest, err := pbhash.ComputeHashString(ctx, spec)
	if err != nil {
		return nil, err
	}

	specBytes, err := Marshaller.Marshal(spec)
	if err != nil {
		return nil, err
	}

	return &TaskSpec{
		Digest: digest,
		Spec:   specBytes,
	}, nil
}

func NewTaskSpecModelFromTraceSpec(ctx context.Context, traceSpec *flyteWorkflow.TraceSpec) (*TaskSpec, error) {
	if traceSpec == nil {
		return nil, nil
	}

	digest, err := pbhash.ComputeHashString(ctx, traceSpec)
	if err != nil {
		return nil, err
	}

	specBytes, err := Marshaller.Marshal(traceSpec)
	if err != nil {
		return nil, err
	}

	return &TaskSpec{
		Digest: digest,
		Spec:   specBytes,
	}, nil
}

// ToTaskSpec uses the deprecated task spec unmarshalling
func ToTaskSpec(specModel *TaskSpec) (*flyteWorkflow.TaskSpec, error) {
	spec := &flyteWorkflow.TaskSpec{}
	if err := proto.Unmarshal(specModel.Spec, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

// ToTraceSpec uses the deprecated trace spec unmarshalling
func ToTraceSpec(specModel *TaskSpec) (*flyteWorkflow.TraceSpec, error) {
	spec := &flyteWorkflow.TraceSpec{}
	if err := proto.Unmarshal(specModel.Spec, spec); err != nil {
		return nil, err
	}
	return spec, nil
}
