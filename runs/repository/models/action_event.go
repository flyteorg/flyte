package models

import (
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// ActionEvent represents a single phase transition event for an action attempt.
// Composite PK: (org, project, domain, run_name, name, attempt, phase, version).
type ActionEvent struct {
	// Composite primary key
	Org     string `gorm:"not null;primaryKey" db:"org"`
	Project string `gorm:"not null;primaryKey" db:"project"`
	Domain  string `gorm:"not null;primaryKey" db:"domain"`
	RunName string `gorm:"not null;primaryKey" db:"run_name"`
	Name    string `gorm:"not null;primaryKey" db:"name"`
	Attempt uint32 `gorm:"not null;primaryKey" db:"attempt"`
	Phase   int32  `gorm:"not null;primaryKey" db:"phase"` // common.ActionPhase
	Version uint32 `gorm:"not null;primaryKey" db:"version"`

	// Serialized workflow.ActionEvent proto
	Info []byte `gorm:"type:bytea" db:"info"`

	// Error kind denormalized from Info for faster queries
	ErrorKind *string `gorm:"index:idx_action_events_error_kind" db:"error_kind"`

	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"updated_at"`
}

func (m *ActionEvent) TableName() string { return "action_events" }

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
		Org:     event.GetId().GetRun().GetOrg(),
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
