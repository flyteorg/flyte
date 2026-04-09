package models

import (
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// ActionEvent represents a single phase transition event for an action attempt.
// Composite PK: (project, domain, run_name, name, attempt, phase, version).
type ActionEvent struct {
	// Composite primary key
	Project string `db:"project"`
	Domain  string `db:"domain"`
	RunName string `db:"run_name"`
	Name    string `db:"name"`
	Attempt uint32 `db:"attempt"`
	Phase   int32  `db:"phase"` // common.ActionPhase
	Version uint32 `db:"version"`

	// Serialized workflow.ActionEvent proto
	Info []byte `db:"info"`

	// Error kind denormalized from Info for faster queries
	ErrorKind *string `db:"error_kind"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ToActionEvent unmarshals the Info bytes into a workflow.ActionEvent proto.
func (m *ActionEvent) ToActionEvent() (*workflow.ActionEvent, error) {
	event := &workflow.ActionEvent{}
	if err := proto.Unmarshal(m.Info, event); err != nil {
		return nil, err
	}
	return event, nil
}

// NewActionEventModel builds an ActionEvent model from a proto ActionEvent.
func NewActionEventModel(event *workflow.ActionEvent) (*ActionEvent, error) {
	m := &ActionEvent{
		Project: event.GetId().GetRun().GetProject(),
		Domain:  event.GetId().GetRun().GetDomain(),
		RunName: event.GetId().GetRun().GetName(),
		Name:    event.GetId().GetName(),
		Attempt: event.GetAttempt(),
		Phase:   int32(event.GetPhase()),
		Version: event.GetVersion(),
	}

	if event.GetErrorInfo() != nil {
		errorKind := event.GetErrorInfo().GetKind().String()
		m.ErrorKind = &errorKind
	}

	info, err := proto.Marshal(event)
	if err != nil {
		return nil, err
	}
	m.Info = info

	return m, nil
}
